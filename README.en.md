# LanQin MailShare Lite

[简体中文](./README.md) | **English**

> An unofficial, community-maintained Lite derivative of [LanQin Email](https://github.com/LanQin996/LanQin-Email), focused on **external IMAP mailbox aggregation, browsing, and read-only sharing**. This is not an official release from the upstream author.

![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-supported-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-GPL--3.0--only-blue)

## Overview

MailShare Lite is intended for personal self-hosting. It retains the components needed to browse, manage, and share messages from existing IMAP mailboxes while reducing the runtime footprint of the full mail server. **It is not a standalone public SMTP mail server.**

### Features

- **External IMAP accounts:** connect existing mailboxes, browse messages, and optionally synchronize them to a local SQLite cache.
- **Read-only sharing:** manage, expire, and revoke controlled mailbox access without granting visitors permission to send mail.
- **Simplified interface:** unified mailbox, sharing, and account settings for desktop and mobile.
- **Three UI languages:** Simplified Chinese, Traditional Chinese, and English. Users can switch languages and keep their preference.
- **Smaller container:** Go API, React, Nginx, and SQLite, without the Postfix, Dovecot, or Rspamd daemons.

### Lite vs. upstream

| Area | MailShare Lite | Upstream LanQin Email |
| --- | --- | --- |
| Primary purpose | IMAP aggregation and read-only sharing | Full webmail and self-hosted mail service |
| External IMAP | Supported | Supported |
| Sharing and read-only browsing | Supported | Part of the broader mail management platform |
| SMTP sending and full mail delivery stack | Disabled in Lite mode | Supported |
| Runtime | Go + React + Nginx + SQLite | Full integrated mail stack |
| UI languages | Simplified Chinese, Traditional Chinese, English | Refer to upstream |

> Lite mode disables outgoing-mail workers and related services, but some upstream API route implementations remain in the code. **Do not expose this deployment to the public internet without an independent security review.**

## Branches

| Branch | Purpose |
| --- | --- |
| [`main`](../../tree/main) | Fork default branch with Lite-related code and documentation |
| [`lite`](../../tree/lite) | Lite version used for deployment |
| [`dev`](../../tree/dev) | Development and integration testing |
| [`original-archive`](../../tree/original-archive) | Snapshot of the original upstream code |

Original project: [LanQin996/LanQin-Email](https://github.com/LanQin996/LanQin-Email). Maintenance of this fork prioritizes **bug fixes, security fixes, and compatibility**, rather than mirroring every upstream feature.

## Build and deployment

The repository provides a dedicated Lite Dockerfile and Compose configuration:

- [Lite deployment files](./deploy/mailshare-lite/)
- [Lite instructions and security notes](./deploy/mailshare-lite/README.md)

Build from the repository root:

```bash
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
```

The `lite-arm64` tag is an example label; the actual target architecture depends on your Docker build environment. Adjust it for other architectures.

Provide your own `.env`, persistent data directory, and Compose deployment directory. Never commit real mailbox passwords, database files, or encryption secrets. The current example Compose configuration only publishes HTTP to localhost:

```text
127.0.0.1:18090 -> container port 80
```

Use a local connection or SSH port forwarding. **Do not directly expose this service or publish it through a public Cloudflare Tunnel before reviewing its security boundaries.**

## Data and security

- Protect stored messages, SQLite data, and external account credentials.
- Use consistent snapshots for live SQLite databases, encrypted backups, and periodic restore verification.
- Verify revocation, access control, account isolation, and remaining API routes before upgrades.
- This self-maintained Lite fork **has not been represented as security-audited for public production hosting**.

## Credits and license

Thanks to [LanQin996](https://github.com/LanQin996) and the upstream contributors. For the full mail server and its deployment instructions, see the [original project](https://github.com/LanQin996/LanQin-Email).

See [GPL-3.0-only](./LICENSE) for the current repository license. Earlier revisions remain subject to the licenses applicable when they were published.
