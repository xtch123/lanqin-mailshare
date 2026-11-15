package app

import "net/http"

// JavaScript is served from same-origin endpoints because the upstream Nginx
// content security policy blocks inline scripts on all application routes.
func (a *App) handleExternalIMAPShareAdminJS(w http.ResponseWriter, r *http.Request) {
	shareScriptHeaders(w)
	_, _ = w.Write([]byte(externalIMAPShareAdminJS))
}

func (a *App) handleExternalIMAPShareViewerJS(w http.ResponseWriter, r *http.Request) {
	shareScriptHeaders(w)
	_, _ = w.Write([]byte(externalIMAPShareViewerJS))
}

func shareScriptHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

const externalIMAPShareAdminJS = `"use strict";
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
loadAccounts().then(loadShares).catch(showError);`

const externalIMAPShareViewerJS = `"use strict";
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
let currentPage=0;
async function loadPage(page,append){
 const data=await get("/messages?page="+page);
 subtitle.textContent=(data.accountName||"外部邮箱")+" · 已加载邮件";
 if(!append){clear(list);err.textContent=""}
 if(!append&&!(data.items||[]).length){list.append(make("p","当前可见范围内没有邮件"))}
 (data.items||[]).forEach((m,i)=>{
  const b=make("button");b.className="mail";b.append(make("b",m.subject),make("small",(m.fromName||m.from)+" · "+new Date(m.receivedAt).toLocaleString()),make("small",m.snippet));
  b.onclick=()=>{list.querySelectorAll(".mail").forEach(x=>x.classList.remove("selected"));b.classList.add("selected");openMail(m.id)};
  list.append(b);
  if(page===0&&i===0&&detail.querySelector("h2")?.textContent==="选择一封邮件"){b.click()}
 });
 if(data.nextPage!==null&&data.nextPage!==undefined){
  const more=make("button","加载更多");more.className="mail";
  more.onclick=async()=>{more.remove();try{await loadPage(data.nextPage,true)}catch(e){err.textContent=e.message}};
  list.append(more)
 }
 currentPage=page;
}
async function refresh(){
 if(!token){subtitle.textContent="共享链接缺少访问令牌";err.textContent="请使用管理员提供的完整链接";return}
 try{await loadPage(0,false)}catch(e){err.textContent=e.message;subtitle.textContent="无法读取共享邮箱"}
}
document.getElementById("refresh").onclick=refresh;
refresh();setInterval(refresh,30000);`
