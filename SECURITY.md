# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

- **The API's JSON answers**, for every call. A crafted answer must not read more than 16 MiB, must not panic the
  decoder, and must not make a later call reach a different origin: the body is read through a cap and redirects are
  refused unless they stay on the request's own scheme and host.
- **Ids and version numbers passed to `Client.Mod` and `Client.Versions`.** A crafted id must not reach the API as
  anything but a short decimal number, and must not become a path segment of a page address: `ValidID` is the one
  check, and `PageURL` answers "" rather than an address for an id it rejects.
- **The file name and game name passed to `DownloadURL`.** These come from an API answer and are interpolated into an
  address without escaping, so a crafted answer must not be able to point the download at another host or outside the
  mod's own directory. `Mod.Name` is the field the API uses for this; `ValidID` covers ids, not names.
- **Error text.** A transport failure must not quote a URL's userinfo or query in an error string; it is redacted
  before it is returned.

## What is out of scope

The caller is trusted to:

- Pass its own `Config.UserAgent`; the client sends what it is given and invents none.
- Check `DownloadURL`'s `name` argument itself, when it does not come straight from `Mod.Name` of an answer about
  that same mod. The address is built by concatenation and does not escape anything.
- Decide that an answer is trustworthy enough to act on: this package proves that a response is well-formed and
  within bounds, not that the mod it describes is safe to install.
- Run its own TLS trust, proxies and timeouts inside `Config.HTTP`, and hand over a client it is happy to share.

Intended behaviour, not a vulnerability: a refusal the API sends with a 200 status (`*APIError`), an empty page, an
unknown `Sort` value sending no order at all, and a page size above `MaxPageSize` being reduced to it.
