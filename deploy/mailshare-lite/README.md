# LanQin MailShare Lite — isolated developer smoke test

This is **an optional separate build**, not a replacement for the current
`lanqin-mailshare-dev` on `127.0.0.1:18089` and not the original
`lanqin-test` on `127.0.0.1:18088`.

## Scope

* Go API + external IMAP sync (INBOX first) + SQLite + local message cache.
* Nginx with a same-origin authenticated sharing manager iframe.
* Owner sidebar: external inboxes, IMAP account setup, sharing manager,
  account settings. No visible SMTP, drafts, send queue, mail labels or
  generic folder management.
* Public guest inbox retains read-only token access and five-second
  foreground polling. No Postfix, Dovecot, Rspamd, DKIM daemon, Supervisor,
  database clients or SMTP listeners.
* `LANQIN_MAILSHARE_LITE=true` disables outgoing, local-delivery,
  notifications and native backup workers **in the Go app**. It does not
  remove the upstream API's entire send-route implementation; avoid making
  this service publicly reachable until the Lite-only server routes and
  security are independently reviewed.

## First build and isolated local test (OCI ARM64)

From repository checkout:

```bash
cd /opt/lanqin-mailshare-dev
git pull --ff-only
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
python3 deploy/mailshare-lite/prepare.py --source /opt/lanqin-mailshare-test --target /opt/lanqin-mailshare-lite
cd /opt/lanqin-mailshare-lite
docker compose up -d
curl -fsS --retry 20 --retry-delay 2 --retry-connrefused --retry-all-errors http://127.0.0.1:18090/readyz
docker stats --no-stream lanqin-mailshare-dev lanqin-mailshare-lite
docker image ls lanqin-mailshare:dev-arm64 lanqin-mailshare:lite-arm64
```

**This creates a separate copied snapshot of the development SQLite DB**
using Python's online backup interface (not a raw copy of a live WAL).
It also copies local attachment/mail files and the dev test `.env` so the
accounts can be decrypted by the same existing key. Keep the Lite test
workspace private; both instances contain sensitive mailbox contents and
account credentials.

The Lite workspace uses a distinct `/opt/lanqin-mailshare-lite` data
directory and binds HTTP to **127.0.0.1:18090 only**. Access through a
dedicated SSH local tunnel (or localhost within the OCI server). Do NOT
share `http://localhost:18090` with friends: it is a local test address.

```powershell
ssh -i "C:\Users\xtch1\Desktop\圣何塞\vps秘钥\ssh-key-2026-10-16.key" -o ExitOnForwardFailure=yes -N -L 127.0.0.1:18090:127.0.0.1:18090 ubuntu@64.181.233.181
```

Test `http://localhost:18090` and `http://localhost:18090/?view=shares`
on desktop and mobile-width browser emulation. Sharing pages and synced
mail must remain scoped to one external IMAP account; revoked links fail
authorization. The source dev instance is **not restarted**.

After creating the clone, both test instances can independently poll
the same seek.li account. Avoid high-volume or destructive testing:
read-state sync and any remotely mutating operations require extra care.

## Deletion and existing local-cache folders

External IMAP sync creates internal storage folders named
`External <account-key> <folder-hash>`. Deleting the external account
removes its connection configuration but does not automatically erase
local messages. The full web UI filters these technical folders out of
the regular mailbox sidebar. The Lite sidebar displays only actual
connected external accounts. This is deliberately non-destructive.

## Security and release limits

Do not expose this test image to a public Cloudflare Tunnel before
replacing `LANQIN_ALLOW_INSECURE_HTTP=true` with a reviewed TLS setup,
reviewing remaining owner routes (including sending/deleting), confirming
backups and cross-account isolation, and completing a real-phone test.
Also ensure any future public URL is set through
`LANQIN_PUBLIC_BASE_URL`; localhost links are not externally reachable.

Avoid checking in `.env`, database snapshots, private keys or the
local `mail` and `data` directories. Do not delete the original
development or LanQin containers during evaluation.
