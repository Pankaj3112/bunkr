# Bunkr Design Document

**Date:** 2026-02-25
**Status:** Approved

## Overview

Bunkr is a CLI tool written in Go that takes a fresh VPS and turns it into a hardened server running any self-hosted app in one command. Supports both local execution (on the VPS) and remote execution (from your laptop via SSH).

```bash
# Local (on the VPS)
bunkr install ghost

# Remote (from your laptop)
bunkr install ghost --on root@167.71.50.23
```

## Architecture

### Two-Phase Execution Model

Every command follows two phases:

1. **Plan phase (always local):** Parse args, fetch recipes, prompt user for config, generate compose files and .env in memory, validate inputs.
2. **Execute phase (via Executor):** Write files to server, run commands, verify results.

The Executor interface abstracts whether commands run locally or remotely. Phase 1 is always local regardless of mode. Phase 2 routes through the executor.

### Executor Interface

```go
type Executor interface {
    Run(ctx context.Context, cmd string) (string, error)
    WriteFile(ctx context.Context, path string, content []byte, mode os.FileMode) error
    ReadFile(ctx context.Context, path string) ([]byte, error)
}
```

Three methods. Everything else (mkdir, test -d, chmod) goes through `Run`.

- **LocalExecutor:** Wraps `os/exec` for Run, `os.WriteFile`/`os.ReadFile` for file ops.
- **RemoteExecutor:** Wraps `golang.org/x/crypto/ssh` for Run, uses SFTP or `cat >` over SSH for file ops. Uses existing SSH config and keys (same as `ssh user@host` working).

### `--on` Flag

Defined on the root command. If present, creates `RemoteExecutor`; otherwise `LocalExecutor`. Stored in cobra command context, passed to all subcommands.

```
bunkr install ghost                         # LocalExecutor
bunkr install ghost --on root@167.71.50.23  # RemoteExecutor
bunkr install ghost --on prod               # RemoteExecutor (SSH config alias)
```

## Project Structure

```
bunkr/
├── cmd/
│   └── bunkr/
│       ├── main.go
│       ├── root.go           # root command, --on flag
│       ├── init.go           # bunkr init
│       ├── install.go        # bunkr install
│       ├── uninstall.go      # bunkr uninstall
│       ├── update.go         # bunkr update
│       ├── list.go           # bunkr list
│       ├── status.go         # bunkr status
│       └── selfupdate.go     # bunkr self-update
├── internal/
│   ├── executor/
│   │   ├── executor.go       # Executor interface
│   │   ├── local.go          # LocalExecutor
│   │   └── remote.go         # RemoteExecutor (x/crypto/ssh)
│   ├── hardening/
│   │   ├── hardening.go      # orchestrator + step definitions
│   │   ├── ssh.go            # SSH hardening
│   │   ├── firewall.go       # UFW setup
│   │   ├── fail2ban.go       # fail2ban
│   │   ├── sysctl.go         # kernel params
│   │   ├── swap.go           # swap setup
│   │   ├── upgrades.go       # unattended upgrades
│   │   └── user.go           # sudo user creation
│   ├── recipe/
│   │   ├── recipe.go         # Recipe struct + YAML parsing
│   │   ├── fetch.go          # fetch from GitHub raw content
│   │   ├── compose.go        # generate docker-compose.yml
│   │   ├── env.go            # generate .env file
│   │   └── prompt.go         # interactive prompts from recipe spec
│   ├── caddy/
│   │   └── caddy.go          # Caddyfile block add/remove/reload
│   ├── docker/
│   │   └── docker.go         # install check, compose up/down/ps
│   ├── state/
│   │   └── state.go          # state.json read/write
│   └── ui/
│       └── ui.go             # colored output helpers
├── recipes/
│   ├── index.yaml            # list of all available recipes
│   ├── uptime-kuma.yaml
│   ├── ghost.yaml
│   └── plausible.yaml
├── scripts/
│   └── install.sh            # curl | sh installer
├── go.mod
├── go.sum
├── .goreleaser.yaml
└── .github/
    └── workflows/
        └── release.yaml
```

