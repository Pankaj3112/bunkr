# Post-Hardening UX Improvements

## Problem

After hardening a VPS, bunkr changes the SSH user to `bunkr`, the port to 2222, and disables root login. Users aren't told how to connect afterward, and if they use `root@ip` on subsequent commands, they get a confusing connection error.

## Solution

Two changes:

### 1. Post-hardening summary

After hardening completes, display a concise summary showing the new SSH connection details. Shown in both `bunkr init` and `bunkr install`.

```
  Server hardened successfully!

  SSH access changed:
    User:  bunkr (root login disabled)
    Port:  2222
    Auth:  SSH key only

  Connect with:
    ssh bunkr@167.71.50.23 -p 2222

  For bunkr commands, use:
    bunkr <command> --on bunkr@167.71.50.23:2222
```

### 2. Auto-reconnect on SSH failure

When `NewRemoteExecutor` fails to connect (e.g., `root@ip:22`), automatically try `bunkr@ip:2222` as a fallback. If the fallback succeeds, print an info message and continue transparently.

## Files to modify

- `internal/executor/remote.go` — fallback connection logic
- `cmd/bunkr/init.go` — post-hardening summary
- `cmd/bunkr/install.go` — post-hardening summary
- `internal/ui/ui.go` — connection info helper (optional)

## Edge cases

- User already uses `bunkr@ip:2222` — no fallback needed
- Custom SSH port — fallback only tries default 2222; post-hardening message shows the correct port
- Non-hardened server with SSH issues — fallback fails harmlessly, original error shown
