# HUMANS.md

Run and use `go-paradoxmods`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-paradoxmods
```

```go
package main

import (
 "context"
 "fmt"
 "os"

 "github.com/Rethunk-Tech/go-paradoxmods"
)

func main() {
 ctx := context.Background()
 c := paradoxmods.New(paradoxmods.Config{UserAgent: "myapp/1.0 (+https://example.com)"})

 list, err := c.Search(ctx, paradoxmods.SearchQuery{
  Game: "cities_skylines_2", Text: "anarchy", Page: 1, Size: 20, Sort: paradoxmods.SortDownloads,
 })
 if err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
 fmt.Println(list.Total, "matches")
 for _, m := range list.Mods {
  fmt.Println(m.ID, m.DisplayName, paradoxmods.PageURL(m.ID))
 }
 if len(list.Mods) == 0 {
  return
 }

 id := list.Mods[0].ID
 m, err := c.Mod(ctx, "cities_skylines_2", id, "Windows")
 if err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
 fmt.Println(m.DisplayName, "needs", m.Required())

 versions, err := c.Versions(ctx, id)
 if err != nil {
  fmt.Fprintln(os.Stderr, err)
  os.Exit(1)
 }
 fmt.Println(paradoxmods.DownloadURL(m.Game, m.Name, versions[0].ID), versions[0].Label())
}
```

## Contracts

| Behaviour | Contract |
| --- | --- |
| `New` | Never fails; `BaseURL` empty means `BaseURL` with trailing slashes trimmed; `Config.HTTP` nil means a client with a timeout; redirects to another origin are refused |
| `Client.Search` | `Page` below 1 reads the first; `Size` zero or above `MaxPageSize` asks for `MaxPageSize`; `Text`, `Tags` and `Sort` are left out when empty, so the API applies its own; a mod's id is not checked |
| `Client.Mod` | Refuses an id `ValidID` rejects, and a mod of another game when `game` is not empty; returns the newest version's requirements as the API lists them |
| `Client.Versions` | Refuses an id `ValidID` rejects; newest first by version number; `Label` is the author's name, or the number when there is none |
| `Mod.Required` | Only requirements whose product type is a mod, whose id is valid, and which is not the mod itself: downloadable content is left out |
| Failures | A refusal the API sends with a 200 is an `*APIError` carrying its code and message; another status is a `*httpkit.StatusError`; a transport failure is redacted; the answer is capped at 16 MiB |
| `PageURL`, `DownloadURL` | Addresses only, no request; `PageURL` is "" for an id that is not a mod id; `DownloadURL` does not escape its arguments, so a caller checks the file name it passes |
| Anything a call returns | Empty on error: a failed call never returns a partial answer as if it were one |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
