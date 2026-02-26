// cmd/bunkr/install.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/caddy"
	"github.com/pankajbeniwal/bunkr/internal/docker"
	"github.com/pankajbeniwal/bunkr/internal/hardening"
	"github.com/pankajbeniwal/bunkr/internal/recipe"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/tailscale"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <recipe> [<recipe>...]",
	Short: "Harden server and install app(s)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRemote(); err != nil {
			return err
		}

		ctx := context.Background()

		// === PLAN PHASE (always local) ===

		// Fetch and validate all recipes
		type planned struct {
			recipe *recipe.Recipe
			values map[string]string
		}
		var plans []planned

		for _, name := range args {
			ui.Header(fmt.Sprintf("Fetching %s...", name))
			r, err := recipe.Fetch(name)
			if err != nil {
				return fmt.Errorf("failed to fetch recipe %s: %w", name, err)
			}

			ui.Header(fmt.Sprintf("Configuring %s...", name))
			values, err := recipe.PromptUser(r.Prompts)
			if err != nil {
				return err
			}

			// Expand auto_generate values in environment
			if r.Environment != nil {
				r.Environment = recipe.ExpandAutoGenerate(r.Environment)
			}

			plans = append(plans, planned{recipe: r, values: values})
		}

		// === EXECUTE PHASE (via executor) ===

		exec, err := newExecutor()
		if err != nil {
			return err
		}

		// Set system-wide apt lock timeout (fresh VPS often has apt running)
		exec.Run(ctx, `echo 'DPkg::Lock::Timeout "120";' > /etc/apt/apt.conf.d/99-bunkr-lock-wait`)

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		// Hardening
		if !s.Hardening.Applied {
			ui.Header("Hardening VPS...")
			if _, err := hardening.Run(ctx, exec, s, sshPortFlag); err != nil {
				return err
			}
			s.Hardening.AppliedAt = time.Now()
			s.Hardening.SSHPort = sshPortFlag

			ui.Result("Server hardened successfully!")
			ui.HardeningSummary(extractHost(onFlag), sshPortFlag)
		}

		// Docker
		ui.Info("Checking Docker...")
		if err := docker.EnsureInstalled(ctx, exec); err != nil {
			return err
		}
		ui.Success("Docker ready")

		// Check which infrastructure is needed
		hasPrivate := false
		hasPublic := false
		for _, p := range plans {
			if p.recipe.Private {
				hasPrivate = true
			} else {
				hasPublic = true
			}
		}

		// Tailscale (only if a private recipe is being installed)
		if hasPrivate {
			ui.Info("Checking Tailscale...")
			if err := tailscale.EnsureInstalled(ctx, exec); err != nil {
				return err
			}

			connected, _ := tailscale.IsConnected(ctx, exec)
			if !connected {
				hostname, err := tailscale.Connect(ctx, exec)
				if err != nil {
					return err
				}
				s.Tailscale.Hostname = hostname
			} else if s.Tailscale.Hostname == "" {
				hostname, err := tailscale.Hostname(ctx, exec)
				if err != nil {
					return err
				}
				s.Tailscale.Hostname = hostname
			}
			s.Tailscale.Installed = true
			s.Tailscale.Connected = true
			ui.Success("Tailscale ready")
		}

		// Caddy (only if a public recipe is being installed)
		if hasPublic {
			ui.Info("Checking Caddy...")
			if err := caddy.EnsureInstalled(ctx, exec); err != nil {
				return err
			}
			ui.Success("Caddy ready")
		}

		// Install each recipe
		for _, p := range plans {
			r := p.recipe
			ui.Header(fmt.Sprintf("Installing %s...", r.Name))

			hostPort := s.AllocatePort(r.Ports[0])

			// Generate files
			composeData, err := recipe.GenerateCompose(r, p.values, hostPort)
			if err != nil {
				return err
			}
			envData := recipe.GenerateEnv(p.values)

			// Create directory
			dir := fmt.Sprintf("/opt/bunkr/%s", r.Name)
			if _, err := exec.Run(ctx, fmt.Sprintf("mkdir -p %s", dir)); err != nil {
				return err
			}

			// Write files
			if err := exec.WriteFile(ctx, dir+"/docker-compose.yml", composeData, 0644); err != nil {
				return err
			}
			ui.Success("Compose file generated")

			if err := exec.WriteFile(ctx, dir+"/.env", envData, 0600); err != nil {
				return err
			}

			// Network: Tailscale for private, Caddy for public
			var domain string
			if r.Private {
				if err := tailscale.Serve(ctx, exec, hostPort); err != nil {
					return err
				}
				domain = s.Tailscale.Hostname
				ui.Success("Tailscale serve configured")
			} else {
				domain = p.values["DOMAIN"]
				if err := caddy.AddBlock(ctx, exec, r.Name, domain, hostPort); err != nil {
					return err
				}
				ui.Success("Caddy configured")
			}

			// Run init command (e.g. "openclaw setup") before starting
			if r.InitCommand != "" {
				ui.Info("Running init...")
				if err := docker.RunInit(ctx, exec, r.Name, r.InitCommand); err != nil {
					ui.Warn("Init command failed — continuing anyway")
				}
			}

			// Run post-init commands (e.g. patching config files)
			if len(r.PostInit) > 0 {
				ui.Info("Running post-init...")
				if err := docker.RunPostInit(ctx, exec, r.Name, r.PostInit); err != nil {
					return fmt.Errorf("post-init failed for %s: %w", r.Name, err)
				}
				ui.Success("Post-init complete")
			}

			// Pull and start containers
			ui.Info("Pulling image...")
			if err := docker.ComposeUp(ctx, exec, r.Name); err != nil {
				ui.Error("Failed to start containers")
				ui.Info(fmt.Sprintf("  Run: docker compose -f %s/docker-compose.yml logs", dir))
				return err
			}
			ui.Success("Containers started")

			// Health check
			if r.HealthCheck != nil {
				if err := docker.HealthCheck(ctx, exec, r.HealthCheck.URL, r.HealthCheck.Timeout, r.HealthCheck.Interval); err != nil {
					ui.Warn("Health check failed — app may still be starting")
				} else {
					ui.Success("Health check passed")
				}
			}

			// Update state
			s.Recipes[r.Name] = state.RecipeState{
				Version:       r.Version,
				Domain:        domain,
				Private:       r.Private,
				InstalledAt:   time.Now(),
				Port:          hostPort,
				ContainerPort: r.Ports[0],
			}
		}

		// Reload Caddy once (only if public recipes were installed)
		if hasPublic {
			if err := caddy.Reload(ctx, exec); err != nil {
				ui.Warn("Caddy reload failed — you may need to run 'caddy reload' manually")
			}
		}

		// Save state
		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		// Print results
		for _, p := range plans {
			rs := s.Recipes[p.recipe.Name]
			ui.Result(fmt.Sprintf("%s is running at https://%s", p.recipe.Name, rs.Domain))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
