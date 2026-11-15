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
<html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>共享邮箱 · 只读</title>
<style>
*{box-sizing:border-box}body{margin:0;background:#f7f8fa;color:#1b2430;font:15px system-ui,-apple-system,"Microsoft YaHei",sans-serif}
header{background:white;padding:19px 24px;border-bottom:1px solid #e0e5ed}h1{font-size:20px;margin:0}#subtitle{font-size:13px;color:#697789;margin-top:5px}
main{max-width:960px;margin:24px auto;padding:0 15px;display:grid;grid-template-columns:310px minmax(0,1fr);gap:18px}
.panel{background:white;border:1px solid #e0e5ed;border-radius:11px;min-height:480px;overflow:hidden}
#list{max-height:72vh;overflow-y:auto}.mail{width:100%;display:block;border:0;border-bottom:1px solid #e9edf2;background:white;padding:16px;text-align:left;cursor:pointer;color:#17243a}
.mail:hover,.mail.selected{background:#f1f5fa}.mail b{display:block;font-size:14px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
.mail small{display:block;color:#64748b;margin-top:7px}.mail .excerpt{white-space:nowrap;overflow:hidden;text-overflow:ellipsis}
#detail{padding:20px}#detail h2{font-size:21px;margin:0 0 12px;overflow-wrap:anywhere}
#meta{color:#708092;font-size:13px;margin:0 0 16px;border-bottom:1px solid #e9edf2;padding-bottom:16px}
iframe{width:100%;height:52vh;border:0;background:white}pre{white-space:pre-wrap;overflow-wrap:anywhere;font:inherit}
button.refresh{border:1px solid #d6e0ec;border-radius:6px;background:white;padding:8px 13px;cursor:pointer}
#error{color:#b52d32;padding:0 24px}
@media(max-width:720px){main{display:block;margin:14px auto}.panel{min-height:auto;margin-bottom:12px}#list{max-height:250px}#detail{min-height:260px}iframe{height:60vh}}
</style></head><body>
<header><h1>共享邮箱 <span style="font-size:12px;color:#64748b">只读</span></h1><div id="subtitle">正在检查访问链接……</div><button class="refresh" id="refresh">刷新邮件</button></header>
<p id="error" role="alert"></p>
<main><div class="panel" id="list"></div><div class="panel" id="detail"><h2>选择一封邮件</h2><p>可以打开正文中的验证链接。此页面无法发送或删除邮件。</p></div></main>
<script defer src="/api/public/shared-inbox/app.js"></script></body></html>`
