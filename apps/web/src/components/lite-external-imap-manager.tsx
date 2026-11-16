import * as React from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { CheckCircle2, Inbox, LoaderCircle, Plus, RefreshCcw, ServerCog, Trash2 } from "lucide-react"
import { api, type ExternalImapAccount, type ExternalImapAccountPayload, type ExternalImapStorageMode, type ExternalImapTlsMode } from "@/lib/api"
import { errorMessage } from "@/lib/ui-errors"
import { formatDateTime } from "@/lib/utils"
import { uiText, useLanguage } from "@/lib/language"
import { LitePageLayout, LiteSection } from "@/components/lite-page-layout"
import { useToast } from "@/hooks/use-toast"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Checkbox } from "@/components/ui/checkbox"
import { Label } from "@/components/ui/label"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"

type Props = {
  accounts: ExternalImapAccount[]
  loading: boolean
  mailboxId: string
  onOpenInbox: (account: ExternalImapAccount) => void
  onOpenShares: () => void
}

type EditorProps = {
  account: ExternalImapAccount | null
  mailboxId: string
  busy: boolean
  onClose: () => void
  onSave: (payload: ExternalImapAccountPayload) => Promise<void>
}

function ExternalAccountEditor({ account, mailboxId, busy, onClose, onSave }: EditorProps) {
  useLanguage()
  const [tlsMode, setTlsMode] = React.useState<ExternalImapTlsMode>(account?.tlsMode ?? "tls")
  const [storageMode, setStorageMode] = React.useState<ExternalImapStorageMode>(account?.storageMode ?? "local")
  const [syncReadState, setSyncReadState] = React.useState(account?.syncReadState ?? true)
  const [enabled, setEnabled] = React.useState(account?.enabled ?? true)
  const [localError, setLocalError] = React.useState("")
  const [passwordVisible, setPasswordVisible] = React.useState(false)

  async function submit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (busy) return
    const form = new FormData(event.currentTarget)
    const payload: ExternalImapAccountPayload = {
      mailboxId: account?.mailboxId || mailboxId,
      name: String(form.get("name") || "").trim(),
      host: String(form.get("host") || "").trim(),
      port: Number(form.get("port") || (tlsMode === "tls" ? 993 : 143)),
      tlsMode,
      username: String(form.get("username") || "").trim(),
      password: String(form.get("password") || ""),
      storageMode,
      syncReadState,
      enabled,
    }
    if (!payload.mailboxId) {
      setLocalError(uiText("缺少本地邮箱容器，无法保存外部 IMAP 账号"))
      return
    }
    if (!account && !payload.password) {
      setLocalError(uiText("新邮箱需要填写 IMAP 密码或应用专用密码"))
      return
    }
    if (!Number.isInteger(payload.port) || payload.port < 1 || payload.port > 65535) {
      setLocalError(uiText("IMAP 端口必须在 1 到 65535 之间"))
      return
    }
    setLocalError("")
    try {
      await onSave(payload)
      onClose()
    } catch (error) {
      setLocalError(errorMessage(error))
    }
  }

  return (
    <form onSubmit={submit} className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="lite-imap-name">{uiText("邮箱名称")}</Label>
          <Input id="lite-imap-name" name="name" defaultValue={account?.name || ""} placeholder={uiText("例如：Netflix 账号 1")} maxLength={128} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-user">{uiText("邮箱地址 / 用户名")}</Label>
          <Input id="lite-imap-user" name="username" type="text" autoComplete="username" required defaultValue={account?.username || ""} placeholder={"name@seek.li"} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-host">{uiText("IMAP 服务器")}</Label>
          <Input id="lite-imap-host" name="host" required defaultValue={account?.host || ""} placeholder={"imap.seek.li"} autoComplete="off" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-port">{uiText("端口")}</Label>
          <Input id="lite-imap-port" name="port" type="number" min={1} max={65535} required defaultValue={account?.port ?? 993} />
        </div>
        <div className="space-y-2">
          <Label>{uiText("连接加密")}</Label>
          <Select value={tlsMode} onValueChange={(value) => setTlsMode(value as ExternalImapTlsMode)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="tls">{uiText("SSL/TLS（推荐）")}</SelectItem>
              <SelectItem value="starttls">STARTTLS</SelectItem>
              <SelectItem value="plain">{uiText("不加密（不推荐）")}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>{uiText("存储模式")}</Label>
          <Select value={storageMode} onValueChange={(value) => setStorageMode(value as ExternalImapStorageMode)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="local">{uiText("同步到本地（共享必选）")}</SelectItem>
              <SelectItem value="remote">{uiText("远端直连（不能共享）")}</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-password">{uiText(account ? "密码（留空保持原密码）" : "IMAP 密码")}</Label>
          <div className="flex gap-2">
            <Input id="lite-imap-password" name="password" type={passwordVisible ? "text" : "password"} autoComplete="new-password" placeholder={uiText(account ? "不修改请留空" : "请输入邮箱密码")} />
            <Button type="button" variant="outline" onClick={() => setPasswordVisible((value) => !value)}>
              {uiText(passwordVisible ? "隐藏" : "显示")}
            </Button>
          </div>
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-5 rounded-lg border bg-muted/20 px-4 py-3">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={enabled} onCheckedChange={(value) => setEnabled(value === true)} />
          {uiText("启用账号")}
        </label>
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={syncReadState} onCheckedChange={(value) => setSyncReadState(value === true)} />
          {uiText("同步已读状态")}
        </label>
      </div>
      <p className="text-xs text-muted-foreground">
        {uiText(storageMode === "remote" ? "远端直连模式不能创建车友共享链接。若需要共享，请选择同步到本地。" : "同步到本地后，车友通过只读链接查看邮件；不提供发信能力。")}
      </p>
      {localError && <p role="alert" className="text-sm text-destructive">{localError}</p>}
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>{uiText("取消")}</Button>
        <Button type="submit" disabled={busy}>{uiText(busy ? "保存中…" : account ? "保存修改" : "添加并保存")}</Button>
      </div>
    </form>
  )
}

export function LiteExternalIMAPManager({ accounts, loading, mailboxId, onOpenInbox, onOpenShares }: Props) {
  useLanguage()
  const qc = useQueryClient()
  const { toast } = useToast()
  const [editor, setEditor] = React.useState<{ open: boolean; account: ExternalImapAccount | null }>({ open: false, account: null })
  const [busyAccountId, setBusyAccountId] = React.useState("")

  async function refresh() {
    await Promise.all([
      qc.invalidateQueries({ queryKey: ["mail-external-accounts"] }),
      qc.invalidateQueries({ queryKey: ["external-imap-accounts"] }),
    ])
  }
  const save = useMutation({
    mutationFn: async (payload: ExternalImapAccountPayload) => {
      if (editor.account) return api.updateExternalImapAccount(editor.account.id, payload)
      return api.createExternalImapAccount(payload)
    },
    onSuccess: async () => {
      await refresh()
      toast({ title: uiText(editor.account ? "外部邮箱已更新" : "外部邮箱已添加") })
    },
  })
  async function runAction(id: string, action: "test" | "sync" | "toggle" | "delete", account: ExternalImapAccount) {
    if (busyAccountId || save.isPending) return
    if (action === "delete" &&
      !window.confirm(uiText("确定删除「{0}」的接入配置？相关共享链接将不可用，已同步的邮件不会清除。", [account.name || account.username]))) return
    setBusyAccountId(id)
    try {
      if (action === "test") {
        const result = await api.testExternalImapAccount(id)
        toast({ title: uiText("连接成功，检测到 {0} 个文件夹", [result.folders]) })
      } else if (action === "sync") {
        await api.syncExternalImapAccount(id)
        toast({ title: uiText("同步完成") })
      } else if (action === "toggle") {
        await api.updateExternalImapAccount(id, {
          mailboxId: account.mailboxId,
          name: account.name,
          host: account.host,
          port: account.port,
          tlsMode: account.tlsMode,
          username: account.username,
          password: "",
          storageMode: account.storageMode,
          syncReadState: account.syncReadState,
          enabled: !account.enabled,
        })
        toast({ title: uiText(account.enabled ? "邮箱已停用" : "邮箱已启用") })
      } else {
        await api.deleteExternalImapAccount(id)
        toast({ title: uiText("邮箱接入配置已删除") })
      }
      await refresh()
    } catch (error) {
      toast({ title: uiText("操作失败"), description: errorMessage(error) })
    } finally {
      setBusyAccountId("")
    }
  }

  return (
    <LitePageLayout
      title={uiText("邮箱接入管理")}
      description={uiText("接入并同步外部 IMAP 邮箱，集中管理同步和车友共享。")}
      action={
        <Button disabled={!mailboxId || save.isPending} onClick={() => setEditor({ open: true, account: null })}>
          <Plus className="mr-2 h-4 w-4" />{uiText("添加外部邮箱")}
        </Button>
      }
    >
      {!mailboxId && !loading && (
        <p role="alert" className="rounded-md border p-4 text-sm text-muted-foreground">
          {uiText("当前账号没有可用的本地邮箱容器，无法保存外部邮箱接入配置。")}
        </p>
      )}
      <LiteSection title={uiText("已接入邮箱")} description={uiText("管理 IMAP 连接、同步状态与账号配置。")}>
        {loading ? (
          <div className="flex items-center gap-2 py-8 text-sm text-muted-foreground">
            <LoaderCircle className="h-4 w-4 animate-spin" />{uiText("正在读取邮箱...")}
          </div>
        ) : accounts.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed px-5 py-10 text-center">
            <Inbox className="h-8 w-8 text-muted-foreground" />
            <p className="font-medium">{uiText("还没有外部 IMAP 邮箱")}</p>
            <p className="text-sm text-muted-foreground">{uiText("添加一个外部邮箱即可同步并生成共享链接。")}</p>
            <Button variant="outline" disabled={!mailboxId} onClick={() => setEditor({ open: true, account: null })}>
              {uiText("添加外部邮箱")}
            </Button>
          </div>
        ) : (
          <div className="divide-y">
            {accounts.map((account) => (
              <section key={account.id} className="space-y-3 py-4 first:pt-0 last:pb-0">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0 space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <ServerCog className="h-4 w-4 text-muted-foreground" />
                      <h2 className="break-words text-sm font-semibold">{account.name || account.username}</h2>
                      <Badge variant={account.enabled ? "secondary" : "outline"}>
                        {uiText(account.enabled ? "已启用" : "已停用")}
                      </Badge>
                      <Badge variant="outline">{uiText(account.storageMode === "local" ? "本地同步" : "远端直连")}</Badge>
                    </div>
                    <p className="break-all text-xs text-muted-foreground">
                      {account.username} · {account.host}:{account.port} · {account.tlsMode.toUpperCase()}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {uiText("状态：")}{uiText(account.lastStatus || "尚未同步")}
                      {account.lastSyncAt && <> · {uiText("最近同步：")}{formatDateTime(account.lastSyncAt)}</>}
                    </p>
                    {account.lastError && <p className="break-words text-xs text-destructive">{account.lastError}</p>}
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => void runAction(account.id, "test", account)}>{uiText("测试连接")}</Button>
                    {account.storageMode === "local" && (
                      <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => void runAction(account.id, "sync", account)}>
                        <RefreshCcw className="mr-1.5 h-3.5 w-3.5" />{uiText(busyAccountId === account.id ? "处理中" : "同步")}
                      </Button>
                    )}
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => setEditor({ open: true, account })}>{uiText("编辑")}</Button>
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => void runAction(account.id, "toggle", account)}>{uiText(account.enabled ? "停用" : "启用")}</Button>
                    <Button size="sm" variant="destructive" disabled={!!busyAccountId} onClick={() => void runAction(account.id, "delete", account)}>
                      <Trash2 className="mr-1.5 h-3.5 w-3.5" />{uiText("删除")}
                    </Button>
                  </div>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button size="sm" disabled={!account.enabled} onClick={() => onOpenInbox(account)}>
                    <Inbox className="mr-1.5 h-3.5 w-3.5" />{uiText("查看收件箱")}
                  </Button>
                  {account.storageMode === "local" && (
                    <Button size="sm" variant="outline" onClick={onOpenShares}>
                      <CheckCircle2 className="mr-1.5 h-3.5 w-3.5" />{uiText("共享管理")}
                    </Button>
                  )}
                </div>
              </section>
            ))}
          </div>
        )}
      </LiteSection>
      <Dialog open={editor.open} onOpenChange={(open) => { if (!save.isPending && !open) setEditor({ open: false, account: null }) }}>
        <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>{uiText(editor.account ? "编辑外部 IMAP 邮箱" : "添加外部 IMAP 邮箱")}</DialogTitle>
          </DialogHeader>
          {editor.open && (
            <ExternalAccountEditor
              key={editor.account?.id || "new"}
              account={editor.account}
              mailboxId={mailboxId}
              busy={save.isPending}
              onClose={() => setEditor({ open: false, account: null })}
              onSave={async (payload) => { await save.mutateAsync(payload) }}
            />
          )}
        </DialogContent>
      </Dialog>
    </LitePageLayout>
  )
}
