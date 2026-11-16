package app

import "net/http"

// First-cut pages are self-contained so they can be tested without changing
// upstream React navigation. They do NOT expose the admin session to viewers.
func (a *App) handleExternalIMAPShareAdminPage(w http.ResponseWriter, r *http.Request) {
	shareHTMLHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(externalIMAPShareAdminHTML))
}

func (a *App) handleExternalIMAPSharePage(w http.ResponseWriter, r *http.Request) {
	shareHTMLHeaders(w)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(externalIMAPShareViewerHTML))
}

func shareHTMLHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'unsafe-inline'; connect-src 'self'; frame-src 'self'; img-src data:; base-uri 'none'; form-action 'none'")
}

const externalIMAPShareAdminHTML = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>外部邮箱共享管理 · LanQin</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#f6f7f9;color:#18202e;font:15px system-ui,-apple-system,"Microsoft YaHei",sans-serif}
main{max-width:890px;margin:36px auto;padding:0 18px}header{margin-bottom:22px}h1{font-size:26px;margin:0 0 7px}p{color:#687485;line-height:1.6}
section{background:white;border:1px solid #e1e5eb;border-radius:12px;padding:22px;margin-bottom:20px}
label{display:block;font-size:13px;margin:15px 0 6px;font-weight:600}
input,select{padding:11px;border:1px solid #cdd4df;border-radius:7px;width:100%;font-size:14px;background:white}
button{border:0;background:#18202e;color:white;padding:11px 17px;border-radius:7px;cursor:pointer;font-weight:600}
button:disabled{opacity:.5;cursor:default}button.secondary{background:#edf1f5;color:#222}button.danger{background:#fff2f1;color:#bd3434}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:14px}.actions{display:flex;gap:9px;margin-top:17px;flex-wrap:wrap}
.item{border-top:1px solid #e5e8ed;padding:16px 0;display:flex;justify-content:space-between;align-items:center;gap:12px}
small{color:#6d7a89;display:block;margin-top:5px;word-break:break-all}.link{border:1px dashed #a0b2c8;background:#f3f7fb;border-radius:8px;padding:13px;word-break:break-all;margin-top:12px}
#message{color:#c34137;min-height:24px}#newLink{display:none}a{color:#2457a6}
@media(max-width:620px){.grid{grid-template-columns:1fr}.item{align-items:flex-start;flex-direction:column}}
</style></head><body><main>
<header><h1>外部邮箱共享管理</h1><p>一个 seek.li 邮箱对应一个拼车账号。每个车友单独生成只读链接，支持随时撤销、设置有效期与历史邮件范围。</p></header>
<section><h2>创建共享链接</h2>
<label for="account">外部 IMAP 邮箱</label><select id="account"></select>
<div class="grid"><div><label for="label">车友备注（选填）</label><input id="label" maxlength="64" placeholder="例如：一号车 · 成员 A"></div>
<div><label for="days">共享链接有效期</label><select id="days"><option value="7">7 天</option><option value="1">1 天</option><option value="30">30 天</option><option value="0">永久有效，直到手动撤销</option></select></div></div>
<label for="visible">可查看多久以前的邮件</label><select id="visible"><option value="1440">最近 24 小时</option><option value="60">最近 1 小时</option><option value="10080">最近 7 天</option><option value="0">全部已同步邮件</option></select>
<div class="grid"><div><label for="customDays">自定义链接有效期（天，可选）</label><input id="customDays" type="number" min="1" max="3650" step="1" placeholder="留空使用上方预设"></div>
<div><label for="customVisible">自定义可见范围（分钟，可选）</label><input id="customVisible" type="number" min="0" max="525600" step="1" placeholder="留空使用上方预设；0 为全部"></div></div>
<div class="actions"><button id="create">生成专属链接</button><button id="refresh" class="secondary">刷新共享记录</button></div>
<p id="message"></p><div id="newLink"><b>请立即复制保存，下次不会再显示完整链接：</b><div id="linkText" class="link"></div><div class="actions"><button id="copy">复制链接</button><a id="preview" href="#" target="_blank" rel="noopener noreferrer">预览共享页面</a></div></div>
</section><section><h2>已有共享</h2><div id="shares">读取中……</div></section>
<p>管理页面需要 LanQin 管理员登录。车友使用专用链接，无需注册。本页为独立开发预览，尚未集成原版侧栏。</p>
</main><script defer src="/api/me/external-imap-share-panel/app.js"></script></body></html>`

const externalIMAPShareViewerHTML = `<!doctype html>
<html lang="zh-CN"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light"><title>共享收件箱 · LanQin</title>
<style>
:root{font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI","Microsoft YaHei",sans-serif;color:#1b304b;background:#f5f7fb;font-synthesis:none}
*{box-sizing:border-box}body{margin:0;min-height:100vh}button{font:inherit;cursor:pointer}button:focus-visible{outline:3px solid #91b7f5;outline-offset:2px}
.topbar{background:#fff;border-bottom:1px solid #e8edf4}
.topbar-inner{max-width:1210px;margin:auto;padding:17px 26px;display:flex;justify-content:space-between;align-items:center;gap:20px}
.brand{display:flex;align-items:center;gap:12px}
.logo{width:37px;height:37px;background:#245fa9;color:white;display:grid;place-items:center;border-radius:11px;font-weight:800;font-size:20px;box-shadow:0 6px 16px #245fa924}
.brand strong{font-size:15px;letter-spacing:.01em}.brand span{display:block;font-size:11px;color:#7c8ea4;margin-top:3px}
.readonly{display:inline-flex;align-items:center;gap:7px;border:1px solid #dae8fa;background:#eff6ff;border-radius:999px;color:#3166ab;font-size:12px;font-weight:650;padding:7px 12px}
.readonly:before{content:"";width:6px;height:6px;border-radius:100%;background:#6eabd7}
.workspace{max-width:1210px;padding:34px 26px 50px;margin:auto}
.hero{display:flex;align-items:flex-end;justify-content:space-between;gap:20px;margin-bottom:25px}
.eyebrow{color:#5282bd;font-size:11px;font-weight:800;letter-spacing:.13em}
.hero h1{font-size:29px;font-weight:750;letter-spacing:-.035em;margin:7px 0 8px;color:#152a45}
#subtitle{font-size:13px;color:#79899e;line-height:1.6}
.refresh{display:inline-flex;gap:7px;align-items:center;justify-content:center;min-height:40px;border-radius:9px;background:#fff;border:1px solid #dce5f0;padding:8px 15px;color:#345a88;font-weight:600;font-size:13px;box-shadow:0 2px 4px #162c4410}
.refresh:hover{background:#f1f6fd}
#error{border:1px solid #f1c6c6;background:#fff4f4;color:#a72d32;padding:13px 16px;border-radius:10px;margin:0 0 17px;font-size:13px}
#error:empty{display:none}
.mail-shell{display:grid;grid-template-columns:minmax(290px,360px) minmax(0,1fr);gap:18px;align-items:stretch;min-height:min(68vh,700px)}
.panel{background:#fff;border:1px solid #e5eaf1;border-radius:15px;box-shadow:0 7px 24px #1b386208;overflow:hidden;min-width:0}
.pane-head{height:63px;display:flex;align-items:center;justify-content:space-between;gap:12px;padding:0 21px;border-bottom:1px solid #eaf0f6}
.pane-title{font-size:14px;font-weight:750;color:#243a58}.pane-desc{font-size:12px;color:#91a0b0}
.list-content{max-height:min(61vh,640px);overflow-y:auto;scrollbar-width:thin}
.mail{width:100%;display:block;border:0;border-bottom:1px solid #eef1f6;background:#fff;padding:16px 19px;text-align:left;color:#263b53;position:relative}
.mail:hover,.mail.selected{background:#f1f6fe}
.mail.selected:before{content:"";position:absolute;left:0;top:0;bottom:0;width:3px;background:#3479c8}
.mail-top{display:flex;justify-content:space-between;align-items:center;gap:8px}
.mail b{font-size:13px;font-weight:750;display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.mail-time{font-size:11px;color:#91a0b0;flex-shrink:0}
.mail-subject{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-size:13px;font-weight:640;margin:8px 0 5px}
.mail-snippet{display:block;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;font-size:12px;color:#8493a6}
.reading-head{height:63px;display:flex;align-items:center;gap:12px;padding:0 22px;border-bottom:1px solid #eaf0f6}
.reading-title{font-size:13px;font-weight:720;color:#3b5270}
.reading-tag{margin-left:auto;font-size:11px;color:#8899ae}
.mobile-back{display:none;border:0;background:#eef5ff;color:#3265a3;padding:8px 12px;border-radius:9px;font-weight:650;font-size:13px}
.detail-content{padding:27px 29px;min-height:400px;overflow-wrap:anywhere}
.detail-content h2{font-size:21px;letter-spacing:-.03em;line-height:1.4;color:#1a304b;margin:0 0 16px}
.meta{font-size:12px;color:#76889f;border-bottom:1px solid #edf0f5;padding:0 0 18px;margin-bottom:20px;line-height:1.7}
iframe{width:100%;height:min(51vh,550px);border:0;background:#fff}pre{font:inherit;font-size:14px;white-space:pre-wrap;overflow-wrap:anywhere;line-height:1.85}
.empty{padding:56px 24px;text-align:center;display:flex;align-items:center;flex-direction:column;justify-content:center;min-height:280px}
.empty-icon{width:62px;height:62px;border-radius:20px;background:#eef5ff;color:#4d85c7;display:grid;place-items:center;font-size:27px;margin-bottom:16px}
.empty h3{font-size:15px;margin:0 0 8px}.empty p{font-size:12px;color:#8392a5;line-height:1.8;max-width:280px;margin:0}
.attachment-title{font-size:13px;margin:24px 0 8px}.attachment-btn{margin:5px 7px 5px 0}
.footnote{text-align:center;font-size:11px;color:#9aa8b9;margin:23px 0}
@media(max-width:900px){.workspace{padding:26px 18px 40px}.topbar-inner{padding:15px 18px}.mail-shell{grid-template-columns:minmax(245px,310px) minmax(0,1fr);gap:12px}.detail-content{padding:22px}}
@media(max-width:720px){
 .topbar-inner{padding:12px 16px}.brand strong{font-size:14px}.logo{width:33px;height:33px;border-radius:9px}
 .workspace{padding:23px 14px 24px}.hero{align-items:flex-start;margin-bottom:18px;gap:14px}.hero h1{font-size:24px}
 #subtitle{font-size:12px}.refresh{padding:9px 11px}.refresh span{display:none}
 .mail-shell{display:block;min-height:calc(100dvh - 238px)}.panel{min-height:calc(100dvh - 238px);border-radius:12px}
 .reading-panel{display:none}.mail-shell.detail-open .inbox-panel{display:none}.mail-shell.detail-open .reading-panel{display:block}
 .mobile-back{display:inline-flex}.reading-head{padding:0 15px;height:57px}
 .pane-head{height:57px;padding:0 16px}.list-content{max-height:none;overflow:visible}
 .mail{padding:17px 16px}.detail-content{padding:19px 17px}.detail-content h2{font-size:20px}
 iframe{height:60dvh}.empty{min-height:300px;padding:44px 20px}
}
@media(max-width:380px){.readonly{padding:6px 8px;font-size:11px}.brand span{font-size:10px}.hero h1{font-size:22px}.refresh{min-width:40px}}
</style></head><body>
<header class="topbar"><div class="topbar-inner">
 <div class="brand"><div class="logo" aria-hidden="true">✉</div><div><strong>LanQin Mail</strong><span>安全邮件共享</span></div></div>
 <span class="readonly">只读访问</span>
</div></header>
<main class="workspace">
 <div class="hero"><div><div class="eyebrow">SHARED INBOX</div><h1>共享收件箱</h1><div id="subtitle" aria-live="polite">正在验证共享链接…</div></div>
 <button class="refresh" id="refresh" type="button" aria-label="刷新邮件">↻ <span>刷新邮件</span></button></div>
 <div id="error" role="alert"></div>
 <div class="mail-shell" id="mailShell">
  <section class="panel inbox-panel" aria-label="邮件列表">
   <div class="pane-head"><span class="pane-title">收件箱</span><span class="pane-desc" id="mailCount">—</span></div>
   <div id="list" class="list-content"><div class="empty"><div class="empty-icon">✉</div><h3>正在加载邮件</h3><p>请稍候…</p></div></div>
  </section>
  <section class="panel reading-panel" aria-label="邮件详情">
   <div class="reading-head"><button class="mobile-back" id="backToList" type="button">← 返回列表</button><span class="reading-title">邮件内容</span><span class="reading-tag">只读模式</span></div>
   <div id="detail" class="detail-content"><div class="empty"><div class="empty-icon">✉</div><h3>选择一封邮件</h3><p>点击左侧列表即可查看邮件正文与验证链接，不支持发送和删除。</p></div></div>
  </section>
 </div>
 <p class="footnote">由 LanQin 提供安全的只读邮件访问 · 共享访问可能随时被撤销</p>
</main>
<script defer src="/api/public/shared-inbox/app.js"></script></body></html>`
