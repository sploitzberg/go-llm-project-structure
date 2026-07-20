---
name: use-rtk
description: Use RTK in Zed to reduce verbose output from Go tests, builds, Taskfile validation, and read-only Git commands. Load this skill before running supported validation commands when summarized output is sufficient; use raw commands for writes, interactive workflows, or detailed debugging.
---

# RTK for Zed Agent Workflows

RTK is an optional token-optimized command proxy. It preserves the underlying command's behavior while presenting compact output.

## Availability

Check whether RTK is installed:

```sh
which rtk
rtk --version
```

If it is unavailable, run the underlying command directly. Do not install RTK or run `rtk init` unless the user requests it.

## Use RTK for verbose validation

Prefer RTK for supported commands whose complete output is not needed:

```sh
rtk git status
rtk git diff
rtk git log
rtk go test ./...
rtk go build ./...
rtk task test
rtk task integration
rtk task benchmark
rtk task lint
```

Start with the narrowest relevant test or build, then broaden validation as confidence increases.

## Use raw commands when detail or writes matter

Do not use RTK for:

- interactive commands or commands that may prompt,
- formatting, generation, module updates, migrations, or other file-modifying commands,
- commands where filtering could hide diagnostics needed to identify a failure,
- long-running servers or file watchers.

Examples that should normally run directly:

```sh
task fmt
go mod tidy
bash -x scripts/ci/ci.sh
```

If an RTK command fails and its summary is insufficient, rerun the focused command directly. Raw proxy mode is also available when RTK analytics are useful:

```sh
rtk proxy go test -race -count=1 ./...
rtk proxy git --no-pager diff
```

## Preserve standard tool caches

Run the repository's Taskfile commands without overriding cache environment variables. Do not set `GOCACHE`, `GOMODCACHE`, `GOPATH`, or `GOLANGCI_LINT_CACHE` to `/tmp` or a project-local directory solely to bypass the Zed terminal sandbox.

Redirecting caches changes normal developer behavior, can force repeated module downloads, and creates large hash-sharded directory trees. In this environment `/tmp` may be cleared between terminal calls, so redirected caches are recreated instead of reused.

When a command fails because the sandbox cannot write to a standard cache:

1. Inspect the configured locations with read-only commands such as `go env GOCACHE GOMODCACHE GOPATH` and, when needed, `golangci-lint cache status`.
2. Resolve those values to literal paths before making a tool call; do not use shell substitution in terminal commands.
3. Request narrow filesystem write access to the existing standard cache directories. This commonly includes the Go build cache, module cache, checksum database under `GOPATH/pkg/sumdb`, and golangci-lint cache. Request `GOPATH/bin` only when an explicitly requested tool installation needs it.
4. Rerun the original command unchanged, for example `task test`, `task lint`, or `CI=true task ci`. If downloading is required, request only the necessary network hosts separately.

Do not persist sandbox-specific cache overrides in the Taskfile, scripts, editor settings, or documentation. If access to the standard cache is not granted, report the validation limitation instead of silently rerouting caches.

Temporary directories remain appropriate for bounded fixtures and staged-index snapshots when they are cleaned with a reliable `trap`; they must not be used as long-lived compiler, module, or linter caches.

## Zed integration

Zed has no dedicated `rtk init` target. Do not install another editor's hooks or settings as a substitute. The checked-in `.zed/tasks.json` tasks prefer RTK for appropriate commands and fall back to raw commands when RTK is absent.

## Meta commands

```sh
rtk gain
rtk gain --history
rtk gain --daily
rtk gain --weekly
rtk discover
rtk session
```

## Reporting

When reporting validation, name the command actually run, including the `rtk` prefix. If RTK was unavailable and a raw fallback was used, say so briefly.