## Hardening System

### Step Pattern

Each hardening step is a struct with Check (already applied?) and Apply (do the work) functions:

```go
type Step struct {
    Name    string
    Label   string
    Check   func(ctx context.Context, exec executor.Executor) (bool, error)
    Apply   func(ctx context.Context, exec executor.Executor) error
}
```

The orchestrator runs steps in order. After each step: print status, update state.json.

### Steps (in order)

| # | Step | Check | Apply |
|---|------|-------|-------|
| 1 | Create sudo user | `id bunkr` | adduser, add to sudo, copy SSH keys from root |
| 2 | SSH hardening | sshd_config has `PermitRootLogin no` | Modify sshd_config: disable root login, disable password auth, change port. Restart sshd |
| 3 | Firewall (UFW) | `ufw status` shows active | Install ufw, allow SSH port + 80 + 443, enable |
| 4 | Fail2ban | `systemctl is-active fail2ban` | Install, create jail.local, enable+start |
| 5 | Unattended upgrades | Package installed check | Install + configure for security updates |
| 6 | Sysctl hardening | `/etc/sysctl.d/99-bunkr.conf` exists | Write params, `sysctl -p` |
| 7 | Swap | `swapon --show` has output | Create 1GB swapfile, configure fstab |

### SSH Hardening Safeguards

- Sudo user is created and verified BEFORE root login is disabled (`su - bunkr -c "whoami"`)
- SSH port is configurable: `bunkr init --ssh-port 2222` (default 2222)
- UFW allows the configured SSH port (reads from hardening config, not hardcoded)
- After SSH hardening, connection details are printed clearly:
  ```
  ⚠ SSH has been reconfigured:
    Port: 2222
    User: bunkr
    Connect with: ssh -p 2222 bunkr@your-server-ip
  ```
- In remote mode: RemoteExecutor reconnects using new port + user after SSH hardening. If reconnection fails, prints new connection details and exits cleanly.

## Recipe System

### Recipe Spec (YAML)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| name | string | yes | Recipe identifier (lowercase) |
| version | string | yes | App version |
| description | string | yes | One-line description |
| image | string | yes | Primary Docker image with tag |
| prompts | array | no | Interactive prompts for user config |
| prompts[].key | string | yes | Env var name |
| prompts[].label | string | yes | Display text |
| prompts[].required | bool | no | Whether input is required |
| prompts[].default | string | no | Default value |
| prompts[].secret | bool | no | Hide input |
| ports | array | yes | Container ports |
| volumes | array | no | Docker volumes (name:path) |
| services | array | no | Additional containers |
| environment | map | no | Env vars (supports `${KEY}` references, `auto_generate_32`, `auto_generate_64`) |
| health_check | object | no | Post-install verification |

### Fetching

- Default URL: `https://raw.githubusercontent.com/<owner>/<repo>/main/recipes/<name>.yaml`
- Same repo as bunkr code, recipes live under `/recipes`
- Override with `BUNKR_RECIPES_URL` env var
- `bunkr install` always fetches fresh; cache is fallback only if GitHub unreachable
- `bunkr list` fetches `index.yaml` which lists all available recipes

### Port Allocation

- Use container port as default host port
- Check state.json for already-allocated ports
- Auto-increment if port is taken
- Store both host port and container port in state
- Caddy proxies to `localhost:<host-port>`

### Compose Generation

- Port mapping: `127.0.0.1:<host-port>:<container-port>` (localhost only, Caddy handles external)
- Environment: expand `${KEY}` from prompt values, generate random strings for `auto_generate_*`
- Additional services from `services` field get their own entries

## Caddy Management

### Adding a block

```
# bunkr:<recipe-name>
<domain> {
    reverse_proxy localhost:<host-port>
}
# /bunkr:<recipe-name>
```

Read Caddyfile → append block → write back → reload.

### Removing a block

Read Caddyfile → find lines between `# bunkr:<name>` and `# /bunkr:<name>` → remove → write back → reload.

### Installing Caddy

