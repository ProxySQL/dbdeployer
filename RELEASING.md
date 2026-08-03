# Releasing dbdeployer

Releases are automated with [GoReleaser](https://goreleaser.com) (config in
`.goreleaser.yaml`). Pushing a version tag builds the binaries and publishes the
GitHub release — there is no manual binary-build step.

## How the version is set

The version is **not** edited in source. It is injected at build time from the git
tag via `-ldflags` into `common.VersionDef` (see `common/version.go`). Local builds
that are not on a tag report `dev`.

## Release steps

### 1. Update the changelog

Move the items under `## Unreleased` in `CHANGELOG.md` into a new versioned section:

```
## X.Y.Z	DD-Mon-YYYY
```

Keep the existing category headings (`NEW FEATURES`, `BUGS FIXED`, `CI`,
`DOCUMENTATION`, `SECURITY`) and leave an empty `## Unreleased` at the top for
future entries.

### 2. Commit to master

```bash
git add CHANGELOG.md
git commit -m "chore: prepare release vX.Y.Z"
git push origin master
```

### 3. Tag and push

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Pushing the tag triggers `.github/workflows/release.yml`, which runs GoReleaser to:

- build `linux` and `darwin` binaries for `amd64` and `arm64`,
- package them as `.tar.gz` archives,
- generate `checksums.txt`,
- create the GitHub release.

### 4. Write the release notes

GoReleaser seeds the release body from the commit subjects since the previous tag.
That list is a changelog, not a readable announcement — replace it with a short,
user-friendly summary of what changed and why it matters: lead with the headline fix
or feature, call out any behavior changes, and confirm the upgrade path.

```bash
gh release edit vX.Y.Z --notes-file release-notes.md
```

### 5. Verify

```bash
gh release view vX.Y.Z
```

Confirm the binaries and `checksums.txt` are attached and that the release notes read
as intended.

## Versioning

dbdeployer follows [Semantic Versioning](https://semver.org):

- **patch** (`X.Y.Z`) — bug fixes only, no behavior/API changes,
- **minor** (`X.Y.0`) — new features, backward compatible,
- **major** (`X.0.0`) — breaking changes.

## Notes

- `common/COMPATIBLE_VERSION` is a legacy marker for archive/template compatibility;
  it is not part of the release flow and is not bumped automatically.
- For local/dev builds (not releases), see `scripts/build.sh` and `CONTRIBUTING.md`.
