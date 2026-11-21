# Changelog

Notable changes to `go-paradoxmods`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.1] - 2026-11-21

### Changed

- Requests go through `httpkit.DoJSON`, so a 429 is a `*httpkit.RateLimitError` carrying its Retry-After reset.

### Fixed

- A wrong or empty mod answer, non-web images and spaced download names are refused.
- A crafted game or name that leaves `DownloadURL`'s directory is rejected.

## [0.1.0] - 2026-11-08

### Added

- `Client.Search` lists a game's mods by text, tag, page and one of six `Sort` orders.
- `Client.Mod` reads one mod's page, with `Mod.Required` for the mods its newest version needs.
- `Client.Versions` lists a mod's versions newest first, each labelled by `Version.Label`.
- `PageURL`, `DownloadURL` and `ValidID` cover the site page, the archive address and mod ids.
- Answers are size-capped, cross-origin redirects are refused, and a refusal becomes an `*APIError`.

[Unreleased]: https://github.com/Rethunk-Tech/go-paradoxmods/compare/v0.1.1...HEAD
[0.1.1]: https://github.com/Rethunk-Tech/go-paradoxmods/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Rethunk-Tech/go-paradoxmods/releases/tag/v0.1.0
