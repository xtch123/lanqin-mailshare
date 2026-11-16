package app

import (
	"context"
	"database/sql"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExternalIMAPShareOwnerCanCopyFilterAndRotate(t *testing.T) {
	a := newTestApp(t)
	a.cfg.ExternalIMAPEnabled = true
	a.cfg.ExternalIMAPSecretKey = "test-share-encryption-key-do-not-use-in-production"
	user, mailbox := defaultAdminUserAndMailbox(t, a)
	stamp := time.Now().UTC().Format(time.RFC3339Nano)
	for _, id := range []string{"external-share-a", "external-share-b"} {
		_, err := a.db.ExecContext(context.Background(), `INSERT INTO external_imap_accounts
			(id,user_id,mailbox_id,name,host,port,tls_mode,username,password_ciphertext,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`,
			id, user.ID, mailbox.ID, "邮箱 "+id, "imap.example.test", 993, "tls",
			id+"@seek.li", "encrypted-for-test", stamp, stamp)
		if err != nil {
			t.Fatal(err)
		}
	}

	server := httptest.NewServer(a.Router())
	defer server.Close()
	admin := &testClient{t: t, server: server}
	if code := admin.do("POST", "/api/auth/login", map[string]string{
		"email": "admin@lanqin.local", "password": "ChangeMe123!",
	}, nil); code != http.StatusOK {
		t.Fatalf("login status %d", code)
	}

	var shareA struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if code := admin.do("POST", "/api/me/external-imap-shares",
		map[string]any{"accountId": "external-share-a", "label": "车友甲", "visibilityMinutes": 0}, &shareA); code != http.StatusCreated {
		t.Fatalf("create share A status %d", code)
	}
	if !strings.Contains(shareA.URL, "#") {
		t.Fatal("share URL lacks client-side fragment token")
	}
	tokenA := strings.SplitN(shareA.URL, "#", 2)[1]
	var shareB struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if code := admin.do("POST", "/api/me/external-imap-shares",
		map[string]any{"accountId": "external-share-b", "label": "车友乙", "visibilityMinutes": 1440}, &shareB); code != http.StatusCreated {
		t.Fatalf("create share B status %d", code)
	}
	var digest, encrypted string
	if err := a.db.QueryRow(`SELECT token_sha256,token_ciphertext
		FROM external_imap_share_links WHERE id=?`, shareA.ID).Scan(&digest, &encrypted); err != nil {
		t.Fatal(err)
	}
	wantDigest := sha256.Sum256([]byte(tokenA))
	if digest != hex.EncodeToString(wantDigest[:]) || encrypted == "" || strings.Contains(encrypted, tokenA) {
		t.Fatal("database must store the expected hash and encrypted token, never plaintext")
	}

	var listing struct {
		Items []externalIMAPShareLink `json:"items"`
	}
	if code := admin.do("GET", "/api/me/external-imap-shares?accountId=external-share-a", nil, &listing); code != http.StatusOK {
		t.Fatalf("filtered listing status %d", code)
	}
	if len(listing.Items) != 1 || listing.Items[0].AccountName != "邮箱 external-share-a" ||
		listing.Items[0].AccountEmail != "external-share-a@seek.li" || !listing.Items[0].CanCopy ||
		listing.Items[0].ID != shareA.ID {
		t.Fatalf("mailbox filtering or labels failed: %+v", listing.Items)
	}
	var copied struct{ URL string `json:"url"` }
	if code := admin.do("GET", "/api/me/external-imap-shares/"+shareA.ID+"/link", nil, &copied); code != http.StatusOK || copied.URL != shareA.URL {
		t.Fatalf("recopy failed: code=%d same=%v", code, copied.URL == shareA.URL)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/me/external-imap-shares/"+shareA.ID+"/link", nil)
	rec := httptest.NewRecorder()
	a.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous token recovery should fail with 401, got %d", rec.Code)
	}

	// A legacy record is valid for guests, but its original token is
	// irrecoverable. Rotate it to issue a new copyable link.
	legacyToken, legacyDigest, err := newExternalIMAPShareToken()
	if err != nil {
		t.Fatal(err)
	}
	_, err = a.db.Exec(`INSERT INTO external_imap_share_links
		(id,account_id,owner_user_id,label,token_sha256,visibility_minutes,created_at)
		VALUES(?,?,?,?,?,?,?)`,
		"legacy-share", "external-share-a", user.ID, "旧链接", legacyDigest, 0, stamp)
	if err != nil {
		t.Fatal(err)
	}
	if code := admin.do("GET", "/api/me/external-imap-shares/legacy-share/link", nil, nil); code != http.StatusConflict {
		t.Fatalf("legacy token recovery status %d, want 409", code)
	}
	var rotated struct{ URL string `json:"url"` }
	if code := admin.do("POST", "/api/me/external-imap-shares/legacy-share/rotate", nil, &rotated); code != http.StatusOK {
		t.Fatalf("rotate legacy share status %d", code)
	}
	if rotated.URL == "" || strings.Contains(rotated.URL, "#"+legacyToken) {
		t.Fatal("rotation should issue a new token")
	}
	if code := admin.do("GET", "/api/me/external-imap-shares/legacy-share/link", nil, &copied); code != http.StatusOK || copied.URL != rotated.URL {
		t.Fatal("rotated link must remain recoverable")
	}
	guestRequest := httptest.NewRequest(http.MethodGet, "/api/public/shared-inbox/messages", nil)
	guestRequest.Header.Set("Authorization", "Bearer "+legacyToken)
	guestResponse := httptest.NewRecorder()
	a.Router().ServeHTTP(guestResponse, guestRequest)
	if guestResponse.Code != http.StatusUnauthorized {
		t.Fatalf("rotated old guest token should be invalid, got %d", guestResponse.Code)
	}

	if code := admin.do("DELETE", "/api/me/external-imap-shares/"+shareA.ID, nil, nil); code != http.StatusOK {
		t.Fatalf("revoke status %d", code)
	}
	if code := admin.do("GET", "/api/me/external-imap-shares/"+shareA.ID+"/link", nil, nil); code != http.StatusNotFound {
		t.Fatalf("revoked token should not be copied: %d", code)
	}
}

func TestExternalIMAPShareLegacyTableMigrationPreservesRows(t *testing.T) {
	a := newTestApp(t)
	if _, err := a.db.Exec(`DROP TABLE external_imap_share_links`); err != nil {
		t.Fatal(err)
	}
	if _, err := a.db.Exec(`CREATE TABLE external_imap_share_links (
		id VARCHAR(64) PRIMARY KEY,
		account_id VARCHAR(64) NOT NULL,
		owner_user_id VARCHAR(64) NOT NULL,
		label VARCHAR(128) NOT NULL DEFAULT '',
		token_sha256 CHAR(64) NOT NULL UNIQUE,
		visibility_minutes BIGINT NOT NULL DEFAULT 1440,
		created_at VARCHAR(40) NOT NULL,
		expires_at VARCHAR(40),
		revoked_at VARCHAR(40),
		last_accessed_at VARCHAR(40)
	)`); err != nil {
		t.Fatal(err)
	}
	_, err := a.db.Exec(`INSERT INTO external_imap_share_links
		(id,account_id,owner_user_id,label,token_sha256,created_at)
		VALUES(?,?,?,?,?,?)`, "old-link", "old-account", "old-user", "retained", strings.Repeat("a", 64), "2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 2; n++ {
		if err := a.migrateExternalIMAPShares(); err != nil {
			t.Fatalf("migration attempt %d: %v", n+1, err)
		}
	}
	var id, label string
	var ciphertext sql.NullString
	if err := a.db.QueryRow(`SELECT id,label,token_ciphertext FROM external_imap_share_links WHERE id='old-link'`).
		Scan(&id, &label, &ciphertext); err != nil {
		t.Fatal(err)
	}
	if id != "old-link" || label != "retained" || ciphertext.Valid {
		t.Fatalf("legacy row modified: id=%s label=%s encrypted=%v", id, label, ciphertext)
	}
}
