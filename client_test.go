package paradoxmods

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/Rethunk-Tech/go-httpkit"
)

const (
	// detailAnswer is a details answer as the API sends it.
	detailAnswer = `{"result":"OK","modDetail":{"author":"yenyang","displayName":"Anarchy","game":"cities_skylines_2",` +
		`"longDescription":"Anarchy disables error checks.","name":"25faeda9-b17e-4525-a08f-ba219ed13524","modId":"74604",` +
		`"modVersion":"40","userModVersion":"1.8.0","tags":["Code Mod"],"displayImagePath":"https://modscontent.test/cover.jpg",` +
		`"requirements":[{"productType":"MOD","id":"74417","name":"Unified Icon Library","version":40},` +
		`{"productType":"DLC","id":"1","name":"Beach Properties","version":1},` +
		`{"productType":"MOD","id":"74604","name":"itself","version":1},` +
		`{"productType":"MOD","id":"../x","name":"climbs out","version":1}],` +
		`"changelog":[{"modVersion":38,"released":"2026-08-01 10:00:00","details":"## Patch V1.7.9","userModVersion":"1.7.9"},` +
		`{"modVersion":40,"released":"2026-09-15 03:23:57","details":"<p>Patch <b>1.8.0</b></p>","userModVersion":"1.8.0"}]}}`

	searchAnswer = `{"totalCount":27068,"result":"OK","mods":[{"author":"yenyang","displayName":"Anarchy",` +
		`"game":"cities_skylines_2","installedCount":1047436,"modVersion":"40","modId":"74604",` +
		`"name":"25faeda9-b17e-4525-a08f-ba219ed13524","os":"Windows","ratingsTotal":14464,` +
		`"shortDescription":"Disables Error Checks and does not show errors.","tags":["Code Mod"],` +
		`"updated":"2026-09-15T19:59:20.000Z","displayImagePath":"https://modscontent.test/cover.jpg"}]}`

	versionsAnswer = `{"modVersions":[{"created":"2024-02-27T18:43:45.000+0000","size":901017,"userVersion":"1.4.0","id":1},` +
		`{"created":"2026-08-01T10:00:00.000+0000","size":500,"userVersion":"","id":38},` +
		`{"created":"2026-09-15T03:23:57.000+0000","size":670237,"userVersion":"1.8.0","id":40}],"result":"OK"}`

	refusedAnswer = `{"result":"Failure","errorCode":"bad-input",` +
		`"errorMessage":"Mod is corrupted or chosen operating system is not supported","detail":""}`
)

// serve answers every request with answer, after letting check inspect it.
func serve(t *testing.T, answer string, check func(*testing.T, *http.Request)) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(t, r)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(srv.Close)
	return New(Config{BaseURL: srv.URL + "/", UserAgent: "test/1.0"})
}

