package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalIMAPShareLiteOnlyStartsEssentialWorkers(t *testing.T) {
	dir := newTestDir(t)
	cfg := Config{
		Addr:              ":0",
		DBPath:            filepath.Join(dir, "lanqin.db"),
		DataDir:           filepath.Join(dir, "data"),
		CookieName:        "lanqin_lite_test",
		SessionTTLHours:   24,
		AdminEmail:        "admin@lanqin.local",
		AdminPassword:     "ChangeMe123!",
		PublicHostname:    "mail.example.test",
		PublicBaseURL:     "http://localhost:18090",
		AllowInsecureHTTP: true,
		ExternalIMAPEnabled: true,
		ExternalIMAPSecretKey: "lite-test-secret-only",
		MailShareLite:     true,
	}
	a, err := New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if !a.workersReady() {
		t.Fatal("lite must have operational background workers")
	}
	// With MaildirRoot unset, only external-IMAP sync and session cleanup
	// should remain; no outgoing mail or notification workers.
	if got := a.health.workersExpected.Load(); got != 2 {
		t.Fatalf("lite started %d workers, expected only sync and session cleanup", got)
	}
}

func TestExternalIMAPShareLiteEmbeddedAdminPage(t *testing.T) {
	a := newTestApp(t)
	rec := httptest.NewRecorder()
	a.handleExternalIMAPShareAdminPage(rec, httptest.NewRequest(http.MethodGet, "/api/me/external-imap-share-panel?embed=1", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "<body class=\"embedded\">") {
		t.Fatalf("lite embedded share management page unavailable: %d", rec.Code)
	}
	standalone := httptest.NewRecorder()
	a.handleExternalIMAPShareAdminPage(standalone, httptest.NewRequest(http.MethodGet, "/api/me/external-imap-share-panel", nil))
	if strings.Contains(standalone.Body.String(), "<body class=\"embedded\">") {
		t.Fatal("standard admin page must not be forcibly embedded")
	}
}

func TestExternalIMAPShareLiteConfigFlag(t *testing.T) {
	t.Setenv("LANQIN_MAILSHARE_LITE", "true")
	if !LoadConfig().MailShareLite {
		t.Fatal("lite env flag was not parsed")
	}
	t.Setenv("LANQIN_MAILSHARE_LITE", "false")
	if LoadConfig().MailShareLite {
		t.Fatal("full installation must remain the default")
	}
}
