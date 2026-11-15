package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/microcosm-cc/bluemonday"
)

// External IMAP sharing is deliberately separate from internal mailbox sharing.
// Access MUST be scoped through external_imap_messages.account_id on EVERY read.

const maxShareVisibilityMinutes int64 = 525600 // one year; zero means all retained messages

type externalIMAPShareLink struct {
	ID                string  `json:"id"`
	AccountID         string  `json:"accountId"`
	Label             string  `json:"label"`
	VisibilityMinutes int64   `json:"visibilityMinutes"`
	CreatedAt         string  `json:"createdAt"`
	ExpiresAt         *string `json:"expiresAt"`
	RevokedAt         *string `json:"revokedAt"`
	LastAccessedAt    *string `json:"lastAccessedAt"`
}

type externalIMAPShareAccess struct {
	ID                string
	AccountID         string
	MailboxID         string
	AccountName       string
	VisibilityMinutes int64
}

func (a *App) migrateExternalIMAPShares() error {
	// Compatible column types for SQLite, MySQL 8, and PostgreSQL 16.
	// This additive table does not change the upstream schema or restore files.
	_, err := a.db.Exec(`CREATE TABLE IF NOT EXISTS external_imap_share_links (
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
	)`)
	return err
}

func newExternalIMAPShareToken() (string, string, error) {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes[:])
	digest := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(digest[:]), nil
}

func validateExternalIMAPShareWindow(minutes int64, rawExpiry *string, now time.Time) (*string, error) {
	if minutes < 0 || minutes > maxShareVisibilityMinutes {
		return nil, errors.New("visibilityMinutes must be 0 (all), or 1..525600")
	}
	if rawExpiry == nil || strings.TrimSpace(*rawExpiry) == "" {
		return nil, nil
	}
	expiry, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(*rawExpiry))
	if err != nil || !expiry.After(now) {
		return nil, errors.New("expiresAt must be a future RFC3339 timestamp")
	}
	utc := expiry.UTC().Format(time.RFC3339Nano)
	return &utc, nil
}

func (a *App) handleCreateExternalIMAPShare(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AccountID         string  `json:"accountId"`
		Label             string  `json:"label"`
		VisibilityMinutes int64   `json:"visibilityMinutes"`
		ExpiresAt         *string `json:"expiresAt"`
	}
	if err := decodeJSON(r, &input); err != nil {
		badRequest(w, err)
		return
	}
	input.AccountID = strings.TrimSpace(input.AccountID)
	input.Label = strings.TrimSpace(input.Label)
	if utf8.RuneCountInString(input.Label) > 64 || input.AccountID == "" {
		badRequest(w, errors.New("accountId is required and label must be at most 64 characters"))
		return
	}
	now := a.now().UTC()
	expiry, err := validateExternalIMAPShareWindow(input.VisibilityMinutes, input.ExpiresAt, now)
	if err != nil {
		badRequest(w, err)
		return
	}
	account, err := a.externalIMAPAccountForUser(r.Context(), currentUser(r).ID, input.AccountID)
	if err != nil || !account.Enabled || account.StorageMode != externalIMAPStorageLocal {
		respondError(w, http.StatusNotFound, "enabled local IMAP account not found")
		return
	}
	token, digest, err := newExternalIMAPShareToken()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "could not generate share token")
		return
	}
	id := newID("xshare")
	stamp := now.Format(time.RFC3339Nano)
	var expires any
	if expiry != nil {
		expires = *expiry
	}
	_, err = a.db.ExecContext(r.Context(), `INSERT INTO external_imap_share_links
		(id,account_id,owner_user_id,label,token_sha256,visibility_minutes,created_at,expires_at)
		VALUES(?,?,?,?,?,?,?,?)`,
		id, account.ID, account.UserID, input.Label, digest, input.VisibilityMinutes, stamp, expires)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create share")
		return
	}
	base := strings.TrimRight(a.cfg.PublicBaseURL, "/")
	if base == "" {
		base = "http://" + r.Host
	}
	respondJSON(w, http.StatusCreated, map[string]any{
		"id": id, "accountId": account.ID, "label": input.Label,
		"visibilityMinutes": input.VisibilityMinutes, "expiresAt": expiry,
		"url": base + "/api/public/shared-inbox#" + token,
	})
}

