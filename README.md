<h1 align="center">go-paradoxmods</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

A client for the Paradox Mods catalogue: search, mod details and releases. Package `paradoxmods`, module `github.com/Rethunk-Tech/go-paradoxmods`.

For mod managers and launchers that browse Paradox Mods. It speaks the public API directly: one type per answer,
orders named by a constant rather than the API's field names, and every failure an error rather than a zero value.
An oversized or hostile answer fails instead of exhausting memory.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-paradoxmods
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- `Client.Search` lists a game's mods by text, tag and one of six orders, with the page capped at the API's largest.
- `Client.Mod` reads one mod's page, refusing an id that is not a mod id and a mod of another game.
- `Mod.Required` gives the mods the newest version needs, leaving out the game's downloadable content.
- `Client.Versions` lists a mod's versions newest first; `Version.Label` is the author's name or the number.
- `PageURL` and `DownloadURL` build the site page and the archive address; `ValidID` checks an id first.
- Answers are size-capped, redirects off the API's origin are refused, and a refusal is an `*APIError` carrying its code and message.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.
