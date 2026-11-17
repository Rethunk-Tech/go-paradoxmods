package paradoxmods

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

func FuzzValidID(f *testing.F) {
	for _, id := range []string{
		"", "1", "0", "123456789", "1234567890", "12a", " 1", "1 ", "-1", "000000001",
	} {
		f.Add(id)
	}
	f.Fuzz(func(t *testing.T, id string) {
		asciiDigits := len(id) >= 1 && len(id) <= 9
		if asciiDigits {
			for i := 0; i < len(id); i++ {
				if id[i] < '0' || id[i] > '9' {
					asciiDigits = false
					break
				}
			}
		}
		got := ValidID(id)
		if got != asciiDigits {
			t.Fatalf("ValidID(%q) = %v, want %v", id, got, asciiDigits)
		}
		page := PageURL(id)
		if !got {
			if page != "" {
				t.Fatalf("PageURL(%q) = %q, want empty", id, page)
			}
			return
		}
		want := siteURL + "/mods/" + id + "/" + pageOS
		if page != want {
			t.Fatalf("PageURL(%q) = %q, want %q", id, page, want)
		}
		u, err := url.Parse(page)
		if err != nil {
			t.Fatalf("PageURL(%q) is not a URL: %v", id, err)
		}
		if u.Path != "/mods/"+id+"/"+pageOS {
			t.Fatalf("PageURL(%q) path = %q", id, u.Path)
		}
	})
}

