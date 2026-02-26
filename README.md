# bunkr

Harden a VPS and deploy self-hosted apps in one command.

Bunkr takes a fresh VPS and turns it into a hardened server running self-hosted applications. It handles security hardening, Docker setup, reverse proxy configuration, and app deployment, all from your local machine over SSH.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/Pankaj3112/bunkr/main/scripts/install.sh | sudo sh
```

Or download a binary from [GitHub Releases](https://github.com/Pankaj3112/bunkr/releases/latest).

The bunkr CLI runs on **Linux, macOS, and Windows** (amd64 and arm64). The target VPS must be running **Ubuntu** (the only supported server OS for now).

## Quick start

```sh
# From your laptop, deploy to a remote VPS
bunkr install uptime-kuma --on root@167.71.50.23

# Or run directly on the VPS itself (no --on flag needed)
bunkr install uptime-kuma
```

That's it. Bunkr will:

1. Harden the server (SSH, firewall, fail2ban, kernel params, swap)
2. Install Docker and Caddy (or Tailscale for private apps)
3. Deploy the app with HTTPS

After hardening, bunkr tells you how to connect:

```
  SSH access changed:
    User:  bunkr (root login disabled)
    Port:  2222
    Auth:  SSH key only

  Connect with:
    ssh bunkr@167.71.50.23 -p 2222
```

For subsequent commands, you can use the hardened credentials, or just keep using `root@ip`. Bunkr auto-reconnects if it detects the server was hardened.

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `bunkr init` | Harden a server (no app install) | `bunkr init --on root@167.71.50.23` |
| `bunkr install` | Harden + install app(s) | `bunkr install ghost --on root@167.71.50.23` |
| `bunkr list` | Show available apps | `bunkr list` |
| `bunkr status` | Show installed apps and status | `bunkr status --on bunkr@167.71.50.23:2222` |
| `bunkr update` | Update an installed app | `bunkr update ghost --on bunkr@167.71.50.23:2222` |
| `bunkr uninstall` | Remove an installed app | `bunkr uninstall ghost --on bunkr@167.71.50.23:2222` |
| `bunkr self-update` | Update bunkr itself | `sudo bunkr self-update` |

### Flags

- `--on <user@host>` - Target a remote server over SSH (e.g., `root@167.71.50.23`)
- `--ssh-port <port>` - Set the SSH port during hardening (default: 2222, used with `init` and `install`)
- `--purge` - Also remove app data when uninstalling

## Available apps

| App | Description | Access |
|-----|-------------|--------|
| [uptime-kuma](https://github.com/louislam/uptime-kuma) | Uptime monitoring | Public |
| [ghost](https://ghost.org) | Publishing platform | Public |
| [plausible](https://plausible.io) | Privacy-friendly analytics | Public |
| [openclaw](https://openclaw.ai/) | Personal AI assistant | Private |

**Public** apps are exposed to the internet with HTTPS via Caddy reverse proxy. You'll be prompted for a domain name during install.

**Private** apps are only accessible over your [Tailscale](https://tailscale.com) network. During install, bunkr sets up Tailscale on the server and gives you an auth URL to connect it to your tailnet. The app is then available at `https://<server>.your-tailnet.ts.net`.

## What hardening does

Bunkr applies 7 hardening steps to your server:

| Step | What it does |
|------|-------------|
| Sudo user | Creates `bunkr` user with SSH key access and sudo privileges |
| SSH | Disables root login and password auth, changes port to 2222 |
| Firewall | Configures UFW to only allow SSH, HTTP, and HTTPS |
| Fail2ban | Installs intrusion prevention |
| Unattended upgrades | Enables automatic security updates |
| Kernel hardening | Applies sysctl security parameters |
| Swap | Creates 1GB swap file |

All steps are idempotent and safe to run multiple times.

## How it works

Bunkr uses a two-phase execution model:

1. **Plan** (runs locally) - Fetch recipes, prompt for config (domain, etc.), generate Docker Compose files
2. **Execute** (runs on server) - Write files, install dependencies, start containers

On the server, each app gets:

- A Docker Compose stack at `/opt/bunkr/<app>/`
- HTTPS via Caddy reverse proxy (public apps) or Tailscale Serve (private apps)
- State tracked in `/etc/bunkr/state.json`

## Build from source

```sh
git clone https://github.com/Pankaj3112/bunkr.git
cd bunkr
go build -o bunkr ./cmd/bunkr
```

Requires Go 1.25+.

## License

MIT
