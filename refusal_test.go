package paradoxmods

import "testing"

func TestModRefusesAnAnswerForAnotherOrNoMod(t *testing.T) {
	t.Parallel()
	for name, answer := range map[string]string{
		"another": detailAnswer,
		"none":    `{"result":"OK"}`,
	} {
		if _, err := serve(t, answer, nil).Mod(t.Context(), "", "99999", "Windows"); err == nil {
			t.Fatalf("%s: answer for the wrong mod was accepted", name)
		}
	}
}

func TestImageKeepsOnlyWebAddresses(t *testing.T) {
	t.Parallel()
	answer := `{"result":"OK","modDetail":{"modId":"1","displayImagePath":"javascript:alert(1)"}}`
	m, err := serve(t, answer, nil).Mod(t.Context(), "", "1", "Windows")
	if err != nil || m.Image != "" {
		t.Fatalf("image = %q, err = %v", m.Image, err)
	}
	list, err := serve(t, `{"result":"OK","mods":[{"displayImagePath":"file:///etc/passwd"}]}`, nil).Search(t.Context(), SearchQuery{Game: "g"})
	if err != nil || list.Mods[0].Image != "" {
		t.Fatalf("listing image = %q, err = %v", list.Mods[0].Image, err)
	}
}

func TestDownloadURLRefusesSpacesAndControlCharacters(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"a b", "a\nb", "a\tb", "a\x7fb"} {
		if got := DownloadURL("g", name, 1); got != "" {
			t.Fatalf("DownloadURL(%q) = %q", name, got)
		}
	}
}
