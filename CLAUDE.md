# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and test

```bash
go build -o dbdeployer .              # plain build
./scripts/build.sh linux              # release binary (also: OSX | all)
MKDOCS=1 ./scripts/build.sh all       # docs-enabled binary (adds `docs` build tag)

./test/go-unit-tests.sh               # unit tests across all packages
go test ./unpack/... ./cmd/...        # narrower run for one or two packages
./scripts/sanity_check.sh             # gofmt + go vet + copyright headers + staticcheck + gosec
./scripts/sanity_check.sh fmt         # one check at a time: fmt | vet | copyright | static | secure | version
```

`./test/go-unit-tests.sh` auto-skips `sandbox/`, `ts/`, `ts_static/` if `$SANDBOX_BINARY` doesn't contain unpacked MySQL installations — those packages need real binaries. The mock-based and functional suites under `test/` (`functional-test.sh`, `mock/*.sh`, `unpack-test.sh`) require actual MySQL tarballs and can run for tens of minutes; only invoke them when explicitly working on that area.

`./scripts/sanity_check.sh copyright` enforces that every `.go` and `.sh` file begins with the `DBDeployer` copyright header — new files must include it or CI will fail.

## How the binary is wired

`main.go` → `cmd.Execute()` (Cobra). All subcommands live in `cmd/` (one file per top-level command: `unpack.go`, `deploy.go`, `single.go`, `replication.go`, `init.go`, etc.). The Cobra layer is thin glue: it parses flags and delegates real work to packages in `ops/` (one file per command worth of business logic). Put new logic in `ops/`, not `cmd/`, so it stays unit-testable without Cobra in scope.

Two filesystem roots, both configurable, both live in `defaults/defaults.go`:

- `SandboxBinary` — default `$HOME/opt/mysql` — where unpacked server binaries land (`unpack` writes here, `deploy` reads here).
- `SandboxHome` — default `$HOME/sandboxes` — where running sandbox directories are created.

`dbdeployer init` is the canonical command that creates both directories (plus downloads a tarball and installs shell completion). When a command needs one of these to exist, point users at `dbdeployer init` (with `--skip-all-downloads` if appropriate) rather than silently creating directories.

## Provider architecture

Multi-database support goes through `providers/provider.go`, which defines a `Provider` interface (`Name`, `ValidateVersion`, `FindBinary`, `CreateSandbox`, `CreateReplica`, …) and a `Registry`. Concrete implementations live in `providers/mysql/`, `providers/postgresql/`, `providers/proxysql/`. The `--provider=` flag on `unpack`/`deploy` selects which one runs. MySQL remains the default and most-exercised path; PostgreSQL and ProxySQL are newer and have narrower topology support (see README's capabilities matrix).

## Sandbox generation

`sandbox/` is where MySQL/MariaDB sandbox directories get materialized from Go-embedded shell templates under `sandbox/templates/<topology>/` (`single/`, `replication/`, `group/`, `galera/`, `pxc/`, `ndb/`, `cluster/`, `import/`, `multiple/`, `tidb/`, `mock/`). Each topology has its own `*_templates.go` plus an implementation file (e.g. `replication.go`, `group_replication.go`). Mock variants under `sandbox/templates/mock/` exist so tests can simulate sandboxes without real binaries.

## Generated docs and README

`README.md`, `docs/dbdeployer_completion.sh`, and `docs/API/*` are **generated** — do not edit by hand. The source is `mkreadme/readme_template.md` plus the templating tool in `mkreadme/make_readme.go`; rebuild with `./mkreadme/build_readme.sh`. Files under `cmd/tree.go` and similar use the `//go:build docs` tag so the docs-generation code only ships in `MKDOCS=1` builds.

## Common conventions worth respecting

- Cobra commands surface user-facing errors via `common.Exit` / `common.Exitf` (multi-line messages take multiple string args). Error messages are part of the UX — when adding a "missing prerequisite" message, include the exact command (e.g. `dbdeployer init --skip-all-downloads`) the user should run.
- File and directory helpers in `common/fileutil.go` (`DirExists`, `FileExists`, `AbsolutePath`, `Mkdir`, `RmdirAll`) are preferred over raw `os` calls — they already handle the symlink/path-resolution edge cases this codebase has hit before.
- Error message constants for the common "X not found" / "X already exists" patterns live in `globals/globals.go` (`ErrDirectoryNotFound`, `ErrFileNotFound`, …) — reuse them rather than reinventing strings.
