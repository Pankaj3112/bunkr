# bunkr

Harden a VPS and deploy self-hosted apps in one command.

Bunkr takes a fresh VPS and turns it into a hardened server running self-hosted applications. It handles security hardening, Docker setup, reverse proxy configuration, and app deployment — all from your local machine over SSH.

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
2. Install Docker and Caddy
3. Deploy the app with HTTPS via Caddy reverse proxy

After hardening, bunkr tells you how to connect:

```
  SSH access changed:
    User:  bunkr (root login disabled)
    Port:  2222
    Auth:  SSH key only

  Connect with:
    ssh bunkr@167.71.50.23 -p 2222
```

For subsequent commands, use the hardened credentials — or don't. Bunkr auto-reconnects if it detects the server was hardened.

## Commands

### `bunkr init`

Harden a server without installing any apps.

```sh
bunkr init --on root@167.71.50.23
bunkr init --on root@167.71.50.23 --ssh-port 3333  # custom SSH port
```

### `bunkr install`

Harden (if needed) and install one or more apps.

```sh
bunkr install uptime-kuma --on root@167.71.50.23
bunkr install ghost plausible --on root@167.71.50.23  # multiple apps
```

### `bunkr list`

Show available apps.

```sh
bunkr list
```

### `bunkr status`

Show installed apps and their status.

```sh
bunkr status --on bunkr@167.71.50.23:2222
```

### `bunkr update`

Update an installed app to the latest version.

```sh
bunkr update ghost --on bunkr@167.71.50.23:2222
```

### `bunkr uninstall`

Remove an installed app.

```sh
bunkr uninstall ghost --on bunkr@167.71.50.23:2222
bunkr uninstall ghost --purge --on bunkr@167.71.50.23:2222  # also remove data
```

### `bunkr self-update`

Update bunkr itself.

```sh
sudo bunkr self-update
```

## Available apps

| App | Description |
|-----|-------------|
| [uptime-kuma](https://github.com/louislam/uptime-kuma) | Uptime monitoring |
| [ghost](https://ghost.org) | Publishing platform |
| [plausible](https://plausible.io) | Privacy-friendly analytics |

## What hardening does

Bunkr applies 7 hardening steps to your server:

| Step | What it does |
|------|-------------|
| Sudo user | Creates `bunkr` user with SSH key access, adds to sudo group |
| SSH | Disables root login and password auth, changes port to 2222 |
| Firewall | Configures UFW — allows SSH, HTTP, HTTPS only |
| Fail2ban | Installs intrusion prevention |
| Unattended upgrades | Enables automatic security updates |
| Kernel hardening | Applies sysctl security parameters |
| Swap | Creates 1GB swap file |

All steps are idempotent — safe to run multiple times.

## How it works

Bunkr uses a two-phase execution model:

1. **Plan** (runs locally): Fetch recipes, prompt for config (domain, etc.), generate Docker Compose files
2. **Execute** (runs on server via SSH): Write files, install dependencies, start containers

On the server, each app gets:
- A Docker Compose stack at `/opt/bunkr/<app>/`
- A Caddy reverse proxy block for HTTPS
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
