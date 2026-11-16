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
const token=decodeURIComponent(location.hash.slice(1));
const list=document.getElementById("list"),detail=document.getElementById("detail");
const err=document.getElementById("error"),subtitle=document.getElementById("subtitle");
const mailShell=document.getElementById("mailShell"),mailCount=document.getElementById("mailCount");
let selectedMessageId="",currentAccountName="";
const mobile=()=>window.matchMedia("(max-width: 720px)").matches;

async function get(path){
 const r=await fetch("/api/public/shared-inbox"+path,{cache:"no-store",headers:{"Authorization":"Bearer "+token}});
 if(!r.ok)throw Error(r.status===401?"共享链接无效、已撤销或已过期":"请求失败（HTTP "+r.status+"）");
 return r.json();
}
function element(tag,text,cls){
 const e=document.createElement(tag);if(text!==undefined)e.textContent=text;if(cls)e.className=cls;return e;
}
function emptyState(heading,description){
 const wrap=element("div",undefined,"empty"),icon=element("div","✉","empty-icon");
 wrap.append(icon,element("h3",heading),element("p",description));return wrap;
}
function dateText(value){
 const date=new Date(value);
 return Number.isNaN(date.getTime())?"":date.toLocaleString("zh-CN",{month:"2-digit",day:"2-digit",hour:"2-digit",minute:"2-digit"});
}
function showMessage(){
 mailShell.classList.add("detail-open");
 if(mobile())window.scrollTo({top:0,behavior:"instant"});
}
document.getElementById("backToList").onclick=()=>{
 mailShell.classList.remove("detail-open");
 if(mobile())window.scrollTo({top:0,behavior:"instant"});
};
async function openMail(id){
 selectedMessageId=id;
 showMessage();
 detail.replaceChildren(emptyState("正在加载邮件","正在安全读取邮件正文…"));
 try{
  const m=await get("/messages/"+encodeURIComponent(id));
  // A second selection can finish first; never show the wrong message.
  if(selectedMessageId!==id)return;
  detail.replaceChildren();
  detail.append(element("h2",m.subject||"（无主题）"));
  detail.append(element("div",(m.fromName||m.from)+" · "+dateText(m.receivedAt),"meta"));
  if(m.bodyHtml){
   const frame=document.createElement("iframe");
   frame.setAttribute("sandbox","allow-popups allow-popups-to-escape-sandbox");
   frame.setAttribute("referrerpolicy","no-referrer");
   frame.setAttribute("title","邮件正文");
   // The server sanitizes the email HTML. This frame also blocks scripts,
   // images from remote hosts and form submissions.
   frame.srcdoc="<meta http-equiv=\"Content-Security-Policy\" content=\"default-src 'none'; style-src 'unsafe-inline'; img-src data:; form-action 'none'\"><base target=\"_blank\">"+m.bodyHtml;
   detail.append(frame);
  }else{
   detail.append(element("pre",m.bodyText||"（邮件没有正文）"));
  }
  if((m.attachments||[]).length){
   const box=element("section");box.append(element("h3","附件","attachment-title"));
   m.attachments.forEach(att=>{
    const button=element("button","↓ "+att.filename+" · "+Math.max(1,Math.round(att.sizeBytes/1024))+" KB","refresh attachment-btn");
    button.type="button";
    button.onclick=async()=>{
     try{
      const res=await fetch("/api/public/shared-inbox/messages/"+encodeURIComponent(m.id)+"/attachments/"+encodeURIComponent(att.id),{cache:"no-store",headers:{"Authorization":"Bearer "+token}});
      if(!res.ok)throw Error("附件下载失败");
      const url=URL.createObjectURL(await res.blob()),anchor=document.createElement("a");
      anchor.href=url;anchor.download=att.filename||"attachment";
      document.body.append(anchor);anchor.click();anchor.remove();
      setTimeout(()=>URL.revokeObjectURL(url),30000);
     }catch(e){err.textContent=e.message}
    };
    box.append(button);
   });
   detail.append(box);
  }
 }catch(e){
  if(selectedMessageId===id)detail.replaceChildren(emptyState("邮件暂不可用",e.message));
 }
}
async function loadPage(page,append){
 const data=await get("/messages?page="+page);
 currentAccountName=data.accountName||"外部邮箱";
 const messages=data.items||[];
 subtitle.textContent=currentAccountName+" · 只读访问";
 if(!append){
  list.replaceChildren();
  err.textContent="";
  if(!messages.length){
   mailCount.textContent="0 封";
   list.append(emptyState("当前范围内没有邮件","共享链接本身有效。可能是邮箱尚未同步，或邮件早于管理员设置的可见时间范围。"));
  }else mailCount.textContent="最新 "+messages.length+" 封";
 }
 messages.forEach(m=>{
  const b=element("button",undefined,"mail"),top=element("div",undefined,"mail-top");
  b.type="button";
  top.append(element("b",m.fromName||m.from||"未知发件人"),element("span",dateText(m.receivedAt),"mail-time"));
  b.append(top,element("span",m.subject||"（无主题）","mail-subject"),element("span",m.snippet||"","mail-snippet"));
  b.title=m.subject||"（无主题）";
  if(m.id===selectedMessageId)b.classList.add("selected");
  b.onclick=()=>{
   list.querySelectorAll(".mail").forEach(x=>x.classList.remove("selected"));
   b.classList.add("selected");
   openMail(m.id);
  };
  list.append(b);
  if(page===0&&!append&&!selectedMessageId&&!mobile()&&messages.indexOf(m)===0)b.click();
 });
 if(data.nextPage!==null&&data.nextPage!==undefined){
  const more=element("button","加载更早的邮件","mail");
  more.type="button";
  more.onclick=async()=>{
   more.disabled=true;
   try{more.remove();await loadPage(data.nextPage,true)}
   catch(e){err.textContent=e.message}
  };
  list.append(more);
 }
}
async function refresh(){
 if(!token){
  subtitle.textContent="未提供访问凭证";
  err.textContent="共享链接缺少访问令牌，请使用管理员提供的完整链接。";
  list.replaceChildren(emptyState("无法访问","请重新打开完整的共享链接。"));
  return;
 }
 try{await loadPage(0,false)}
 catch(e){
  err.textContent=e.message;
  subtitle.textContent="无法访问共享收件箱";
  list.replaceChildren(emptyState("无法读取邮件",e.message));
 }
}
document.getElementById("refresh").onclick=refresh;
refresh();
setInterval(refresh,30000);`
