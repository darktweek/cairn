# Contributing to Cairn

Thanks for your interest! Please read this before opening a pull request.

## Developer Certificate of Origin (DCO)

Sign off your commits with `git commit -s` to certify you wrote the code (or
have the right to submit it) under the project's license.

## Contributor license grant (required for dual licensing)

Cairn is **dual-licensed** — GNU AGPL-v3 **and** a commercial license (see
[`LICENSING.md`](LICENSING.md)). To keep that model possible, **by submitting a
contribution you agree that:**

1. your contribution is provided under the project's open-source license
   (**GNU AGPL-v3**); **and**
2. you grant the project's copyright holder (**darktweek**) a perpetual,
   worldwide, non-exclusive, royalty-free, irrevocable license to use,
   reproduce, modify, sublicense and distribute your contribution, **including
   under other license terms** (such as a commercial license).

You **retain copyright** to your contribution. If you can't agree to this,
please don't submit it.

## Development

You don't need Go installed — everything builds in Docker (see the README
section *Making changes*).

Before opening a PR, make sure these pass (CI enforces them):

```
go build ./...
go vet ./...
go test ./...
govulncheck ./...
```

Guidelines:

- Keep the **frontend dependency-free** (vanilla JS, no build step).
- Database changes go through an embedded **goose** migration; never edit an
  already-released migration.
- Match the surrounding code style; keep changes focused.
- For anything security-sensitive, see [`SECURITY.md`](SECURITY.md).
