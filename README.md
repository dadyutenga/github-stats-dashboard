# GitHub Stats Dashboard 🚀

A self-hosted terminal dashboard for your GitHub activity. Built in Go with ANSI colors and ASCII panels — no fancy UI libs.

## Preview

```
┌─ GitHub Stats Dashboard ──────────────────────────────────┐
│ @yourname  Your Full Name                                  │
│ Sunday, Feb 22 2026  23:41:07                             │
│ repos: 42  followers: 120  runtime: go1.21/amd64          │
└────────────────────────────────────────────────────────────┘
┌─ Activity ─────────────────────────────────────────────────┐
│   12 commits today   89 this month                         │
│                                                            │
│   ⚡ 3 day streak                                          │
└────────────────────────────────────────────────────────────┘
┌─ Commit Graph (last 7 days) ───────────────────────────────┐
│  Today │ ████████████░░░░░░░░  12                          │
│    -1d │ ████████░░░░░░░░░░░░   8                          │
│    -2d │ ██████░░░░░░░░░░░░░░   6                          │
│    -3d │ ░░░░░░░░░░░░░░░░░░░░   0                          │
│    -4d │ ████████████████░░░░  16                          │
│    -5d │ ██████░░░░░░░░░░░░░░   6                          │
│    -6d │ ████░░░░░░░░░░░░░░░░   4                          │
└────────────────────────────────────────────────────────────┘
```

## Setup

### 1. Get a GitHub Personal Access Token

1. Go to https://github.com/settings/tokens
2. Click **Generate new token (classic)**
3. Select scopes: `repo` and `read:user`
4. Copy the token

### 2. Set Environment Variables

**Linux / macOS (bash/zsh):**
```bash
export GITHUB_TOKEN=ghp_yourTokenHere
# Optional: override username (auto-detected from token if omitted)
export GITHUB_USERNAME=yourusername
```
Add to `~/.bashrc` or `~/.zshrc` to persist.

**Windows (PowerShell):**
```powershell
$env:GITHUB_TOKEN = "ghp_yourTokenHere"
# Optional:
$env:GITHUB_USERNAME = "yourusername"
```
To persist, add to your PowerShell profile (`$PROFILE`) or set via **System → Advanced → Environment Variables**.

**Windows (cmd):**
```cmd
set GITHUB_TOKEN=ghp_yourTokenHere
set GITHUB_USERNAME=yourusername
```

### 3. Build & Run

**Linux / macOS:**
```bash
git clone <this-repo>
cd github-stats-dashboard

go build -o github-dash .
./github-dash
```

**Windows:**
```powershell
git clone <this-repo>
cd github-stats-dashboard

go build -o github-dash.exe .
.\github-dash.exe
```

Or run directly (all platforms):
```bash
go run .
```

## Project Structure

```
github-stats-dashboard/
├── main.go              # Entry point, refresh loop, --web flag
├── go.mod               # No external deps!
├── config/
│   └── config.go        # Env var loading
├── api/
│   └── client.go        # GitHub REST API calls
├── renderer/
│   └── renderer.go      # ANSI terminal rendering
└── web/
    └── web.go           # Web dashboard (HTML server)
```

### 4. Web Mode (optional)

Run the dashboard as a local web page instead of a terminal UI:

```bash
go run . --web
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

To use a custom port:
```bash
go run . --web --addr :3000
```

The web dashboard auto-refreshes every 60 seconds and shows the same stats as the terminal version.

## Features

- **Daily commits** — count + 7-day bar graph
- **Monthly totals** — across all owned repos
- **Streak tracking** — consecutive days with commits
- **Top repos** — ranked by this month's commits, today's highlighted in green
- **Auto-refresh** — every 60 seconds
- **Terminal mode** — ANSI-colored terminal dashboard (default)
- **Web mode** — browser-based dashboard with `--web` flag
- **Clean exit** — Ctrl+C restores cursor and clears screen

## Notes

- Only checks your **owned, non-forked** repos (up to 20 most recently pushed to avoid rate limit hits)
- GitHub API rate limit: 5000 req/hour with auth — you're fine
- No external dependencies — pure Go stdlib + ANSI codes
