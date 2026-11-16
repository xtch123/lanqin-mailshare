# LanQin MailShare Lite：部署与运维指南

[← 项目首页（简体中文）](../../README.md) | [English overview](../../README.en.md)

> 适用于 `lite` 分支的源码自建部署。以下公网共享方式是**示例防护架构**，不是对所有版本或环境的安全审计。**不要直接公开完整管理/API 服务。**

## 1. 部署结构

```text
管理者 -- SSH 本地转发 --> 127.0.0.1:18090 --> lanqin-mailshare-lite
访客 -- HTTPS --> Cloudflare Tunnel --> 127.0.0.1:18091
                                        mailshare-public-gateway (Nginx)
                                               仅允许共享 GET/HEAD
                                                     |
                                                     v
                                            127.0.0.1:18090
```

- `18090`：完整 Lite 界面，仅监听 OCI 本机回环地址；不要直接暴露到公网。
- `18091`：可选的独立共享网关，也仅监听本机回环地址。
- `cloudflared` 可复用已有的 Tunnel；它访问网关，不直接访问完整 Lite 后台。
- 共享页面固定路径：`/api/public/shared-inbox`；完整链接末尾的 `#<访问令牌>` 是访问凭证。

## 2. 构建和本地部署

从 `lite` 分支构建。请根据自己的平台调整镜像标签：

```bash
git clone https://github.com/xtch123/lanqin-mailshare.git
cd lanqin-mailshare
git switch lite
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
```

[Compose 示例](./compose.yaml) 使用镜像 `lanqin-mailshare:lite-arm64`。将它复制到独立的私有运行目录，在该目录创建 `.env`、`data/`、`mail/`，再执行 `docker compose up -d`。**不要将私有数据目录放在 Git 仓库内提交。**

`.env` 至少需要根据实际账号设置管理员凭据与 `LANQIN_EXTERNAL_IMAP_SECRET_KEY` 等必要密钥；密钥必须安全生成并长期保存，**不要随意更换已有系统的 IMAP 加密密钥**，否则已保存的账号凭据可能无法解密。Compose 示例已启用 Lite、SQLite、外部 IMAP 同步，并将完整应用映射为 `127.0.0.1:18090:80`。

确认服务准备就绪：

```bash
docker compose ps && curl -fsS http://127.0.0.1:18090/readyz
```

管理员可通过 SSH 本地端口转发连接服务器上的 `127.0.0.1:18090`，在本地浏览器打开 `http://localhost:18090`。切勿让终端输出或 README 包含真实服务器私钥路径、密码、数据库或完整共享 Token。

## 3. 公网共享地址与网关

**先让本地 Lite 正常同步邮件，再考虑公开只读共享。**

在运行目录的私有 `.env` 中新增（替换为你自己的域名）：

```dotenv
LANQIN_EXTERNAL_IMAP_SHARE_BASE_URL=https://share.example.com
```

这是**独立的共享链接基础地址**。它不要求改变现有的 `LANQIN_PUBLIC_BASE_URL=http://localhost:18090`。修改 `.env` 后需要重新创建 Lite 容器，环境变量才会生效。

示例 Nginx 网关配置（`default.conf`），适用于 **`network_mode: host` 的独立网关容器**：

```nginx
server {
    listen 127.0.0.1:18091;
    server_name _;
    server_tokens off;
    add_header Cache-Control "no-store" always;
    add_header Referrer-Policy "no-referrer" always;
    add_header X-Content-Type-Options "nosniff" always;

    proxy_set_header Host $host;
    proxy_set_header True-Client-IP "";
    proxy_set_header X-Forwarded-For "";

    if ($request_method !~ ^(GET|HEAD)$) { return 405; }

    location = /api/public/shared-inbox { proxy_pass http://127.0.0.1:18090; }
    location = /api/public/shared-inbox/app.js { proxy_pass http://127.0.0.1:18090; }
    location = /api/public/shared-inbox/messages { proxy_pass http://127.0.0.1:18090; }
    location ~ ^/api/public/shared-inbox/messages/[^/]+$ {
        proxy_pass http://127.0.0.1:18090;
    }
    location ~ ^/api/public/shared-inbox/messages/[^/]+/attachments/[^/]+$ {
        proxy_pass http://127.0.0.1:18090;
    }
    location / { return 404; }
}
```

