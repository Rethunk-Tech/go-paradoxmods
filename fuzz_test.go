package paradoxmods

import (
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"testing"
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
		got := m.Required()
		if got == nil {
			t.Fatal("Required() = nil, want empty slice")
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
