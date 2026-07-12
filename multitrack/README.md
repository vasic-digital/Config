# config/multitrack — SFTP multi-track development configuration

**Revision:** 1
**Last modified:** 2026-07-11T16:59:00Z
**Status:** active
**Classification:** project-specific (§11.4.17) — per-host track layout, aliases, and
conductor designation for the SFTP project on host `nezha`.

## Overview

This directory holds the per-host configuration consumed by the
constitution's multi-track engine (`constitution/scripts/multitrack/`,
inherited BY REFERENCE per §11.4.28(B)/§11.4.177). The engine is
project-agnostic — this directory supplies the project-specific DATA
(tracks, aliases, conductor designation) the engine reads at runtime.

**One file per host**: `config/multitrack/<hostname>.yaml`.
The current host is `nezha` (file: `nezha.yaml`).

## How to configure tracks

1. **Edit `config/multitrack/<hostname>.yaml`** — add or modify tracks
   under the `tracks:` block. Each track requires:
   - `id`: unique track identifier (e.g., `track-1`, `track-2`)
   - `role`: `main` (conductor track) or `feature` (worker track)
   - `branch`: the git branch this track works on (usually `main`)
   - `drive_serial`: logical identifier (for projects without dedicated
     physical drives, use a descriptive placeholder)
   - `mount`: the filesystem path (e.g., `/mnt/track1`)
   - `fs`: filesystem type (usually `btrfs`)
   - `focus`: human-readable description of the track's purpose

2. **Provision the track directory** — for each track, create the
   directory structure:
   ```bash
   mkdir -p /mnt/track<N>/sftp/
   ```
   (Or mount a dedicated drive at `/mnt/track<N>` if using physical
   drives per §11.4.179 corruption-isolated clones.)

3. **Clone the project into each track** — each track gets its own
   `.git` (corruption-isolated per §11.4.179):
   ```bash
   git clone /run/media/milosvasic/DATA4TB/Projects/sftp /mnt/track1/sftp/
   git clone /run/media/milosvasic/DATA4TB/Projects/sftp /mnt/track2/sftp/
   ```

4. **Add/update aliases** in the `aliases:` block as new Claude Code
   subscriptions or provider accounts become available. NAMES ONLY —
   never keys/tokens (§11.4.10).

5. **Commit the config** — the config is tracked in git. It describes
   the INTENDED track layout; the operator provisions the physical
   resources (drives/directories) separately.

## How to start the multi-track ruler

### Prerequisites

- The per-host config file exists and is valid YAML
- At least one Claude Code alias is configured and authenticated
- Track directories exist at the declared mount paths

### Bootstrap (one-time, idempotent)

```bash
cd /run/media/milosvasic/DATA4TB/Projects/sftp
MT_REPO_ROOT="$PWD" \
  bash constitution/scripts/multitrack/multitrack_bootstrap.sh
```

This installs the cwd-hook, validates the config, and seeds the
orchestrator state. It is idempotent — safe to run multiple times.

### Start the ruler (supervisor)

```bash
MT_REPO_ROOT=/run/media/milosvasic/DATA4TB/Projects/sftp \
MT_SUPERVISOR_WATCH_ENABLE=1 \
  bash constitution/scripts/multitrack/multitrack_bootstrap.sh
```

Or start the supervisor daemon directly:

```bash
MT_REPO_ROOT=/run/media/milosvasic/DATA4TB/Projects/sftp \
  bash constitution/scripts/multitrack/multitrack_supervisor.sh watch
```

### Verify status

```bash
MT_REPO_ROOT=/run/media/milosvasic/DATA4TB/Projects/sftp \
  bash constitution/scripts/multitrack/multitrack_alias_orchestrator.sh status
```

### Map aliases to worktrees

```bash
MT_REPO_ROOT=/run/media/milosvasic/DATA4TB/Projects/sftp \
  bash constitution/scripts/multitrack/multitrack_resolve_worktree.sh map
```

## Current track layout (host: nezha)

| Track | Mount | Project Path | Branch | Role |
|---|---|---|---|---|
| track-1 | /mnt/track1 | /mnt/track1/sftp/ | main | Primary — conductor + commit/push + reviews |
| track-2 | /mnt/track2 | /mnt/track2/sftp/ | main | Secondary — features + fixes + tests |

## References

- **§11.4.187** — Automatic multi-track ruler orchestration (the
  engine this config feeds)
- **§11.4.178** — Track-qualified identity (`<project>__<track>__<role>`)
  for session/lock/log namespace isolation
- **§11.4.176** — Exactly-once claim registry + deadlock-proof
  device-lock arbitration
- **§11.4.179** — Corruption-isolated parallel git streams (own-`.git`
  clones per track)
- **§11.4.182** — Track+branch work-stream identity label
  `(T<N>/<branch> - <alias>)` on every agent/subagent dispatch
- **§11.4.192** — Continuous multi-track auto-backfill (FREE track
  immediately re-assigned)
- **§11.4.177** — Developer-tooling project-decoupling (engine
  inherited BY REFERENCE, never copied)
- **§11.4.103** — Continuous parallel-stream working routine (the
  operational model the tracks enable)
- Engine source: `constitution/scripts/multitrack/`
- Engine documentation: `constitution/scripts/multitrack/README.md`