Check `which caddy`. If missing, install via official Caddy apt repository.

## State Management

File: `/etc/bunkr/state.json`

```go
type State struct {
    Hardening HardeningState          `json:"hardening"`
    Recipes   map[string]RecipeState  `json:"recipes"`
}

type HardeningState struct {
    Applied   bool            `json:"applied"`
    Steps     map[string]bool `json:"steps"`
    AppliedAt time.Time       `json:"applied_at"`
    SSHPort   int             `json:"ssh_port"`
}

type RecipeState struct {
    Version       string    `json:"version"`
    Domain        string    `json:"domain"`
    InstalledAt   time.Time `json:"installed_at"`
    Port          int       `json:"port"`            // host port
    ContainerPort int       `json:"container_port"`  // container port
}
```

All state access goes through the Executor (ReadFile/WriteFile), so remote state works transparently.

## Command Flows

### `bunkr init [--ssh-port 2222] [--on user@host]`

1. Create executor (local or remote)
2. Load state (create if doesn't exist)
3. Run hardening orchestrator (skip already-done steps)
4. Save state after each step
5. Print summary with connection details

### `bunkr install <recipe> [<recipe>...] [--on user@host]`

1. Create executor
2. **Plan phase (local):**
   - Fetch recipe YAML(s) from GitHub
   - Parse and validate
   - Prompt user for each recipe's config
   - Generate compose + .env in memory
   - Resolve host ports (check state for conflicts)
3. **Execute phase (via executor):**
   - Hardening if not done
   - Install Docker if needed
   - Install Caddy if needed
   - For each recipe: mkdir, write files, add Caddy block, docker compose up, health check, update state
   - Reload Caddy once at the end (health checks use localhost)
4. Print success with URLs

### `bunkr uninstall <recipe> [--purge] [--on user@host]`

1. Verify recipe is installed (from state)
2. If `--purge`, confirm with user
3. `docker compose down` (with `-v` if purge)
4. Remove Caddy block, reload
5. Remove `/opt/bunkr/<recipe>/`
6. Update state

### `bunkr list`

Fetch `index.yaml`, display table (name, description, version). No executor needed.

### `bunkr status [--on user@host]`

Load state, `docker compose ps` for each recipe, display table (name, version, domain, status, port).

### `bunkr update <recipe> [--on user@host]`

Fetch latest recipe, compare version, if newer: regenerate compose, pull image, recreate, health check, update state.

### `bunkr self-update`

Check GitHub Releases API, download if newer, replace binary.

## VPS Directory Structure

```
/etc/bunkr/
  state.json
  recipes/              # cached recipe YAML (fallback only)

/opt/bunkr/
  <recipe>/
    docker-compose.yml
    .env

/etc/caddy/
  Caddyfile             # bunkr appends/removes blocks
```

## Error Handling

- No automatic rollback on failure
- Print what completed and what failed
- Print actionable cleanup hints on failure
- All executor.Run() calls capture stderr in error messages
- Recipe YAML validated before any execution (fail fast)
- DNS check: warn if domain doesn't resolve to server IP, but don't block

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/fatih/color` | Colored terminal output |
| `golang.org/x/crypto/ssh` | SSH client for RemoteExecutor |
| `golang.org/x/term` | Password input (hidden) |
| `gopkg.in/yaml.v3` | YAML parsing |

## Build & Release

- GoReleaser: `linux/amd64` and `linux/arm64`
- Version via ldflags: `-X main.version=...`
- GitHub Actions: tag push → GoReleaser → GitHub Releases
- Install script: `curl -fsSL https://get.bunkr.sh | sh`

## Testing (v1)

- Unit tests with mock executor per package
- Mock records commands, returns preset outputs
- Test: recipe parsing, compose gen, env gen, Caddyfile manipulation, state management, port allocation, hardening step commands
- No integration tests — manual testing on real VPS

## Scope (v1 exclusions)

- No web UI
- No backup/restore
- No multi-server management
- No custom Docker networks per recipe
- No cross-recipe dependencies
- Linux only (amd64, arm64)
