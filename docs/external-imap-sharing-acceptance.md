# 外部 IMAP 邮箱共享开发版 · 验收说明

> 开发分支：`feature/external-mail-sharing`。**已有 ARM64 Go 全量测试通过和 OCI 隔离开发容器构建；本轮新改动仍需重新测试。公网安全验收未完成，禁止直接对公网提供车友链接。**

## 已提交功能

- 原版外部 IMAP 管理及本地同步逻辑不改动。
- 新增 `external_imap_share_links` 表。链接与 `external_imap_accounts.id` 绑定，而不是共享管理员内部邮箱。
- 管理员页面：`/api/me/external-imap-share-panel`。需先登录 LanQin。支持选择已启用、本地同步的外部邮箱；创建备注、共享有效期、历史可见范围；按邮箱筛选已有共享、显示邮箱名称/地址、复制或撤销链接。
- 车友页面：`/api/public/shared-inbox#访问令牌`。免账号，只读浏览邮件（分页）、安全隔离渲染 HTML 正文、打开邮件中的链接、下载授权附件。
- 后端公开只读 API 使用 `Authorization: Bearer <令牌>`；数据库保存 SHA-256 校验哈希和 AES-GCM 加密令牌（依赖 `LANQIN_EXTERNAL_IMAP_SECRET_KEY`）。仅登录且拥有邮箱的用户可通过专用接口再次复制完整链接；旧版哈希专用记录可重置令牌，原链接立即失效。
- 收信的 `messages` 必须经 `external_imap_messages` 映射校验，不允许跨外部账号访问。

## 开发期部署前检查

1. **先在隔离环境运行 Go 测试**：`cd apps/api && gofmt -l internal/app/external_imap_sharing*.go && go test ./internal/app -run TestExternalIMAPShare -count=1`，然后运行 `go test ./...`。
2. 只使用独立测试数据库，**不要覆盖当前运行的 `/opt/lanqin-test/data/lanqin.db`**。
3. 浏览器先用现有 SSH 隧道进行测试，再接 Cloudflare Tunnel 的 HTTPS；开发版使用 `http://localhost:18089`，通过 SSH 隧道访问的本机链接不能直接发给其他设备或车友。
4. 准备两组 seek.li IMAP 邮箱，各同步一封不同的邮件，验证 A 的链接无法请求 B 的邮件 ID 或附件 ID。
5. 测试有效期、手动撤销、无效令牌、HTML 脚本、外部 HTTPS 验证链接、手机浏览器、远程图片默认禁止加载。
6. 将 `LANQIN_PUBLIC_BASE_URL` 设置为最终 HTTPS 域名，关闭生产环境的 `LANQIN_ALLOW_INSECURE_HTTP`。不映射邮件协议公网端口。
7. 备份脚本需新增/核对 `external_imap_share_links` 表的保存与恢复；不要假定原有迁移工具自动覆盖新表。

## 注意

- 当前管理页面还是独立地址，没有添加到上游 React 侧边菜单。
- 只读指 **不能通过本系统修改/发送/删除 seek.li 邮件**；邮件中的授权链接本身可能允许在第三方服务上执行操作。
- `visibilityMinutes=0` 表示显示所有已同步邮件，**不会删除** seek.li 原邮件；共享链接到期后仅停止访问。
- 当前车友页面在可见状态下每 5 秒检查一次邮件，数据不变时不重新渲染列表；具体 IMAP 拉取频率由 `LANQIN_EXTERNAL_IMAP_SYNC_SECONDS` 决定（OCI 开发版建议设为 20 秒）。外部 IMAP 同步优先处理 INBOX，随后处理其他文件夹；多个账号仍串行检查，不能保证严格的 25 秒内到达。
- 原项目仍遵循 GPL-3.0 许可证。
