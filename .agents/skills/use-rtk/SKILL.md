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
