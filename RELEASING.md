# Releasing

Go has no package registry to upload to: a Go module version is a git tag.
Once `vX.Y.Z` is pushed, `go get github.com/nusii/nusii-go@vX.Y.Z` works, and
the [Release workflow](.github/workflows/release.yml) creates the GitHub
release and makes sure [pkg.go.dev](https://pkg.go.dev/github.com/nusii/nusii-go)
lists the version. Nobody needs a token to release.

## Releasing a new version

1. **While you work**, describe user facing changes in `CHANGELOG.md`, under
   `## Unreleased`:

   ```markdown
   ## Unreleased

   - Add `ProposalsService.Duplicate`.
   - Fix the type of `Proposal.DisplayDate`.
   ```

   Commit everything: the release script refuses to run with uncommitted changes.

2. **Pick the version** following [semver](https://semver.org) and run one of:

   ```bash
   go run scripts/release.go patch   # 0.1.0 -> 0.1.1  bug fixes
   go run scripts/release.go minor   # 0.1.0 -> 0.2.0  new features (or breaking changes while on 0.x)
   go run scripts/release.go major   # 0.1.0 -> 1.0.0  breaking changes, see below
   ```

   It refuses to run with an empty `## Unreleased` section, then runs gofmt,
   go vet and the tests. It writes the version to `version.go` (sent in the
   `User-Agent` header), moves the notes under `## x.y.z (date)` in the
   changelog, commits as "Release vx.y.z" and creates the `vx.y.z` tag.

   For the very first release, pass the version explicitly:
   `go run scripts/release.go 0.1.0`.

3. **Push** the commit and the tag:

   ```bash
   git push --follow-tags
   ```

The workflow then checks that the tag matches `version.go`, runs the tests
again, creates a [GitHub release](https://github.com/Nusii/nusii-go/releases)
with the changelog notes, and requests the version from proxy.golang.org.
Follow it with `gh run watch`, and confirm with:

```bash
go list -m github.com/nusii/nusii-go@latest
```

### Prereleases

```bash
go run scripts/release.go 0.2.0-beta.1
git push --follow-tags
```

`go get github.com/nusii/nusii-go@latest` skips prereleases;
`go get github.com/nusii/nusii-go@v0.2.0-beta.1` installs one. The GitHub
release is marked as a prerelease.

## Versions are permanent

proxy.golang.org keeps a copy of every version it has served, forever, and
`go.sum` files pin its exact content. So:

- **Never move or re-create a tag** that has been pushed. Release a new version instead.
- **A broken version went out**: release a fixed version, then retract the broken
  one by adding a `retract` directive to `go.mod` in that fix:

  ```
  retract v0.3.0 // Sections.Copy drops the position.
  ```

  `go get` then skips the retracted version and `go list -m -retracted` explains why.

## Breaking changes after 1.0

From v2 on, Go requires the major version in the module path:
`github.com/nusii/nusii-go/v2`. Change the `module` line in `go.mod`, update
imports in the tests and README, and release `v2.0.0`. Stay on 0.x until the
API of this package is settled: breaking changes are allowed in 0.x minor
versions.

## When something goes wrong

- **The workflow fails** (a test, a flaky network...): fix the cause on `main`
  if needed, then re-run the workflow from the Actions tab. Re-running is safe:
  a release that already exists is skipped. The tag itself is already usable
  by `go get`.
- **"Tag does not match the Version constant"**: the tag was not created by
  `scripts/release.go`. If nobody has fetched it yet, delete it
  (`git push --delete origin vx.y.z`, `git tag -d vx.y.z`) and use the script.
  Otherwise release the next version.