func (a *App) handleListExternalIMAPShares(w http.ResponseWriter, r *http.Request) {
	accountID := strings.TrimSpace(r.URL.Query().Get("accountId"))
	if accountID != "" {
		if _, err := a.externalIMAPAccountForUser(r.Context(), currentUser(r).ID, accountID); err != nil {
			respondError(w, http.StatusNotFound, "account not found")
			return
		}
	}
	query := `SELECT s.id,s.account_id,s.label,s.visibility_minutes,s.created_at,s.expires_at,s.revoked_at,s.last_accessed_at
		FROM external_imap_share_links s JOIN external_imap_accounts a ON a.id=s.account_id
		WHERE a.user_id=? AND s.owner_user_id=a.user_id`
	args := []any{currentUser(r).ID}
	if accountID != "" {
		query += " AND s.account_id=?"
		args = append(args, accountID)
	}
	query += " ORDER BY s.created_at DESC LIMIT 500"
	rows, err := a.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to read shares")
		return
	}
	defer rows.Close()
	items := []externalIMAPShareLink{}
	for rows.Next() {
		var item externalIMAPShareLink
		var expiry, revoked, accessed sql.NullString
		if err := rows.Scan(&item.ID, &item.AccountID, &item.Label, &item.VisibilityMinutes, &item.CreatedAt, &expiry, &revoked, &accessed); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to read shares")
			return
		}
		if expiry.Valid {
			item.ExpiresAt = &expiry.String
		}
		if revoked.Valid {
			item.RevokedAt = &revoked.String
		}
		if accessed.Valid {
			item.LastAccessedAt = &accessed.String
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to read shares")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *App) handleRevokeExternalIMAPShare(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	result, err := a.db.ExecContext(r.Context(), `UPDATE external_imap_share_links
		SET revoked_at=?
		WHERE id=? AND revoked_at IS NULL AND EXISTS (
			SELECT 1 FROM external_imap_accounts a
			WHERE a.id=external_imap_share_links.account_id
			AND a.user_id=? AND a.user_id=external_imap_share_links.owner_user_id
		)`, a.now().UTC().Format(time.RFC3339Nano), id, currentUser(r).ID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to revoke share")
		return
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		respondError(w, http.StatusNotFound, "active share not found")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Public requests NEVER acquire a regular user context or reuse the owner's
// authenticated external-IMAP endpoints. Each request is independently scoped.
func (a *App) externalIMAPShareFromRequest(r *http.Request) (*externalIMAPShareAccess, error) {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(value, "Bearer ") {
		return nil, errors.New("missing token")
	}
	token := strings.TrimPrefix(value, "Bearer ")
	if len(token) != 43 {
		return nil, errors.New("invalid token")
	}
	digest := sha256.Sum256([]byte(token))
	var access externalIMAPShareAccess
	var expiry sql.NullString
	err := a.db.QueryRowContext(r.Context(), `SELECT s.id,s.account_id,a.mailbox_id,a.name,s.visibility_minutes,s.expires_at
		FROM external_imap_share_links s
		JOIN external_imap_accounts a ON a.id=s.account_id
		WHERE s.token_sha256=? AND s.revoked_at IS NULL AND a.enabled=1
			AND a.storage_mode='local' AND s.owner_user_id=a.user_id`,
		hex.EncodeToString(digest[:])).Scan(&access.ID, &access.AccountID, &access.MailboxID,
		&access.AccountName, &access.VisibilityMinutes, &expiry)
	if err != nil {
		return nil, errors.New("share not available")
	}
	if expiry.Valid {
		end, err := time.Parse(time.RFC3339Nano, expiry.String)
		if err != nil || !a.now().Before(end) {
			return nil, errors.New("share expired")
		}
	}
	// Throttle bookkeeping; failure must not change access rights.
	_, _ = a.db.ExecContext(r.Context(), `UPDATE external_imap_share_links SET last_accessed_at=?
		WHERE id=? AND (last_accessed_at IS NULL OR last_accessed_at<?)`,
		a.now().UTC().Format(time.RFC3339Nano), access.ID, a.now().Add(-10*time.Minute).UTC().Format(time.RFC3339Nano))
	return &access, nil
}

func (a *App) requireExternalIMAPShare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		access, err := a.externalIMAPShareFromRequest(r)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired sharing link")
			return
		}
		next.ServeHTTP(w, r.WithContext(withExternalShareAccess(r.Context(), access)))
	})
}

type externalShareContextKey struct{}

func withExternalShareAccess(ctx context.Context, access *externalIMAPShareAccess) context.Context {
	return context.WithValue(ctx, externalShareContextKey{}, access)
}

func shareAccessFromRequest(r *http.Request) *externalIMAPShareAccess {
	access, _ := r.Context().Value(externalShareContextKey{}).(*externalIMAPShareAccess)
	return access
}

func (a *App) shareMessageIsVisible(received string, minutes int64) bool {
	if minutes == 0 {
		return true
	}
	stamp, err := time.Parse(time.RFC3339Nano, received)
	return err == nil && !stamp.After(a.now()) &&
		!stamp.Before(a.now().Add(-time.Duration(minutes)*time.Minute))
}

type externalShareMessage struct {
	ID         string `json:"id"`
	Subject    string `json:"subject"`
	From       string `json:"from"`
	FromName   string `json:"fromName"`
	ReceivedAt string `json:"receivedAt"`
	Snippet    string `json:"snippet"`
	BodyText   string `json:"bodyText,omitempty"`
	BodyHTML   string `json:"bodyHtml,omitempty"`
}

func (a *App) handleExternalIMAPShareMessages(w http.ResponseWriter, r *http.Request) {
	access := shareAccessFromRequest(r)
	rows, err := a.db.QueryContext(r.Context(), `SELECT m.id,m.subject,m.from_addr,COALESCE(m.from_name,''),
		m.received_at,COALESCE(m.snippet,'')
		FROM external_imap_messages em JOIN messages m ON m.id=em.local_message_id
		WHERE em.account_id=? AND m.mailbox_id=?
		ORDER BY m.received_at DESC,m.id DESC LIMIT 300`, access.AccountID, access.MailboxID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	defer rows.Close()
	items := []externalShareMessage{}
	for rows.Next() {
		var item externalShareMessage
		if err := rows.Scan(&item.ID, &item.Subject, &item.From, &item.FromName, &item.ReceivedAt, &item.Snippet); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to list messages")
			return
		}
		if a.shareMessageIsVisible(item.ReceivedAt, access.VisibilityMinutes) {
			items = append(items, item)
		}
	}
	if err := rows.Err(); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"accountName": access.AccountName, "items": items})
}

func (a *App) handleExternalIMAPShareMessage(w http.ResponseWriter, r *http.Request) {
	access := shareAccessFromRequest(r)
	var item externalShareMessage
	err := a.db.QueryRowContext(r.Context(), `SELECT m.id,m.subject,m.from_addr,
		COALESCE(m.from_name,''),m.received_at,COALESCE(m.snippet,''),
		COALESCE(m.body_text,''),COALESCE(m.body_html,'')
		FROM external_imap_messages em JOIN messages m ON m.id=em.local_message_id
		WHERE em.account_id=? AND m.mailbox_id=? AND m.id=? LIMIT 1`,
		access.AccountID, access.MailboxID, chi.URLParam(r, "id")).Scan(
		&item.ID, &item.Subject, &item.From, &item.FromName, &item.ReceivedAt,
		&item.Snippet, &item.BodyText, &item.BodyHTML)
	if err != nil || !a.shareMessageIsVisible(item.ReceivedAt, access.VisibilityMinutes) {
		respondError(w, http.StatusNotFound, "message not found")
		return
	}
	item.BodyHTML = bluemonday.UGCPolicy().Sanitize(item.BodyHTML)
	respondJSON(w, http.StatusOK, item)
}

func (a *App) handleExternalIMAPShareAttachment(w http.ResponseWriter, r *http.Request) {
	access := shareAccessFromRequest(r)
	var path, filename, received string
	err := a.db.QueryRowContext(r.Context(), `SELECT att.storage_path,att.filename,m.received_at
		FROM attachments att
		JOIN messages m ON m.id=att.message_id
		JOIN external_imap_messages em ON em.local_message_id=m.id
		WHERE att.id=? AND m.id=? AND m.mailbox_id=? AND em.account_id=? LIMIT 1`,
		chi.URLParam(r, "attachmentId"), chi.URLParam(r, "id"), access.MailboxID, access.AccountID).
		Scan(&path, &filename, &received)
	if err != nil || !a.shareMessageIsVisible(received, access.VisibilityMinutes) {
		respondError(w, http.StatusNotFound, "attachment not found")
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		respondError(w, http.StatusNotFound, "attachment not found")
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	w.Header().Set("Content-Length", fmt.Sprint(len(data)))
	_, _ = w.Write(data)
}
