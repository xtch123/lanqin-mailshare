# 外部 IMAP 只读共享设计（开发分支）

> 项目：`xtch123/lanqin-mailshare`；上游：`LanQin996/LanQin-Email`。  
> 状态：设计基线，**功能尚未实现**。不得将本文件视为可对外开放共享的完成证明。

## 已验证的原版能力

- OCI ARM64 原版容器 `lanqin-test` 正常运行，`/readyz` 为 HTTP 200。
- seek.li IMAP 连接、同步邮件、阅读 HTML 正文和打开正文超链接均已手动验证。
- `external_imap_accounts` 记录外部邮箱，`external_imap_messages` 记录 `account_id → local_message_id` 映射；本地存储模式为 `local`。
- 原版 `mailbox_sharing.go` 共享的是内部 `mailboxes`，不能用于跨外部 IMAP 账号隔离；不能直接共享管理员内部邮箱。
- 原版 `router_auth.go` 目前要求登录才能调用 `/api/mail/external-accounts/*`，不得直接绕过鉴权。

## 首版目标

1. 管理员通过现有界面管理多个外部 IMAP 邮箱，不改原有添加/同步流程。
2. 每个外部 IMAP 账号可以创建多个独立只读共享链接，车友无需注册本站账号。
3. 每个共享链接独立配置：备注、绝对到期时间（或永久有效）、邮件可见时间范围（最近 1 小时/24 小时/7 天/全部，以及自定义值）。
4. 管理员可查看共享记录、撤销单个共享、查看最近访问时间。撤销和过期必须立即阻止后续 API 调用。
5. 仅显示指定外部账号已同步的邮件，包括完整、经过安全过滤的 HTML 正文和可用的 HTTPS 验证链接；允许打开 Netflix 同户/临时访问链接，不自动代替用户执行任何链接操作。
6. 只读：不提供发信、删除、移动、标记已读、修改标签或修改外部邮箱凭据接口；不改变 seek.li 原件。
7. 原版 `lanqin-test` 不直接覆盖；新版本在独立测试容器与独立 SQLite 数据上验收。

## 数据隔离与访问控制

- 共享主体必须绑定 **`external_imap_accounts.id`**，不能只用 `mailboxes.id`。
- 管理 API 必须同时验证当前登录用户拥有目标外部账号；车友 API 只接受共享凭据，不继承管理员 Session。
- 每次获取列表、详情或附件时均验证：共享存在、未撤销、未过期，目标外部账号仍启用，目标邮件有 `external_imap_messages.account_id = share.account_id` 的映射，并满足可见时间范围。
- 数据访问必须采用参数化 SQL 和服务端范围校验；不得只依靠前端隐藏。跨账号邮件 ID 和附件 ID 应返回 404，不能泄露其它邮箱存在性。
- 共享 Token 使用密码学安全随机数（不少于 256 bit），数据库只存哈希。只在创建时显示完整链接一次；列表接口不返回原始 Token。
- 优先把 Token 放在 URL fragment（不发送到服务器访问日志），由网页用请求头提交；响应设置 `Cache-Control: no-store`、`Referrer-Policy: no-referrer`，禁用索引，并配置基本限流。
- 邮件 HTML 必须在隔离 iframe 中显示，经严格清理，禁止脚本、表单、iframe 嵌套、javascript: 等可执行内容。外链只允许安全协议（HTTPS 等），`target=_blank` 同时使用 `rel=noopener noreferrer`；默认不加载远程图片。原文链接保留可点击能力。
- 附件下载应单独鉴权，强制安全 Content-Disposition 与危险 MIME 下载，避免浏览器执行附件。
- 建议“一个拼车账号一个专用邮箱”，并提示原始登录/同户验证链接本身拥有授权能力。

## 里程碑与验收

- **M1 数据库 + 管理 API**：创建、列出、撤销单个外部邮箱共享；单元测试覆盖非法有效期、无权限、撤销及到期。
- **M2 车友只读 API**：列表、详情、附件；所有查询必须经过 `account_id` 映射限制；测试跨账号隔离、邮件历史范围、Token 猜测/伪造。
- **M3 前端界面**：外部邮箱管理面板添加「共享设置」；独立移动端只读邮件页；完整 HTML 与原始 HTTPS 验证链接。
- **M4 集成验收**：两个 seek.li 邮箱、多个车友链接、撤销、过期、XSS、附件、Netflix 真正验证邮件、OCI ARM64 镜像构建测试。
- **M5 灰度发布**：专用容器及数据库、HTTPS + Cloudflare Tunnel，不暴露 SMTP/IMAP 端口；完成备份范围与还原验证后替换测试实例。

## 不做的事情

- 不自己搭建公网收件服务器、修改域名 MX 或在 OCI 公网暴露 25/465/587/993/995。
- 不替代或修改 Netflix 平台的同户规则。
- 不在共享访问页面暴露普通 Webmail、管理员、账户凭据或者 API Token。
- 不用“清理网页邮件”去删除 seek.li 原始邮件。