func FuzzAnswer(f *testing.F) {
	f.Add([]byte(`{"result":"OK"}`))
	f.Add([]byte(`{"result":"Error","errorCode":"bad-input","errorMessage":"invalid"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"result":"OK","errorCode":"x","errorMessage":"y"}`))
	f.Add([]byte(`{"result":"","errorCode":"","errorMessage":""}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		var a answer
		if err := json.Unmarshal(raw, &a); err != nil {
			return
		}
		e := a.err()
		if a.Result == "OK" {
			if e != nil {
				t.Fatalf("OK answer: err() = %v", e)
			}
			return
		}
		var ae *APIError
		if !errors.As(e, &ae) {
			t.Fatalf("non-OK: err() is %T %v, want *APIError", e, e)
		}
		if ae.Code != a.ErrorCode || ae.Message != a.ErrorMessage {
			t.Fatalf("APIError %+v does not match answer %+v", ae, a)
		}
	})
}

func FuzzModRequired(f *testing.F) {
	f.Add([]byte(`{"modId":"1","requirements":[]}`))
	f.Add([]byte(`{"modId":"1","requirements":[{"productType":"MOD","id":"2"}]}`))
	f.Add([]byte(`{"modId":"1","requirements":[{"productType":"MOD","id":"1"}]}`))
	f.Add([]byte(`{"modId":"1","requirements":[{"productType":"DLC","id":"2"}]}`))
	f.Add([]byte(`{"modId":"123456789","requirements":[{"productType":"MOD","id":"0"},{"productType":"MOD","id":"not-an-id"}]}`))
	f.Fuzz(func(t *testing.T, raw []byte) {
		var m Mod
		if err := json.Unmarshal(raw, &m); err != nil {
			return
		}
		enc, err := json.Marshal(m)
		if err != nil {
			t.Fatal(err)
		}
		var m2 Mod
		if err := json.Unmarshal(enc, &m2); err != nil {
			t.Fatal(err)
		}
		enc2, err := json.Marshal(m2)
		if err != nil {
			t.Fatal(err)
		}
		if string(enc) != string(enc2) {
			t.Fatalf("round-trip encoding changed:\n%s\n%s", enc, enc2)
		}
		got := m.Required()
		if got == nil {
			t.Fatal("Required() = nil, want empty slice")
		}
		if len(got) > len(m.Requirements) {
			t.Fatalf("Required allocated %d from %d requirements", len(got), len(m.Requirements))
		}
		var want []string
		for _, r := range m.Requirements {
			if r.ProductType == "MOD" && ValidID(r.ID) && r.ID != m.ID {
				want = append(want, r.ID)
			}
		}
		if !slices.Equal(got, want) {
			t.Fatalf("Required() = %v, want %v", got, want)
		}
	})
}

func FuzzVersionJSON(f *testing.F) {
	f.Add([]byte(`{"id":1,"userVersion":"1.0","created":"2024-01-01","size":10}`))
	f.Add([]byte(`{"id":0}`))
	f.Add([]byte(`{"id":-1,"userVersion":""}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Fuzz(func(t *testing.T, raw []byte) {
		var v Version
		if err := json.Unmarshal(raw, &v); err != nil {
			return
		}
		enc, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		var v2 Version
		if err := json.Unmarshal(enc, &v2); err != nil {
			t.Fatal(err)
		}
		enc2, err := json.Marshal(v2)
		if err != nil {
			t.Fatal(err)
		}
		if string(enc) != string(enc2) {
			t.Fatalf("round-trip encoding changed:\n%s\n%s", enc, enc2)
		}
		got := v.Label()
		want := v.Name
		if want == "" {
			want = strconv.Itoa(v.ID)
		}
		if got != want {
			t.Fatalf("Label()=%q want %q", got, want)
		}
	})
}

func FuzzDownloadURL(f *testing.F) {
	const name = "25faeda9-b17e-4525-a08f-ba219ed13524"
	f.Add("cities_skylines_2", name, 40)
	f.Add("cities_skylines_2", "../x", 40)
	f.Add("..", name, 40)
	f.Add("g?x", name, 40)
	f.Add("g#x", name, 40)
	f.Add("cities_skylines_2", "a b", 1)
	f.Add("cities_skylines_2", "a\nb", 1)
	f.Add("cities_skylines_2", "a\tb", 1)
	f.Add("cities_skylines_2", "a\x7fb", 1)
	f.Add("cities_skylines_2", name, 0)
	f.Add("cities_skylines_2", name, -1)
	f.Add("", name, 1)
	f.Add("g", ".", 1)
	f.Add("g", "..", 1)
	f.Add("g", "x/y", 1)
	f.Add("g\\h", "n", 1)
	f.Add("g", "n%2e%2e", 1)
	f.Add("g", "", 1)
	f.Fuzz(func(t *testing.T, game, name string, n int) {
		got := DownloadURL(game, name, n)
		if n < 1 || !pathSegment(game) || !pathSegment(name) {
			if got != "" {
				t.Fatalf("DownloadURL(%q, %q, %d) = %q, want empty", game, name, n, got)
			}
			return
		}
		want := contentURL + "/" + game + "/" + name + "/content/sources/" + name + "_" + strconv.Itoa(n) + ".zip"
		if got != want {
			t.Fatalf("DownloadURL = %q, want %q", got, want)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatalf("not a URL: %v", err)
		}
		if u.Scheme != "https" || u.Host != "modscontent.paradox-interactive.com" {
			t.Fatalf("left the content host: %q", got)
		}
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if slices.Contains(parts, "..") || slices.Contains(parts, ".") {
			t.Fatalf("path escaped the root: %q", u.Path)
		}
		if len(parts) != 5 || parts[0] != game || parts[1] != name {
			t.Fatalf("path = %q", u.Path)
		}
	})
}

func FuzzWebImage(f *testing.F) {
	f.Add("https://modscontent.test/cover.jpg")
	f.Add("javascript:alert(1)")
	f.Add("file:///etc/passwd")
	f.Add("")
	f.Add("http://example.com/x")
	f.Add("//evil.com/x")
	f.Add("/relative")
	f.Add("https://")
	f.Add("data:text/html,x")
	f.Add("https://modscontent.test/cover.jpg?x=1#y")
	f.Fuzz(func(t *testing.T, raw string) {
		got := webImage(raw)
		if got == "" {
			u, err := url.Parse(raw)
			if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
				return
			}
			t.Fatalf("webImage(%q) = empty, want kept", raw)
		}
		if got != raw {
			t.Fatalf("webImage returned %q from %q", got, raw)
		}
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("kept unparseable %q: %v", raw, err)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			t.Fatalf("kept non-web scheme %q", raw)
		}
		if u.Host == "" {
			t.Fatalf("kept empty host %q", raw)
		}
	})
}

func FuzzPathSegment(f *testing.F) {
	f.Add("cities_skylines_2")
	f.Add("25faeda9-b17e-4525-a08f-ba219ed13524")
	f.Add("../x")
	f.Add("..")
	f.Add(".")
	f.Add("")
	f.Add("g?x")
	f.Add("a b")
	f.Add("a\nb")
	f.Add("n%2e%2e")
	f.Fuzz(func(t *testing.T, s string) {
		got := pathSegment(s)
		bad := s == "" || s == "." || s == ".." || strings.ContainsAny(s, "/\\?#% \x00") ||
			strings.IndexFunc(s, unicode.IsControl) >= 0
		if got != !bad {
			t.Fatalf("pathSegment(%q) = %v, want %v", s, got, !bad)
		}
	})
}
