package app

import (
	"net/http"
)

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
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; frame-src 'self'; img-src data:; base-uri 'none'; form-action 'none'")
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
</main><script>
"use strict";
const $=id=>document.getElementById(id);
const api=async(path,opt={})=>{const r=await fetch("/api"+path,Object.assign({credentials:"same-origin",cache:"no-store"},opt));if(!r.ok)throw Error("HTTP "+r.status+"：请检查是否仍在登录状态");return r.json()};
let lastLink="";
function showError(e){$("message").textContent=e.message||String(e)}
async function loadAccounts(){
  const data=await api("/me/external-imap-accounts");
  $("account").replaceChildren();
  (data.items||[]).filter(a=>a.enabled&&a.storageMode==="local").forEach(a=>{
    const o=document.createElement("option");o.value=a.id;o.textContent=(a.name||a.username)+" ("+a.username+")";$("account").append(o)
  });
  if(!$("account").options.length)$("message").textContent="暂无已启用且同步到本地的外部邮箱";
}
async function loadShares(){
 const result=await api("/me/external-imap-shares");
 $("shares").replaceChildren();
 if(!(result.items||[]).length){$("shares").textContent="尚无共享链接";return}
 result.items.forEach(s=>{
  const row=document.createElement("div");row.className="item";
  const d=document.createElement("div"),t=document.createElement("b"),sub=document.createElement("small");
  t.textContent=s.label||s.accountId;
  const status=s.revokedAt?"已撤销":(s.expiresAt&&new Date(s.expiresAt)<new Date()?"已过期":"有效");
  sub.textContent=status+" · 历史范围 "+(s.visibilityMinutes===0?"全部邮件":s.visibilityMinutes+" 分钟")+" · 过期 "+(s.expiresAt?new Date(s.expiresAt).toLocaleString():"永久")+" · "+s.accountId;
  d.append(t,sub);row.append(d);
  if(status==="有效"){const b=document.createElement("button");b.className="danger";b.textContent="撤销";
    b.onclick=async()=>{if(!confirm("确认撤销？该车友的链接将立即失效。"))return;try{await api("/me/external-imap-shares/"+encodeURIComponent(s.id),{method:"DELETE"});await loadShares()}catch(e){showError(e)}};row.append(b)}
  $("shares").append(row)
 });
}
$("create").onclick=async()=>{
 $("message").textContent="";$("newLink").style.display="none";
 const accountId=$("account").value;if(!accountId){$("message").textContent="请先添加一个外部邮箱";return}
 const days=$("customDays").value.trim()===""?Number($("days").value):Number($("customDays").value);
 const visible=$("customVisible").value.trim()===""?Number($("visible").value):Number($("customVisible").value);
 if(!Number.isInteger(days)||days<0||days>3650||!Number.isInteger(visible)||visible<0||visible>525600){
  $("message").textContent="自定义参数超出允许范围";return
 }
 const payload={accountId,label:$("label").value,visibilityMinutes:visible};
 if(days)payload.expiresAt=new Date(Date.now()+days*86400000).toISOString();
 $("create").disabled=true;
 try{const v=await api("/me/external-imap-shares",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload)});
    lastLink=v.url;$("linkText").textContent=lastLink;$("preview").href=lastLink;$("newLink").style.display="block";
    await loadShares()}
 catch(e){showError(e)}finally{$("create").disabled=false}
};
$("refresh").onclick=()=>loadShares().catch(showError);
$("copy").onclick=async()=>{try{await navigator.clipboard.writeText(lastLink);$("copy").textContent="已复制"}catch(e){showError(e)}};
loadAccounts().then(loadShares).catch(showError);
</script></body></html>`

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
<script>
"use strict";
const token=decodeURIComponent(location.hash.slice(1)),list=document.getElementById("list"),
 detail=document.getElementById("detail"),err=document.getElementById("error"),subtitle=document.getElementById("subtitle");
async function get(path){
 const r=await fetch("/api/public/shared-inbox"+path,{cache:"no-store",headers:{"Authorization":"Bearer "+token}});
 if(!r.ok)throw Error(r.status===401?"共享链接无效、已撤销或已过期":"请求失败 HTTP "+r.status);
 return r.json();
}
function clear(node){node.replaceChildren()}
function make(tag,text){const n=document.createElement(tag);n.textContent=text||"";return n}
async function openMail(id){
 clear(detail);detail.append(make("p","正在加载邮件……"));
 try{
  const m=await get("/messages/"+encodeURIComponent(id));
  clear(detail);detail.append(make("h2",m.subject),make("div",(m.fromName||m.from)+" · "+new Date(m.receivedAt).toLocaleString()));
  const frame=document.createElement("iframe");
  frame.setAttribute("sandbox","allow-popups allow-popups-to-escape-sandbox");
  frame.setAttribute("referrerpolicy","no-referrer");
  frame.setAttribute("title","邮件正文");
  const html=m.bodyHtml||"";
  if(html){
   frame.srcdoc="<meta http-equiv=\"Content-Security-Policy\" content=\"default-src 'none'; style-src 'unsafe-inline'; img-src data:; form-action 'none'\"><base target=\"_blank\">"+html;
   detail.append(frame)
  }else{detail.append(make("pre",m.bodyText||"（邮件没有正文）"))}
  if((m.attachments||[]).length){
   const box=make("section");box.append(make("h3","附件"));
   (m.attachments||[]).forEach(att=>{
    const button=make("button","下载 "+att.filename+" ("+Math.round(att.sizeBytes/1024)+" KB)");
    button.className="refresh";button.style.margin="5px";
    button.onclick=async()=>{
     try{
      const res=await fetch("/api/public/shared-inbox/messages/"+encodeURIComponent(m.id)+"/attachments/"+encodeURIComponent(att.id),{
       cache:"no-store",headers:{"Authorization":"Bearer "+token}
      });
      if(!res.ok)throw Error("附件下载失败");
      const u=URL.createObjectURL(await res.blob()),link=document.createElement("a");
      link.href=u;link.download=att.filename||"attachment";document.body.append(link);link.click();link.remove();
      setTimeout(()=>URL.revokeObjectURL(u),30000)
     }catch(e){err.textContent=e.message}
    };
    box.append(button)
   });
   detail.append(box)
  }
 }catch(e){clear(detail);detail.append(make("p",e.message))}
}
async function refresh(){
 if(!token){subtitle.textContent="共享链接缺少访问令牌";err.textContent="请使用管理员提供的完整链接";return}
 try{
  const data=await get("/messages");subtitle.textContent=(data.accountName||"外部邮箱")+" · 最近 "+data.items.length+" 封邮件";
  clear(list);err.textContent="";
  if(!data.items.length){list.append(make("p","当前可见范围内没有邮件"));return}
  data.items.forEach((m,i)=>{
   const b=make("button");b.className="mail";b.append(make("b",m.subject),make("small",(m.fromName||m.from)+" · "+new Date(m.receivedAt).toLocaleString()),make("small",m.snippet));
   b.onclick=()=>{list.querySelectorAll(".mail").forEach(x=>x.classList.remove("selected"));b.classList.add("selected");openMail(m.id)};
   list.append(b);
   if(i===0&&detail.querySelector("h2")?.textContent==="选择一封邮件"){b.click()}
  })
 }catch(e){err.textContent=e.message;subtitle.textContent="无法读取共享邮箱"}
}
document.getElementById("refresh").onclick=refresh;
refresh();setInterval(refresh,30000);
</script></body></html>`
