package paradoxmods

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

type hostileAnswer struct {
	code int
	body []byte
}

// FuzzHostileAnswer holds that no status or body makes a call panic. One server serves the whole run: a server per
// input runs the machine out of ports.
func FuzzHostileAnswer(f *testing.F) {
	f.Add(200, []byte(searchAnswer))
	f.Add(200, []byte(`{"result":"OK","mod":{"requiredMods":[{}],"thumbnailUrl":"javascript:x"}}`))
	f.Add(200, []byte(strings.Repeat(`{"mods":[`, 20000)))
	f.Add(200, []byte(`{"result":"Error","errorCode":1}`))
	f.Add(502, []byte(nil))
	var a atomic.Pointer[hostileAnswer]
	a.Store(&hostileAnswer{code: 200})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := a.Load()
		w.WriteHeader(cur.code)
		_, _ = w.Write(cur.body)
	}))
	f.Cleanup(srv.Close)
	c := New(Config{BaseURL: srv.URL})
	f.Fuzz(func(t *testing.T, code int, body []byte) {
		if code < 200 || code > 599 || code/100 == 3 {
			code = 200
		}
		a.Store(&hostileAnswer{code: code, body: body})
		ctx := t.Context()
		if l, err := c.Search(ctx, SearchQuery{Game: "g"}); err == nil {
			for _, m := range l.Mods {
				_ = m.Required()
			}
		}
		_, _ = c.Mod(ctx, "g", "74604", "Windows")
		if vs, err := c.Versions(ctx, "74604"); err == nil {
			for _, v := range vs {
				_ = v.Label()
			}
		}
	})
}

func TestALyingContentLengthFails(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, rw, err := http.NewResponseController(w).Hijack()
		if err != nil {
			t.Error(err)
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = rw.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 1099511627776\r\n\r\n{\"result\":\"OK\"}")
		_ = rw.Flush()
	}))
	t.Cleanup(srv.Close)
	if _, err := New(Config{BaseURL: srv.URL}).Search(t.Context(), SearchQuery{Game: "g"}); err == nil {
		t.Fatal("a body cut short of its Content-Length was accepted")
	}
}
