# LanQin MailShare Lite

[简体中文](./README.md) | **English**

> An **unofficial, independently maintained, self-hosting-focused** Lite derivative of [LanQin Email](https://github.com/LanQin996/LanQin-Email). It connects existing IMAP mailboxes and offers revocable, read-only email sharing. This is not an official upstream Lite release or a standalone public SMTP mail server.

![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-self--hosted-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-GPL--3.0--only-blue)

## Current status (November 2026)

- **Verified personal deployment:** OCI ARM64, Docker Compose, SQLite, external IMAP synchronization, and an HTTPS guest inbox accessed from desktop and mobile browsers.
- **Incremental sync issue fixed:** some IMAP servers return older UIDs in search results. They could fill every sync batch, producing "0 imported / 100 skipped" while still reporting `ok`. The `lite` branch now filters stale UIDs and includes regression tests; a live deployment has subsequently imported new messages.
- **Public access is separated:** administrators use a loopback-only service; guests reach a dedicated, read-only reverse proxy over Cloudflare Tunnel.
- **No prebuilt official image or GitHub Release:** build and deploy from source. These observations are **not** a public production security audit or a compatibility guarantee for all IMAP providers.

## Features

- **External IMAP aggregation:** connect existing mailboxes, browse messages, and synchronize into local SQLite and message storage.
- **Read-only share links:** scope links to one external IMAP account; configure visible-message history and link expiry; retrieve, rotate, and revoke links.
- **Guest inbox:** mobile and desktop layout, periodic refresh, read-only message bodies and attachments; no ability to send/delete mail or change read state.
- **Languages:** Simplified Chinese, Traditional Chinese, and English.
- **Lean runtime:** Go API, React, Nginx, SQLite; no Postfix, Dovecot, or Rspamd daemon.

**Link expiry** and **message-visibility history** are separate. A link valid for 30 days and limited to messages from the last 24 hours does not expose 30 days of past mail.

## Branches

| Branch | Purpose |
| --- | --- |
| [`lite`](../../tree/lite) | **Actively used Lite source**, including the latest sharing and IMAP sync fixes |
| [`main`](../../tree/main) | Repository default branch; **do not assume its application source matches `lite`** |
| [`dev`](../../tree/dev) | Historical development/integration testing |
| [`original-archive`](../../tree/original-archive) | Snapshot of the original upstream source |

**Explicitly switch to `lite` before building.** Do not build the default `main` branch by mistake.

## Build from source

```bash
git clone https://github.com/xtch123/lanqin-mailshare.git
cd lanqin-mailshare
git switch lite
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
```

The `lite-arm64` label is an example; the target platform is determined by your Docker build environment. Provide a private `.env`, persistent SQLite and mail directories, and use the [Lite Compose example](./deploy/mailshare-lite/compose.yaml). Never commit real credentials, secrets, or databases.

See the **[Lite deployment and operations guide](./deploy/mailshare-lite/README.md)** for a private administration setup, restricted guest gateway, Cloudflare Tunnel notes, and upgrade checks.

## Separate administration from public sharing

```text
Admin -- SSH tunnel --> 127.0.0.1:18090 --> full Lite app
Guest -- HTTPS --> Cloudflare Tunnel --> 127.0.0.1:18091
                                         restricted GET/HEAD gateway
                                              --> 127.0.0.1:18090
```

The sample Compose deployment binds only to `127.0.0.1:18090`. **Do not expose the full application on the public internet.** Instead, publish a separate proxy which permits only the `/api/public/shared-inbox` page, script, message-list/detail, and attachment-read endpoints, and rejects every other path.

A dedicated optional environment variable generates public guest links independently of the local admin URL:

```dotenv
LANQIN_EXTERNAL_IMAP_SHARE_BASE_URL=https://share.example.com
```

Generated links then look like `https://share.example.com/api/public/shared-inbox#<access-token>`. When unset, this falls back to `LANQIN_PUBLIC_BASE_URL`, which remains a localhost URL in the Lite Compose sample.

## Security and maintenance

- **Anyone holding the full share link holds read access to the scoped mailbox.** Protect tokens, restrict visibility/expiry, and revoke or rotate when appropriate.
- **Read-only still means sensitive data:** mail bodies, verification codes, and attachments can be confidential. Prefer a mailbox dedicated to sharing.
- Lite mode disables parts of the upstream service but some upstream API routes still exist. A restricted guest proxy **does not constitute a complete security audit**; never publish all administration/API endpoints.
- Monitor actual IMAP import counts, not just a status of `ok`. UID/UIDVALIDITY behavior varies across providers and needs continued validation.
- Back up your `.env`, mail storage, and a **consistent SQLite snapshot** before upgrades; retain rollback images and verify guest access, attachments, revocation, and unauthenticated `401` responses.
- Maintenance is primarily **bug fixes, security improvements, and essential compatibility updates**. There is no fixed release schedule or commitment to track every upstream feature.

## Credits and license

Thanks to [LanQin996](https://github.com/LanQin996) and the upstream contributors. See the [upstream LanQin Email project](https://github.com/LanQin996/LanQin-Email) for the full mail server.

The current repository license is [GPL-3.0-only](./LICENSE). Earlier revisions remain subject to their applicable licenses.
