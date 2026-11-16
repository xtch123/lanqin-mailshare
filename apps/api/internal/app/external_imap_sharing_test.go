package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExternalIMAPShareTokenAndTimeWindows(t *testing.T) {
	a, digestA, err := newExternalIMAPShareToken()
	if err != nil {
		t.Fatal(err)
	}
	b, digestB, err := newExternalIMAPShareToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 43 || len(digestA) != 64 || a == b || digestA == digestB {
		t.Fatalf("unexpected token shape or collision: %d, %d", len(a), len(digestA))
	}
	now := time.Now().UTC()
	future := now.Add(4 * time.Hour).Format(time.RFC3339Nano)
	if _, err := validateExternalIMAPShareWindow(1440, &future, now); err != nil {
		t.Fatal(err)
	}
	past := now.Add(-time.Minute).Format(time.RFC3339Nano)
	if _, err := validateExternalIMAPShareWindow(1440, &past, now); err == nil {
		t.Fatal("past expiresAt must fail")
	}
	if _, err := validateExternalIMAPShareWindow(-1, nil, now); err == nil {
		t.Fatal("negative visibility must fail")
	}
	if _, err := validateExternalIMAPShareWindow(maxShareVisibilityMinutes+1, nil, now); err == nil {
		t.Fatal("excessive visibility must fail")
	}
}

