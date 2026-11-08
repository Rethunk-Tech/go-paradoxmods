package paradoxmods

import (
	"cmp"
	"regexp"
	"strconv"
)

// modIDPattern is a mod id: the API's ids are short decimal numbers.
var modIDPattern = regexp.MustCompile(`^[0-9]{1,9}$`)

// Sort is one order a listing can take.
type Sort string

// The listing's orders. Their meaning is the API's; the zero value asks for none and keeps the API's own order.
const (
	// SortBest is the API's relevance order, which needs no search text to be meaningful.
	SortBest Sort = "best"
	// SortDownloads is most installed first.
	SortDownloads Sort = "downloads"
	// SortRating is highest rated first.
	SortRating Sort = "rating"
	// SortUpdated is most recently changed first.
	SortUpdated Sort = "updated"
	// SortNewest is most recently published first.
	SortNewest Sort = "newest"
	// SortName is by display name, A first.
	SortName Sort = "name"
)

// sortFields is the mapping data: each order as the two body fields the API takes. It sorts by name ascending and by
// everything else descending.
var sortFields = map[Sort]struct{ by, order string }{
	SortBest:      {"best", "desc"},
	SortDownloads: {"popularity", "desc"},
	SortRating:    {"rating", "desc"},
	SortUpdated:   {"updated", "desc"},
	SortNewest:    {"created", "desc"},
	SortName:      {"displayName", "asc"},
}

// fields is the order as the API names it; an unknown value asks for none.
func (s Sort) fields() (by, order string) {
	f := sortFields[s]
	return f.by, f.order
}

// ModList is one page of a listing.
type ModList struct {
	// Total is how many mods the whole query matches, not just this page.
	Total int
	// Mods is the page's mods in the API's order.
	Mods []Mod
}

// Mod is one mod as a listing or a details answer gives it; a listing leaves the details fields empty.
type Mod struct {
	// Game is the API's name for the game the mod belongs to.
	Game string `json:"game"`
	// ID is the mod's numeric id, as a string.
	ID string `json:"modId"`
	// Name is the opaque file name the download address carries, not a display name.
	Name string `json:"name"`
	// DisplayName is the title the author gave it.
	DisplayName string `json:"displayName"`
	// Author is the author's account name.
	Author string `json:"author"`
	// ShortDescription is the listing's one-line summary.
	ShortDescription string `json:"shortDescription"`
	// LongDescription is the page's full description: Markdown, often with HTML in it.
	LongDescription string `json:"longDescription"`
	// ModVersion is the newest version's number, as a string.
	ModVersion string `json:"modVersion"`
	// UserModVersion is the newest version's label as the author set it; empty when the author set none.
	UserModVersion string `json:"userModVersion"`
	// Tags are the mod's categories.
	Tags []string `json:"tags"`
	// Requirements are what the newest version needs: other mods, and the game's own downloadable content.
	Requirements []Requirement `json:"requirements"`
	// Changelog holds every version's notes, in no particular order.
	Changelog []ChangeEntry `json:"changelog"`
	// Image is the listing picture's address.
	Image string `json:"displayImagePath"`
	// InstalledCount is how many players have the mod installed.
	InstalledCount int `json:"installedCount"`
	// RatingsTotal is how many players rated it.
	RatingsTotal int `json:"ratingsTotal"`
	// Updated is when the mod last changed, as the API formats it.
	Updated string `json:"updated"`
}

// Required is the ids of the mods the newest version needs: its requirements that are mods with well-formed ids, minus
// the mod's own entry. The API files the game's downloadable content under the same field, under another product type,
// so it is left out here.
func (m Mod) Required() []string {
	out := []string{}
	for _, r := range m.Requirements {
		if r.ProductType == "MOD" && ValidID(r.ID) && r.ID != m.ID {
			out = append(out, r.ID)
		}
	}
	return out
}

// Requirement is one product a version needs.
type Requirement struct {
	// ProductType is "MOD" for another mod, and the game's own kind for downloadable content.
	ProductType string `json:"productType"`
	// ID is the required product's id, opaque unless ProductType is "MOD".
	ID string `json:"id"`
}

// ChangeEntry is one version's release notes, as a mod's changelog holds them.
type ChangeEntry struct {
	// ModVersion is the version number the notes belong to.
	ModVersion int `json:"modVersion"`
	// UserVersion is the author's own label for that version; empty when the author set none.
	UserVersion string `json:"userModVersion"`
	// Released is when the version was published, "YYYY-MM-DD HH:MM:SS", the date first.
	Released string `json:"released"`
	// Details is the notes: Markdown, sometimes HTML.
	Details string `json:"details"`
}

// Version is one published version of a mod.
type Version struct {
	// ID is the version number, and the number the download address carries.
	ID int `json:"id"`
	// Name is the author's label for the version; empty when the author set none.
	Name string `json:"userVersion"`
	// Created is when it was published, as the API formats it.
	Created string `json:"created"`
	// Size is the archive's size in bytes.
	Size int64 `json:"size"`
}

// Label is the version's name, or its number when the author set none.
func (v Version) Label() string { return cmp.Or(v.Name, strconv.Itoa(v.ID)) }
