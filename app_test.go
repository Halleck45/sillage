package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestWithoutFlag(t *testing.T) {
	in := []string{"sillage", "-app", "-addr", "127.0.0.1:9000", "--app=true", "-apps", "-data", "/x"}
	want := []string{"sillage", "-addr", "127.0.0.1:9000", "-apps", "-data", "/x"}
	if got := withoutFlag(in, "app"); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if in[1] != "-app" {
		t.Fatal("input slice must not be modified")
	}
}

func TestAppURL(t *testing.T) {
	cases := map[string]string{
		"127.0.0.1:8787": "http://127.0.0.1:8787/",
		":8787":          "http://127.0.0.1:8787/",
		"0.0.0.0:9000":   "http://127.0.0.1:9000/",
		"[::]:9000":      "http://127.0.0.1:9000/",
		"[::1]:9000":     "http://[::1]:9000/",
		"100.64.0.2:80":  "http://100.64.0.2:80/",
	}
	for addr, want := range cases {
		if got := appURL(addr); got != want {
			t.Errorf("appURL(%q) = %q, want %q", addr, got, want)
		}
	}
}

func TestSillageServing(t *testing.T) {
	sillage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<!doctype html><head><title>Sillage</title></head>"))
	}))
	defer sillage.Close()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<title>Grafana</title>"))
	}))
	defer other.Close()

	if !sillageServing(sillage.URL + "/") {
		t.Error("a Sillage instance must be detected")
	}
	if sillageServing(other.URL + "/") {
		t.Error("another program on the port must not pass for Sillage")
	}
	closed := httptest.NewServer(http.NotFoundHandler())
	url := closed.URL + "/"
	closed.Close()
	if sillageServing(url) {
		t.Error("a closed port must not pass for Sillage")
	}
}

func TestDesktopQuote(t *testing.T) {
	cases := map[string]string{
		"/usr/bin/sillage": "/usr/bin/sillage",
		"/home/me/My Data": `"/home/me/My Data"`,
		"/tmp/100%":        "/tmp/100%%",
		`/a"b`:             `"/a\\"b"`,
		"/cost$HOME":       `"/cost\\$HOME"`,
		`/back\slash dir`:  `"/back\\\\slash dir"`,
		"127.0.0.1:9000":   "127.0.0.1:9000",
	}
	for in, want := range cases {
		if got := desktopQuote(in); got != want {
			t.Errorf("desktopQuote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestInstallDesktopEntry(t *testing.T) {
	if _, err := os.Stat("/proc/self/exe"); err != nil {
		t.Skip("Linux only")
	}
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)
	web := fstest.MapFS{"favicon.svg": {Data: []byte("<svg/>")}}

	path, err := installDesktopEntry(web, []string{"-data", "/home/me/My Data"})
	if err != nil {
		t.Fatal(err)
	}
	entry, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(entry)
	for _, want := range []string{"[Desktop Entry]", "Icon=sillage", " -app -data \"/home/me/My Data\"\n", "Terminal=false"} {
		if !strings.Contains(text, want) {
			t.Errorf("desktop entry lacks %q:\n%s", want, text)
		}
	}
	if _, err := os.Stat(filepath.Join(dataHome, "icons", "hicolor", "scalable", "apps", "sillage.svg")); err != nil {
		t.Errorf("icon not installed: %v", err)
	}
}

func TestInstanceLock(t *testing.T) {
	dir := t.TempDir()
	first, err := acquireInstanceLock(dir)
	if err != nil {
		t.Fatal(err)
	}
	// flock porte sur la description de fichier : une seconde ouverture, même
	// depuis le même process, se heurte au verrou comme le ferait un autre binaire.
	if _, err := acquireInstanceLock(dir); !errors.Is(err, errInstanceRunning) {
		t.Fatalf("second lock: got %v, want errInstanceRunning", err)
	}
	if _, err := readInstance(dir, 0); err == nil {
		t.Fatal("an unpublished instance must not yield an address")
	}
	if err := publishInstance(first, instanceInfo{PID: 42, Addr: "127.0.0.1:8788", Version: "dev"}); err != nil {
		t.Fatal(err)
	}
	info, err := readInstance(dir, 0)
	if err != nil || info.Addr != "127.0.0.1:8788" || info.PID != 42 {
		t.Fatalf("readInstance = %+v, %v", info, err)
	}
	// Une adresse plus courte ne garde aucun reste de la précédente.
	if err := publishInstance(first, instanceInfo{PID: 7, Addr: ":1"}); err != nil {
		t.Fatal(err)
	}
	if info, _ := readInstance(dir, 0); info.Addr != ":1" || info.Version != "" {
		t.Fatalf("stale content after republish: %+v", info)
	}
	first.Close()
	again, err := acquireInstanceLock(dir)
	if err != nil {
		t.Fatalf("lock must be free once released: %v", err)
	}
	again.Close()
}
