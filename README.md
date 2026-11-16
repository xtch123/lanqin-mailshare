# LanQin MailShare Lite

**简体中文** | [English](./README.en.md)

> 基于 [LanQin Email](https://github.com/LanQin996/LanQin-Email) 的非官方精简分支，专注于**外部 IMAP 邮箱聚合、邮件浏览与只读共享**。此仓库由社区独立维护，并非原作者的官方 Lite 版本。

![Go](https://img.shields.io/badge/Go-1.27%2B-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-18.3-61DAFB?logo=react)
![SQLite](https://img.shields.io/badge/SQLite-003B57?logo=sqlite)
![Docker](https://img.shields.io/badge/Docker-supported-2496ED?logo=docker)
![License](https://img.shields.io/badge/License-GPL--3.0--only-blue)

## 项目介绍

MailShare Lite 以个人自托管为目标，保留收件、管理和共享所需的组件，减少完整邮件服务器的运行开销。适合连接已有的 IMAP 邮箱并提供受控的邮件只读访问，**不是用于自行收发公网邮件的 SMTP 邮件服务器**。

### 主要功能

- **外部 IMAP 邮箱**：连接现有邮箱，浏览邮件并按配置同步至本地 SQLite 缓存。
- **只读邮件共享**：通过受控共享访问邮箱内容；可管理、撤销共享，设置可用期限。访客不获得邮件发送权限。
- **简化的管理界面**：统一的邮箱、共享和账户设置入口，适配桌面与移动端。
- **三种界面语言**：简体中文、繁体中文、English；用户可自行切换。应用会记住语言偏好。
- **轻量容器**：Go API、React 前端、Nginx 与 SQLite；Lite 镜像不包含 Postfix、Dovecot 和 Rspamd 守护进程。

### 与原版的区别

| 项目 | MailShare Lite | 原版 LanQin Email |
| --- | --- | --- |
| 主要用途 | IMAP 邮件聚合与只读共享 | 完整 Webmail / 自托管邮件服务 |
| 外部 IMAP 接入 | 支持 | 支持 |
| 共享与只读浏览 | 支持 | 以完整邮件管理功能为主 |
| SMTP 发信、完整邮件投递栈 | Lite 模式禁用 | 支持 |
| 运行组件 | Go + React + Nginx + SQLite | 包含完整邮件服务栈 |
| 语言 | 简体中文、繁体中文、英文 | 以原作者版本为准 |

> Lite 模式虽然关闭了发信等后台能力，但上游部分路由实现仍保留在代码中。**未经额外安全审计，不要直接把本服务暴露到公网。**

## 代码与分支

| 分支 | 用途 |
| --- | --- |
| [`main`](../../tree/main) | Fork 主分支，包含 Lite 相关代码与文档 |
| [`lite`](../../tree/lite) | 实际使用的 Lite 分支 |
| [`dev`](../../tree/dev) | 开发、修复及集成测试 |
| [`original-archive`](../../tree/original-archive) | 原作者代码快照，便于追溯 |
 
上游原项目：[LanQin996/LanQin-Email](https://github.com/LanQin996/LanQin-Email)。这个 Fork 以必要的 **Bug 修复、安全修复及兼容性维护** 为主，不承诺同步原版所有功能。

## 构建与部署

仓库中提供单独的 Lite Dockerfile、Compose 配置和操作文档：

- [Lite 部署文件](./deploy/mailshare-lite/)
- [Lite 说明与安全注意事项](./deploy/mailshare-lite/README.md)

从源码构建镜像（在仓库根目录执行）：

```bash
docker build -f deploy/mailshare-lite/Dockerfile -t lanqin-mailshare:lite-arm64 .
```

其中 `lite-arm64` 是示例镜像标签；实际平台由 Docker 构建环境决定，其他架构请自行调整标签或构建参数。

使用前需自行准备 `.env`、持久化数据目录和 Compose 部署目录；不要把开发环境的真实邮箱密码、数据库或密钥提交到仓库。现有示例 Compose 仅将 HTTP 映射到：

```text
127.0.0.1:18090 -> 容器 80
```

可在服务器本机访问，或使用 SSH 本地端口转发。**不应直接开放公网端口，也不应未经审计就经 Cloudflare Tunnel 公网发布。**

## 数据与安全

- 邮件内容、SQLite 数据库与外部邮箱凭据需要妥善保护。
- 备份前应对活动中的 SQLite 数据库生成一致性快照；备份数据须加密保存并定期验证可恢复性。
- 撤销共享后应检查访问已失效；升级前应在测试环境复核账号隔离、共享权限及 API 路由。
- 本 Fork 的 Lite 版本以自用维护为主，**不构成已完成公网部署安全审计的声明**。

## 致谢与许可

感谢 [LanQin996](https://github.com/LanQin996) 及原项目贡献者。本文介绍的是此 Fork 的精简定制部分；原版完整邮件服务的说明和部署方式请参阅[上游仓库](https://github.com/LanQin996/LanQin-Email)。

代码许可证沿用仓库中的 [GPL-3.0-only](./LICENSE)；历史提交的授权以其当时适用的许可证为准。
