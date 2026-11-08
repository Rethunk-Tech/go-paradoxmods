# Changelog

Notable changes to `go-paradoxmods`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `Client.Search` lists a game's mods by text, tag, page and one of six `Sort` orders.
- `Client.Mod` reads one mod's page, with `Mod.Required` for the mods its newest version needs.
- `Client.Versions` lists a mod's versions newest first, each labelled by `Version.Label`.
- `PageURL`, `DownloadURL` and `ValidID` cover the site page, the archive address and mod ids.
- Answers are size-capped, cross-origin redirects are refused, and a refusal becomes an `*APIError`.

[Unreleased]: https://github.com/Rethunk-Tech/go-paradoxmods/commits/main
