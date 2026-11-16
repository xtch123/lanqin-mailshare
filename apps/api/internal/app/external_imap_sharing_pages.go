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
<html lang="zh-CN"><head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta name="color-scheme" content="light"><title>邮箱共享管理 · LanQin</title>
<style>
:root{font-family:Inter,-apple-system,BlinkMacSystemFont,"Segoe UI","Microsoft YaHei",sans-serif;color:#20344d;background:#f5f7fb;font-synthesis:none}
*{box-sizing:border-box}body{margin:0;min-height:100vh}button,input,select{font:inherit}
button{cursor:pointer}button:focus-visible,input:focus-visible,select:focus-visible{outline:3px solid #91b7f5;outline-offset:2px}
.topbar{background:white;border-bottom:1px solid #e8edf4}.top-inner{max-width:1150px;margin:auto;padding:17px 24px;display:flex;align-items:center;justify-content:space-between;gap:16px}
.brand{display:flex;align-items:center;gap:11px;font-size:14px;font-weight:760}.logo{width:36px;height:36px;border-radius:10px;background:#245fa9;color:white;display:grid;place-items:center;font-size:19px}
.top-tag{font-size:11px;border:1px solid #dbe6f6;background:#f1f6fd;padding:7px 11px;border-radius:999px;color:#4775ab}
.workspace{max-width:1150px;margin:0 auto;padding:36px 24px 48px}
.eyebrow{font-size:11px;letter-spacing:.12em;color:#5a89c3;font-weight:850}
h1{font-size:28px;letter-spacing:-.04em;color:#162d49;margin:8px 0 9px}p{font-size:13px;line-height:1.8;color:#7e8da0;margin:0}
.intro{margin-bottom:27px;max-width:680px}
.layout{display:grid;grid-template-columns:minmax(0,1.6fr) minmax(310px,1fr);gap:18px;align-items:start}
.card{background:#fff;border:1px solid #e5ebf2;border-radius:15px;box-shadow:0 7px 24px #1b386208;min-width:0;overflow:hidden}
.card-head{padding:22px 25px 18px;border-bottom:1px solid #eef1f6}
.card-head h2{font-size:16px;letter-spacing:-.015em;margin:0 0 6px}.card-head p{font-size:12px}
.card-body{padding:18px 25px 24px}
.field{margin-bottom:17px}label{display:block;font-weight:660;font-size:12px;color:#50647c;margin:0 0 8px}
input,select{background:#fff;border:1px solid #dce4ee;border-radius:9px;padding:11px 12px;width:100%;font-size:13px;color:#263c56;min-height:43px}
input:hover,select:hover{border-color:#aebfd6}input::placeholder{color:#a3afbe}
.two-col{display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr);gap:14px}
.help{display:block;font-size:11px;color:#9aa7b7;margin-top:7px;line-height:1.5}
button{border:0;border-radius:9px;background:#285fa7;color:#fff;font-weight:720;font-size:13px;padding:11px 16px;min-height:41px;transition:background .15s}
button:hover{background:#204f92}button:disabled{opacity:.55;cursor:default}
button.secondary{background:#eef3f9;color:#415a77}button.secondary:hover{background:#e3ecf7}
button.danger{background:#fff3f3;color:#b74747;padding:9px 12px;font-size:12px;min-height:35px}
button.danger:hover{background:#ffeaea}
.actions{display:flex;gap:10px;flex-wrap:wrap;align-items:center;margin-top:4px}
#message{color:#bb4040;min-height:20px;margin:12px 0 0;font-size:12px}
#newLink{display:none;border:1px solid #d4e7f9;background:#f3f8fe;border-radius:12px;padding:16px;margin-top:14px}
#newLink b{font-size:12px;color:#315b88}.link{background:white;border:1px dashed #b8cde3;border-radius:9px;padding:12px;margin-top:10px;font-size:12px;line-height:1.6;word-break:break-all;color:#295c94}
#preview{font-size:12px;color:#285fa7;font-weight:700;text-decoration:none}#preview:hover{text-decoration:underline}
#shares{padding:0 20px 20px;min-height:74px;font-size:12px;color:#8b98ab}
.item{padding:16px 0;border-bottom:1px solid #edf0f5;display:flex;align-items:center;justify-content:space-between;gap:10px}
.item:last-child{border-bottom:0}.item b{display:block;font-size:13px;color:#263c56}
.item small{display:block;color:#91a0b0;line-height:1.6;font-size:11px;margin-top:6px;overflow-wrap:anywhere}
.note{font-size:12px;color:#8596a8;padding:17px 21px;background:#f8fafc;border-top:1px solid #edf1f6;line-height:1.7}
.footer{color:#9ba7b6;font-size:11px;text-align:center;margin:30px 0}
[hidden]{display:none!important}
@media(max-width:900px){.layout{grid-template-columns:minmax(0,1fr)}.workspace{max-width:760px}}
@media(max-width:580px){
 .top-inner{padding:13px 16px}.workspace{padding:23px 14px 32px}
 h1{font-size:24px}.intro{margin-bottom:20px}.layout{gap:14px}
 .card{border-radius:12px}.card-head{padding:19px 17px 15px}.card-body{padding:18px 17px 21px}
 .two-col{grid-template-columns:minmax(0,1fr);gap:0}
 .actions{display:grid;grid-template-columns:minmax(0,1fr) minmax(0,1fr)}
 .actions button{padding-left:8px;padding-right:8px}
 #newLink .actions{display:flex}#shares{padding:0 17px 15px}
}
</style></head><body>
<header class="topbar"><div class="top-inner">
 <div class="brand"><span class="logo" aria-hidden="true">✉</span><span>LanQin Mail</span></div>
 <span class="top-tag">外部邮箱 · 共享管理</span>
</div></header>
<main class="workspace">
 <div class="eyebrow">MAIL SHARING</div>
 <h1>外部邮箱共享管理</h1>
 <p class="intro">为每位车友创建一条独立的只读访问链接，随时控制有效期和可见邮件范围。车友无需登录 LanQin。</p>
 <div class="layout">
  <section class="card">
   <div class="card-head"><h2>创建共享链接</h2><p>请选择要共享的外部 IMAP 邮箱。</p></div>
   <div class="card-body">
    <div class="field"><label for="account">外部邮箱</label><select id="account"></select></div>
    <div class="field"><label for="label">车友备注（选填）</label><input id="label" maxlength="64" placeholder="例如：一号车 · 成员 A"></div>
    <div class="two-col">
     <div class="field">
      <label for="days">共享链接有效期</label><select id="days">
       <option value="30">30 天</option><option value="7">7 天</option><option value="1">1 天</option>
       <option value="0">永久有效，直到撤销</option><option value="custom">自定义天数…</option>
      </select>
      <div id="customDaysWrap" class="field" hidden><label for="customDays">自定义有效期（天）</label><input id="customDays" type="number" min="1" max="3650" step="1" placeholder="输入 1–3650"><small class="help">覆盖上方预设，仅在选「自定义」时生效</small></div>
     </div>
     <div class="field">
      <label for="visible">历史邮件可见范围</label><select id="visible">
       <option value="1440">最近 24 小时</option><option value="60">最近 1 小时</option>
       <option value="10080">最近 7 天</option><option value="0">全部已同步邮件</option>
       <option value="custom">自定义分钟数…</option>
      </select>
      <div id="customVisibleWrap" class="field" hidden><label for="customVisible">自定义可见范围（分钟）</label><input id="customVisible" type="number" min="0" max="525600" step="1" placeholder="0 为全部"><small class="help">例如 30 = 最近半小时</small></div>
     </div>
    </div>
    <div class="actions"><button id="create" type="button">生成专属链接</button><button id="refresh" type="button" class="secondary">刷新共享记录</button></div>
    <p id="message" role="alert"></p>
    <div id="newLink"><b>专属链接已生成 · 请立即复制保存</b><div id="linkText" class="link"></div>
     <div class="actions"><button id="copy" type="button">复制链接</button><a id="preview" href="#" target="_blank" rel="noopener noreferrer">预览车友页面 ↗</a></div>
    </div>
   </div>
  </section>
  <aside class="card">
   <div class="card-head"><h2>已有共享</h2><p>每条链接独立管理，撤销后立即失效。</p></div>
   <div id="shares">读取中…</div>
   <div class="note">建议每位车友单独生成链接，避免一人退出后需要更换所有人的访问凭证。</div>
  </aside>
 </div>
 <p class="footer">仅共享已同步的邮件，不提供发信、删除或更改邮件状态的功能。</p>
</main>
<script defer src="/api/me/external-imap-share-panel/app.js"></script></body></html>`
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
