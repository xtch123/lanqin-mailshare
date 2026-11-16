# LanQin MailShare Lite

**简体中文** | [English](./README.en.md)

> 基于 [LanQin Email](https://github.com/LanQin996/LanQin-Email) 的**非官方、自用优先**精简分支：聚合已有 IMAP 邮箱，并提供可撤销的只读邮件共享。不是上游作者发布的官方 Lite 版本，也不是独立的公网 SMTP 邮件服务器。

![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-self--hosted-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-GPL--3.0--only-blue)

## 当前状态（2026 年 11 月）

- **已验证的自用部署**：OCI ARM64 + Docker Compose + SQLite，外部 IMAP 邮件同步，PC 和手机打开 HTTPS 共享页面。
- **已修复增量同步问题**：某些 IMAP 服务器返回包含旧 UID 的搜索结果时，旧消息可能反复占满同步批次，出现“导入 0 / 跳过 100、状态仍为 ok”。`lite` 分支已增加 UID 过滤和回归测试，实际部署中已验证新邮件能够继续导入。
- **公网访问使用独立只读网关**：管理员仍通过仅监听本地的服务访问，访客流量经过 Cloudflare Tunnel 和只放行共享接口的反向代理。
- **未发布官方构建镜像或 GitHub Release**：当前以源码构建、自行部署为主；这些结果不构成正式公网安全审计或所有 IMAP 提供商的兼容性保证。

## 功能

- **外部 IMAP 聚合**：接入已有邮箱，浏览邮件，可同步到本地 SQLite 与邮件文件存储。
- **只读共享**：每个共享链接限定一个外部邮箱；可设置历史邮件可见范围、链接有效期，并支持复制、轮换、撤销。
- **访客收件箱**：移动端和桌面端适配、自动刷新；可以只读查看正文和附件，不能发信、删除邮件或改变邮件已读状态。
- **语言**：简体中文、繁体中文、English。
- **轻量运行**：Go API + React + Nginx + SQLite，不运行 Postfix、Dovecot 或 Rspamd 服务。

共享链接的**有效期**与**可见邮件时间范围**是两个独立条件。例如“链接有效 30 天、只显示最近 24 小时邮件”，不意味着可查看过去 30 天的邮件。

## 分支说明

| 分支 | 用途 |
| --- | --- |
| [`lite`](../../tree/lite) | **实际维护和部署的 Lite 源码**，包含最新共享功能及 IMAP 同步修复 |
| [`main`](../../tree/main) | 仓库默认展示分支；**不应假定其应用源码与 `lite` 一致** |
| [`dev`](../../tree/dev) | 历史开发与集成测试 |
| [`original-archive`](../../tree/original-archive) | 上游原版代码快照 |

**部署时请显式切换到 `lite` 分支**，不要仅因仓库默认显示 `main` 就直接构建。

## 源码构建

```bash
git clone https://github.com/xtch123/lanqin-mailshare.git
cd lanqin-mailshare
git switch lite
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
```

`lite-arm64` 只是示例标签；实际架构由 Docker 构建环境决定。运行时需单独创建私有 `.env`、SQLite 数据目录和邮件文件目录，并使用 [Lite Compose 示例](./deploy/mailshare-lite/compose.yaml)。不要将真实密钥、邮箱凭据或数据库提交到 GitHub。

详细部署、共享网关示例、Cloudflare Tunnel 配置和升级检查见 **[Lite 部署与运维指南](./deploy/mailshare-lite/README.md)**。

## 本地管理与公网共享分离

```text
管理员 -- SSH 隧道 --> 127.0.0.1:18090 --> Lite 完整界面
访客 -- HTTPS --> Cloudflare Tunnel --> 127.0.0.1:18091
                                          共享专用网关（GET/HEAD 白名单）
                                                --> 127.0.0.1:18090
```

Compose 示例将应用监听在 `127.0.0.1:18090`，**不要将整个 18090 服务直接发布到公网**。推荐将访客流量交给另一层仅放行 `/api/public/shared-inbox` 页面、脚本、邮件列表/正文和附件读取接口的网关，并拒绝其他路径。

新的可选环境变量允许独立生成公网共享地址，而不影响本地管理入口：

```dotenv
LANQIN_EXTERNAL_IMAP_SHARE_BASE_URL=https://share.example.com
```

例如生成的链接为 `https://share.example.com/api/public/shared-inbox#<访问令牌>`。若不设置该变量，仍回退使用 `LANQIN_PUBLIC_BASE_URL`；后者在 Lite Compose 示例中保持本地地址。

## 安全与维护

- **持有完整链接就相当于持有只读邮箱访问权**。不要公开链接或令牌；应设置适当的邮件可见范围、有效期，并在必要时轮换或撤销。
- **只读不等于没有敏感数据**：邮件正文、验证码和附件可能非常敏感。优先使用专门用于共享的邮箱。
- Lite 模式关闭部分上游功能，但仍保留一些原版后端路由实现。**仅通过受限网关开放共享 API 不等于整个系统已通过安全审计**；不要直接公开管理/API 全部路径。
- 应定期检查 IMAP 同步日志和实际邮件数量，尤其留意状态为 `ok` 但新邮件长期不导入的情况；不同 IMAP 服务器的 UID/UIDVALIDITY 行为仍需持续验证。
- 升级前备份 `.env`、邮件文件及 **SQLite 一致性快照**；保留可回滚镜像，升级后检查链接访问、附件、撤销及无令牌 `401`。
- 此仓库以**修复 Bug、安全问题及必要的上游兼容性更新**为主，不承诺跟进上游全部新功能，也没有固定发版周期。

## 致谢与许可

感谢 [LanQin996](https://github.com/LanQin996) 和原项目贡献者；完整版请见 [LanQin Email 上游项目](https://github.com/LanQin996/LanQin-Email)。

许可证沿用仓库的 [GPL-3.0-only](./LICENSE)；历史版本遵循其发布时适用的许可条款。
