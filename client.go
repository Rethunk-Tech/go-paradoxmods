package paradoxmods

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	// BaseURL is the public API root.
	BaseURL = "https://api.paradox-interactive.com"
	// MaxPageSize is the API's largest page.
	MaxPageSize = 100

	requestTimeout = 20 * time.Second
	// maxBody caps one API answer, so a hostile response cannot grow the read without bound.
	maxBody = 16 << 20
)

const (
	siteURL    = "https://mods.paradoxplaza.com"
	contentURL = "https://modscontent.paradox-interactive.com"
	// pageOS is the operating-system segment of a mod's page address; the site serves a page per build.
	pageOS = "Windows"
)

// APIError is an answer the API refused. A refusal arrives with a 200 status and its own result field, so it is not a
// *httpkit.StatusError.
type APIError struct {
	// Code is the API's errorCode, e.g. "bad-input".
	Code string
	// Message is the API's errorMessage, the reason in words.
	Message string
}

// Error names the code and the reason.
func (e *APIError) Error() string {
	return "paradox mods refused: " + cmp.Or(e.Message, e.Code, "no reason given")
}

// Config is what a Client is built from.
type Config struct {
	// BaseURL replaces the API root; empty means BaseURL. Trailing slashes are trimmed.
	BaseURL string
	// HTTP sends every request; nil means a client with a timeout. Redirects to another origin are refused.
	HTTP *http.Client
	// UserAgent is sent with every request when set; the caller names itself.
	UserAgent string
}

// Client calls the Paradox Mods API. It holds no state beyond its configuration, so it is safe for concurrent use.
type Client struct {
	base, agent string
	hc          *http.Client
}

// New builds a Client.
func New(cfg Config) *Client {
	return &Client{
		base:  strings.TrimRight(cmp.Or(strings.TrimSpace(cfg.BaseURL), BaseURL), "/"),
		agent: cfg.UserAgent,
		hc:    httpkit.WithRedirect(cfg.HTTP, httpkit.SameOrigin),
	}
}

// ValidID reports whether id is a Paradox Mods mod id: up to nine digits.
func ValidID(id string) bool { return modIDPattern.MatchString(id) }

// PageURL is the mod's page on the site, or "" when id is not a mod id.
func PageURL(id string) string {
	if !ValidID(id) {
		return ""
	}
	return siteURL + "/mods/" + id + "/" + pageOS
}

// DownloadURL is the anonymous archive address of version n of the mod whose opaque file name is name, or "" when
// game or name is not a single path segment or n is below 1.
func DownloadURL(game, name string, n int) string {
	if n < 1 || !pathSegment(game) || !pathSegment(name) {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s/content/sources/%s_%d.zip", contentURL, game, name, name, n)
}

// pathSegment reports whether s is one URL path segment: non-empty, not "."/"..", and carrying none of the
// characters that would leave the mod directory or split the address into query or fragment.
func pathSegment(s string) bool {
	return s != "" && s != "." && s != ".." && !strings.ContainsAny(s, "/\\?#%\x00")
}

// answer is the envelope every call is wrapped in, refusal included.
type answer struct {
	Result       string `json:"result"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

func (a answer) err() error {
	if a.Result == "OK" {
		return nil
	}
	return &APIError{Code: a.ErrorCode, Message: a.ErrorMessage}
}

// searchRequest is the listing body; a zero field is left out so the API applies its own.
type searchRequest struct {
	Game   string   `json:"gameName"`
	Page   int      `json:"page"`
	Size   int      `json:"pageSize"`
	Text   string   `json:"searchQuery,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	SortBy string   `json:"sortBy,omitempty"`
	Order  string   `json:"orderBy,omitempty"`
}

// SearchQuery is one page of a mods listing.
type SearchQuery struct {
	// Game is the API's game name, e.g. "cities_skylines_2".
	Game string
	// Text matches names and descriptions; empty lists the game's mods.
	Text string
	// Page is the 1-based page to read; anything below 1 reads the first.
	Page int
	// Size is the page length; zero, and anything above MaxPageSize, reads MaxPageSize.
	Size int
	// Tags asks for mods carrying the listed tags.
	Tags []string
	// Sort is one of the Sort values; the zero value keeps the API's own order.
	Sort Sort
}

// Search lists the game's mods.
func (c *Client) Search(ctx context.Context, q SearchQuery) (ModList, error) {
	by, order := q.Sort.fields()
	body := searchRequest{
		Game: q.Game, Page: max(q.Page, 1), Size: min(max(cmp.Or(q.Size, MaxPageSize), 1), MaxPageSize),
		Text: q.Text, Tags: q.Tags, SortBy: by, Order: order,
	}
	var out struct {
		answer
		TotalCount int   `json:"totalCount"`
		Mods       []Mod `json:"mods"`
	}
	if err := c.do(ctx, http.MethodPost, "/mods", nil, body, &out); err != nil {
		return ModList{}, err
	}
	if err := out.err(); err != nil {
		return ModList{}, err
	}
	return ModList{Total: out.TotalCount, Mods: out.Mods}, nil
}

// Mod reads one mod as published for an operating system. A non-empty game refuses an answer for another game, since
// mod ids are global across Paradox's games.
func (c *Client) Mod(ctx context.Context, game, id, os string) (Mod, error) {
	if !ValidID(id) {
		return Mod{}, fmt.Errorf("paradoxmods: %q is not a mod id", id)
	}
	var out struct {
		answer
		Detail Mod `json:"modDetail"`
	}
	params := url.Values{"modId": {id}, "os": {os}}
	if err := c.do(ctx, http.MethodGet, "/mods", params, nil, &out); err != nil {
		return Mod{}, err
	}
	if err := out.err(); err != nil {
		return Mod{}, err
	}
	if game != "" && out.Detail.Game != game {
		return Mod{}, fmt.Errorf("paradoxmods: mod %s is a %q mod, not %q", id, out.Detail.Game, game)
	}
	return out.Detail, nil
}

// Versions lists the mod's versions, newest first.
func (c *Client) Versions(ctx context.Context, id string) ([]Version, error) {
	if !ValidID(id) {
		return nil, fmt.Errorf("paradoxmods: %q is not a mod id", id)
	}
	var out struct {
		answer
		ModVersions []Version `json:"modVersions"`
	}
	if err := c.do(ctx, http.MethodGet, "/mods/versions", url.Values{"modId": {id}}, nil, &out); err != nil {
		return nil, err
	}
	if err := out.err(); err != nil {
		return nil, err
	}
	slices.SortFunc(out.ModVersions, func(a, b Version) int { return cmp.Compare(b.ID, a.ID) })
	return out.ModVersions, nil
}

// do sends one request and decodes the answer into out. A non-2xx status is a *httpkit.StatusError, a transport failure
// a redacted error, and a body over maxBody a safefs.ErrTooLarge.
func (c *Client) do(ctx context.Context, method, path string, params url.Values, body, out any) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	u := c.base + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	var payload io.Reader = http.NoBody
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("paradoxmods: %w", err)
		}
		payload = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, payload)
	if err != nil {
		return fmt.Errorf("paradoxmods: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.agent != "" {
		req.Header.Set("User-Agent", c.agent)
	}
	raw, err := httpkit.DoCapped(c.hc, req, maxBody, nil)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("paradoxmods: %w", err)
	}
	return nil
}