func TestSearchSendsTheListingAndDecodesIt(t *testing.T) {
	t.Parallel()
	var sent []byte
	c := serve(t, searchAnswer, func(t *testing.T, r *http.Request) {
		t.Helper()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		sent = raw
		if r.Method != http.MethodPost || r.RequestURI != "/mods" {
			t.Errorf("%s %s", r.Method, r.RequestURI)
		}
		if got := r.Header.Get("User-Agent"); got != "test/1.0" {
			t.Errorf("User-Agent = %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
	})
	list, err := c.Search(t.Context(), SearchQuery{
		Game: "cities_skylines_2", Text: "road", Page: 2, Size: 50, Tags: []string{"Map"}, Sort: SortName,
	})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"gameName":"cities_skylines_2","page":2,"pageSize":50,"searchQuery":"road","tags":["Map"],` +
		`"sortBy":"displayName","orderBy":"asc"}`
	if string(sent) != want {
		t.Fatalf("body = %s", sent)
	}
	if list.Total != 27068 || len(list.Mods) != 1 {
		t.Fatalf("list = %+v", list)
	}
	m := list.Mods[0]
	if m.ID != "74604" || m.DisplayName != "Anarchy" || m.InstalledCount != 1047436 || m.RatingsTotal != 14464 ||
		m.ShortDescription == "" || m.Updated != "2026-09-15T19:59:20.000Z" || m.Image != "https://modscontent.test/cover.jpg" {
		t.Fatalf("mod = %+v", m)
	}
}

func TestSearchAsksForAPageWithinTheCap(t *testing.T) {
	t.Parallel()
	var sent []byte
	c := serve(t, searchAnswer, func(t *testing.T, r *http.Request) {
		t.Helper()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		sent = raw
	})
	if _, err := c.Search(t.Context(), SearchQuery{Game: "g", Size: 5000}); err != nil {
		t.Fatal(err)
	}
	if want := `{"gameName":"g","page":1,"pageSize":100}`; string(sent) != want {
		t.Fatalf("body = %s", sent)
	}
}

func TestModReadsTheDetailAndItsRequiredMods(t *testing.T) {
	t.Parallel()
	c := serve(t, detailAnswer, func(t *testing.T, r *http.Request) {
		t.Helper()
		if r.Method != http.MethodGet || r.URL.Path != "/mods" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query(); got.Get("modId") != "74604" || got.Get("os") != "Windows" {
			t.Errorf("query %v", got)
		}
	})
	m, err := c.Mod(t.Context(), "cities_skylines_2", "74604", "Windows")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != "74604" || m.DisplayName != "Anarchy" || m.Name != "25faeda9-b17e-4525-a08f-ba219ed13524" ||
		m.ModVersion != "40" || m.UserModVersion != "1.8.0" || len(m.Changelog) != 2 || m.LongDescription == "" {
		t.Fatalf("mod = %+v", m)
	}
	// The downloadable content, the mod's own entry and the id that would climb out of an address are not mods to fetch.
	if got := m.Required(); !slices.Equal(got, []string{"74417"}) {
		t.Fatalf("required = %v", got)
	}
	if _, err := c.Mod(t.Context(), "cities_skylines", "74604", "Windows"); err == nil {
		t.Fatal("another game's mod was accepted")
	}
	for _, id := range []string{"../x", "", "1234567890"} {
		if _, err := c.Mod(t.Context(), "", id, "Windows"); err == nil {
			t.Fatalf("Mod(%q) accepted an id that is no mod id", id)
		}
	}
}

func TestVersionsListsNewestFirstAndLabelsUnnamedOnesByNumber(t *testing.T) {
	t.Parallel()
	c := serve(t, versionsAnswer, func(t *testing.T, r *http.Request) {
		t.Helper()
		if r.URL.Path != "/mods/versions" || r.URL.Query().Get("modId") != "74604" {
			t.Errorf("%s %v", r.URL.Path, r.URL.Query())
		}
	})
	got, err := c.Versions(t.Context(), "74604")
	if err != nil {
		t.Fatal(err)
	}
	labels := make([]string, len(got))
	for i, v := range got {
		labels[i] = v.Label()
	}
	if !slices.Equal(labels, []string{"1.8.0", "38", "1.4.0"}) || got[0].Size != 670237 || got[0].Created == "" {
		t.Fatalf("versions = %+v", got)
	}
	if _, err := c.Versions(t.Context(), "../x"); err == nil {
		t.Fatal("a versions call for an id that is no mod id ran")
	}
}

func TestARefusalSentWithA200IsAnAPIError(t *testing.T) {
	t.Parallel()
	_, err := serve(t, refusedAnswer, nil).Search(t.Context(), SearchQuery{Game: "cities_skylines_2"})
	var refused *APIError
	if !errors.As(err, &refused) || refused.Code != "bad-input" || !strings.Contains(refused.Error(), "not supported") {
		t.Fatalf("refusal = %v", err)
	}
}

func TestAStatusTheAPIDidNotAnswerWithIsAStatusError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	_, err := New(Config{BaseURL: srv.URL}).Versions(t.Context(), "1")
	var status *httpkit.StatusError
	if !errors.As(err, &status) || status.Code != http.StatusBadGateway {
		t.Fatalf("status = %v", err)
	}
}

func TestAnAnswerOverTheCapFailsInsteadOfGrowing(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("a"), maxBody+1))
	}))
	t.Cleanup(srv.Close)
	if list, err := New(Config{BaseURL: srv.URL}).Search(t.Context(), SearchQuery{Game: "g"}); err == nil {
		t.Fatalf("an oversized answer decoded: %+v", list)
	}
}

func TestARedirectOffOriginIsRefused(t *testing.T) {
	t.Parallel()
	elsewhere := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, searchAnswer)
	}))
	t.Cleanup(elsewhere.Close)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL+"/mods", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{BaseURL: srv.URL}).Search(t.Context(), SearchQuery{Game: "g"}); err == nil {
		t.Fatal("a redirect to another origin was followed")
	}
}

func TestAddresses(t *testing.T) {
	t.Parallel()
	if PageURL("74604") != "https://mods.paradoxplaza.com/mods/74604/Windows" || PageURL("../x") != "" {
		t.Fatalf("PageURL = %q, %q", PageURL("74604"), PageURL("../x"))
	}
	if !ValidID("74604") || ValidID(" 74604") || ValidID("") || ValidID("1234567890") {
		t.Fatal("ValidID accepts ids that are no mod id")
	}
	const name = "25faeda9-b17e-4525-a08f-ba219ed13524"
	want := "https://modscontent.paradox-interactive.com/cities_skylines_2/" + name + "/content/sources/" + name + "_40.zip"
	if got := DownloadURL("cities_skylines_2", name, 40); got != want {
		t.Fatalf("DownloadURL = %q", got)
	}
}

func TestSortsMapToTheAPIFields(t *testing.T) {
	t.Parallel()
	for sort, want := range map[Sort]string{
		SortBest: "best desc", SortDownloads: "popularity desc", SortRating: "rating desc",
		SortUpdated: "updated desc", SortNewest: "created desc", SortName: "displayName asc", Sort(""): " ",
	} {
		by, order := sort.fields()
		if got := by + " " + order; got != want {
			t.Errorf("%q maps to %q, want %q", sort, got, want)
		}
	}
}