func TestExternalIMAPShareIsolatedReadAndRevoke(t *testing.T) {
	a := newTestApp(t)
	user, mb := defaultAdminUserAndMailbox(t, a)
	now := time.Now().UTC()
	stamp := now.Format(time.RFC3339Nano)
	ctx := context.Background()
	for _, accountID := range []string{"ext-one", "ext-two"} {
		_, err := a.db.ExecContext(ctx, `INSERT INTO external_imap_accounts
			(id,user_id,mailbox_id,name,host,port,tls_mode,username,password_ciphertext,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			accountID, user.ID, mb.ID, accountID, "imap.example.test", 993, "tls",
			accountID+"@example.test", "encrypted-for-test", stamp, stamp)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i, accountID := range []string{"ext-one", "ext-two"} {
		messageID := "shared-message-" + accountID
		_, err := a.db.ExecContext(ctx, `INSERT INTO messages
			(id,mailbox_id,message_uid,message_id,subject,from_addr,to_addrs,
			 sent_at,received_at,snippet,body_text,body_html,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			messageID, mb.ID, messageID, messageID, "Mail "+accountID,
			"sender@example.test", "[]", stamp, stamp, "test", "hello",
			`<a href="https://www.netflix.com/">Get Code</a><script>alert(1)</script>`, stamp, stamp)
		if err != nil {
			t.Fatal(err)
		}
		_, err = a.db.ExecContext(ctx, `INSERT INTO external_imap_messages
			(account_id,remote_folder,uid_validity,uid,local_message_id,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?)`,
			accountID, "INBOX", 1, i+1, messageID, stamp, stamp)
		if err != nil {
			t.Fatal(err)
		}
	}
	token, digest, err := newExternalIMAPShareToken()
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.db.ExecContext(ctx, `INSERT INTO external_imap_share_links
		(id,account_id,owner_user_id,label,token_sha256,visibility_minutes,created_at)
		VALUES(?,?,?,?,?,?,?)`,
		"shared-link", "ext-one", user.ID, "one", digest, 1440, stamp)
	if err != nil {
		t.Fatal(err)
	}
	handler := a.Router()
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	list := get("/api/public/shared-inbox/messages")
	if list.Code != http.StatusOK {
		t.Fatalf("list status %d body %s", list.Code, list.Body.String())
	}
	var result struct {
		Items []externalShareMessage `json:"items"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "shared-message-ext-one" {
		t.Fatalf("cross-account leak: %+v", result.Items)
	}
	other := get("/api/public/shared-inbox/messages/shared-message-ext-two")
	if other.Code != http.StatusNotFound {
		t.Fatalf("other account message returned %d", other.Code)
	}
	detail := get("/api/public/shared-inbox/messages/shared-message-ext-one")
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status %d: %s", detail.Code, detail.Body.String())
	}
	if strings.Contains(detail.Body.String(), "<script") || !strings.Contains(detail.Body.String(), "https://www.netflix.com/") {
		t.Fatalf("HTML sanitizer failed or lost verification link: %s", detail.Body.String())
	}
	_, err = a.db.ExecContext(ctx, "UPDATE external_imap_share_links SET revoked_at=? WHERE id=?", stamp, "shared-link")
	if err != nil {
		t.Fatal(err)
	}
	if got := get("/api/public/shared-inbox/messages").Code; got != http.StatusUnauthorized {
		t.Fatalf("revoked share status = %d", got)
	}
	past := now.Add(-time.Minute).Format(time.RFC3339Nano)
	_, err = a.db.ExecContext(ctx, "UPDATE external_imap_share_links SET revoked_at=NULL,expires_at=? WHERE id=?", past, "shared-link")
	if err != nil {
		t.Fatal(err)
	}
	if got := get("/api/public/shared-inbox/messages").Code; got != http.StatusUnauthorized {
		t.Fatalf("expired share status = %d", got)
	}
	_, err = a.db.ExecContext(ctx, "UPDATE external_imap_share_links SET expires_at=NULL WHERE id=?", "shared-link")
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.db.ExecContext(ctx, "UPDATE external_imap_accounts SET enabled=0 WHERE id=?", "ext-one")
	if err != nil {
		t.Fatal(err)
	}
	if got := get("/api/public/shared-inbox/messages").Code; got != http.StatusUnauthorized {
		t.Fatalf("disabled external IMAP account status = %d", got)
	}
}

func TestExternalIMAPSharePagesRespectCSP(t *testing.T) {
	a := newTestApp(t)
	router := a.Router()
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	page := get("/api/public/shared-inbox")
	if page.Code != http.StatusOK {
		t.Fatalf("viewer page status: %d", page.Code)
	}
	if strings.Contains(page.Body.String(), "<script>") ||
		!strings.Contains(page.Body.String(), "<script defer src=\"/api/public/shared-inbox/app.js\"></script>") {
		t.Fatal("viewer must use an external same-origin JS file, never inline scripts")
	}
	if !strings.Contains(page.Header().Get("Content-Security-Policy"), "script-src 'self'") {
		t.Fatal("CSP must permit same-origin scripts without unsafe-inline")
	}
	js := get("/api/public/shared-inbox/app.js")
	if js.Code != http.StatusOK || !strings.HasPrefix(js.Header().Get("Content-Type"), "application/javascript") {
		t.Fatalf("viewer JS route failed: code=%d, type=%s", js.Code, js.Header().Get("Content-Type"))
	}
	if strings.Contains(js.Body.String(), "<script>") || !strings.Contains(js.Body.String(), "async function refresh()") {
		t.Fatal("viewer JS payload missing or incorrectly wrapped")
	}
	admin := get("/api/me/external-imap-share-panel/app.js")
	if admin.Code != http.StatusUnauthorized {
		t.Fatalf("admin script exposed without authentication: %d", admin.Code)
	}
}

func TestExternalIMAPShareMobileLayouts(t *testing.T) {
	a := newTestApp(t)
	viewer := httptest.NewRecorder()
	a.handleExternalIMAPSharePage(viewer, httptest.NewRequest(http.MethodGet, "/api/public/shared-inbox", nil))
	html := viewer.Body.String()
	for _, required := range []string{
		`name="viewport" content="width=device-width,initial-scale=1"`,
		`@media(max-width:720px)`,
		`class="mobile-back" id="backToList"`,
		`mail-shell.detail-open`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("viewer is missing mobile requirement %q", required)
		}
	}
	js := httptest.NewRecorder()
	a.handleExternalIMAPShareViewerJS(js, httptest.NewRequest(http.MethodGet, "/api/public/shared-inbox/app.js", nil))
	for _, required := range []string{`matchMedia`, `backToList`, `detail-open`} {
		if !strings.Contains(js.Body.String(), required) {
			t.Fatalf("viewer JS is missing mobile requirement %q", required)
		}
	}
	admin := httptest.NewRecorder()
	a.handleExternalIMAPShareAdminPage(admin, httptest.NewRequest(http.MethodGet, "/api/me/external-imap-share-panel", nil))
	for _, required := range []string{`@media(max-width:580px)`, `id="customDaysWrap"`, `id="customVisibleWrap"`} {
		if !strings.Contains(admin.Body.String(), required) {
			t.Fatalf("share manager is missing mobile requirement %q", required)
		}
	}
}
