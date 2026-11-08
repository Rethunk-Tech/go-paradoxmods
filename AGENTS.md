# AGENTS.md

`paradoxmods` is a Go client for the Paradox Mods API: search a game's mods, read one mod's details, its versions and
its download address. Mod managers and launchers import it; it imports no consumer package, and never learns what a
caller does with an answer.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH.

## Shape

- **`Client.do` is the only path to the network.** It sets the timeout, the cap, the redirect policy and the decoding,
  so a new call is a new method that fills a body or params and decodes into a local anonymous struct, never a second
  `http.Client` call. Reason: every guarantee in SECURITY.md lives in that one function.
- **`answer` is decoded into every reply, and `answer.err` is the only refusal check.** The API answers a refusal with
  200 and its own `result` field, so a status code alone cannot tell success from a refusal. Reason: a call that skips
  the check returns zero values as if the API had agreed.
- **`maxBody` is policy, `MaxPageSize` is the API's.** The cap is ours to tune and can move; the page size is the API's
  limit, and asking for more gets a page at the cap instead of a rejection. Reason: a caller passing 500 wants a big
  page, not an error.
- **Exported types carry only the fields a caller reads.** Nothing else is decoded, and unknown JSON is ignored, so an
  API addition does not change this package's surface. Reason: every field here is a promise kept for the life of the
  module.
- **`modIDPattern` is the only id check**, reached through `ValidID`, and every call that takes an id refuses one that
  fails it. Reason: an id enters a request and, through `PageURL`, a URL path; one predicate is one place to get that
  right.
- **`Sort` values are this package's, mapped in `sortFields`.** The mapping data is the API's field names; the
  constants are ours, and an unknown value sends nothing. Reason: callers name an order, not the API's two fields.
- **No user agent is invented.** `Config.UserAgent` is what the caller says, or nothing. Reason: the client is a
  library, and the caller is the one making the request.

## Files

| File | Holds |
| --- | --- |
| `client.go` | `Config`, `New`, the four calls, `Client.do`, `APIError`, `BaseURL`, `MaxPageSize`, `ValidID`, `PageURL`, `DownloadURL` |
| `types.go` | `Sort` and its mapping, `ModList`, `Mod` and `Mod.Required`, `Requirement`, `ChangeEntry`, `Version` and `Version.Label` |
| `doc.go` | Package documentation. |