示例的网关 Compose 文件：

```yaml
services:
  mailshare-public-gateway:
    image: nginx:stable-alpine
    container_name: mailshare-public-gateway
    network_mode: host
    restart: unless-stopped
    volumes:
      - ./default.conf:/etc/nginx/conf.d/default.conf:ro
    security_opt:
      - no-new-privileges:true
```

按需增加镜像锁定、请求限速和日志保留策略。网关是额外隔离层；它不能替代服务本身的令牌验证和安全审计。

## 4. Cloudflare Tunnel（本地管理）

如果 Tunnel 使用**本地 `config.yml`** 管理，不要在 Cloudflare 控制台里直接修改 ingress。备份现有配置，将以下规则追加到**兜底 `http_status:404` 之前**（不要覆盖其他站点的规则）：

```yaml
ingress:
  # 保留其他已存在的 hostname 规则
  - hostname: share.example.com
    service: http://127.0.0.1:18091
  - service: http_status:404
```

验证后重启 cloudflared，并给 `share.example.com` 配置一条橙云 CNAME，指向同一个 Tunnel 的 `<tunnel-uuid>.cfargotunnel.com`。远程管理的 Tunnel 请使用与其模式相匹配的路由配置方式。

访客不应经过完整 Lite 的 `18090` 入口，也无需获取管理员登录凭据。共享 Token 必须保密；Cloudflare Access 登录并非默认共享模型的必需部分。

## 5. 验收清单

```bash
curl -sS -o /dev/null -w '共享页 %{http_code}\n' https://share.example.com/api/public/shared-inbox
curl -sS -o /dev/null -w '无令牌 %{http_code}\n' https://share.example.com/api/public/shared-inbox/messages
curl -sS -o /dev/null -w '后台 %{http_code}\n' https://share.example.com/login
```

预期为 `200`、`401`、`404`。**但这仅证明部分路径行为符合预期，不代表完整安全审计通过。**

再用真实测试邮箱验证：最近收到的邮件能够导入、共享列表显示、正文与附件可读、链接可复制、失效或撤销后访问失败。访客不能发送邮件或访问管理 API。邮件正文中的第三方链接（例如 Netflix）是否能打开，取决于访客自己的网络可达性，并不由 Lite 或 Cloudflare 网关保证。

## 6. IMAP 同步、备份和升级

- `lite` 已修复一类**旧 UID 被搜索结果重复返回、占满 100 封同步批次**的问题，并添加测试。实际部署曾从“0 导入 / 100 跳过”恢复为导入新邮件、后续无重复跳过。
- 同步显示 `ok` 并不等同于有新邮件成功导入。需要同时观察 `external_imap_sync_runs` 的 `imported`/`skipped`/`failed` 和本地邮件最新接收时间；不要在未诊断前手动清空 `external_imap_folder_states` 或修改 UID。
- 其它 IMAP 提供商的 UIDVALIDITY 变更、异常 SEARCH 结果及大邮箱行为仍需进一步验证。
- 升级前备份 `.env`、邮件附件目录，以及通过 SQLite Backup API 获得的**一致性数据库快照**；保留旧镜像，测试新镜像后再重建运行容器。数据库备份应加密保存并定期验证恢复。
- 如果需要将服务提供给其他人，应自行完成威胁建模、接口审计、限流和长期维护；本仓库当前不是经正式安全审计的托管服务。

## 历史与许可

本目录旧版说明曾用于 `lanqin-mailshare-dev` 和 `lanqin-test` 的隔离试验；这些测试实例与路径已不再是当前推荐部署步骤。当前请以 `lite` 分支和独立私有运行目录为准。

源项目：[LanQin996/LanQin-Email](https://github.com/LanQin996/LanQin-Email)；许可证：[GPL-3.0-only](../../LICENSE)。
