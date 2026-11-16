import * as React from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { CheckCircle2, Inbox, LoaderCircle, Plus, RefreshCcw, ServerCog, Trash2 } from "lucide-react"
import { api, type ExternalImapAccount, type ExternalImapAccountPayload, type ExternalImapStorageMode, type ExternalImapTlsMode } from "@/lib/api"
import { errorMessage } from "@/lib/ui-errors"
import { formatDateTime } from "@/lib/utils"
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
      setLocalError("缺少本地邮箱容器，无法保存外部 IMAP 账号")
      return
    }
    if (!account && !payload.password) {
      setLocalError("新邮箱需要填写 IMAP 密码或应用专用密码")
      return
    }
    if (!Number.isInteger(payload.port) || payload.port < 1 || payload.port > 65535) {
      setLocalError("IMAP 端口必须在 1 到 65535 之间")
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
          <Label htmlFor="lite-imap-name">邮箱名称</Label>
          <Input id="lite-imap-name" name="name" defaultValue={account?.name || ""} placeholder="例如：尼区奈飞 1" maxLength={128} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-user">邮箱地址 / 用户名</Label>
          <Input id="lite-imap-user" name="username" type="text" autoComplete="username" required defaultValue={account?.username || ""} placeholder="name@seek.li" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-host">IMAP 服务器</Label>
          <Input id="lite-imap-host" name="host" required defaultValue={account?.host || ""} placeholder="imap.seek.li" autoComplete="off" />
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-port">端口</Label>
          <Input id="lite-imap-port" name="port" type="number" min={1} max={65535} required defaultValue={account?.port ?? 993} />
        </div>
        <div className="space-y-2">
          <Label>连接加密</Label>
          <Select value={tlsMode} onValueChange={(value) => setTlsMode(value as ExternalImapTlsMode)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="tls">SSL/TLS（推荐）</SelectItem>
              <SelectItem value="starttls">STARTTLS</SelectItem>
              <SelectItem value="plain">不加密（不推荐）</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label>存储模式</Label>
          <Select value={storageMode} onValueChange={(value) => setStorageMode(value as ExternalImapStorageMode)}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="local">同步到本地（共享必选）</SelectItem>
              <SelectItem value="remote">远端直连（不能共享）</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="lite-imap-password">{account ? "密码（留空保持原密码）" : "IMAP 密码"}</Label>
          <div className="flex gap-2">
            <Input id="lite-imap-password" name="password" type={passwordVisible ? "text" : "password"} autoComplete="new-password" placeholder={account ? "不修改请留空" : "请输入邮箱密码"} />
            <Button type="button" variant="outline" onClick={() => setPasswordVisible((value) => !value)}>
              {passwordVisible ? "隐藏" : "显示"}
            </Button>
          </div>
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-5 rounded-lg border bg-muted/20 px-4 py-3">
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={enabled} onCheckedChange={(value) => setEnabled(value === true)} />
          启用账号
        </label>
        <label className="flex cursor-pointer items-center gap-2 text-sm">
          <Checkbox checked={syncReadState} onCheckedChange={(value) => setSyncReadState(value === true)} />
          同步已读状态
        </label>
      </div>
      <p className="text-xs text-muted-foreground">
        {storageMode === "remote" ? "远端直连模式不能创建车友共享链接。若需要共享，请选择同步到本地。" : "同步到本地后，车友通过只读链接查看邮件；不提供发信能力。"}
      </p>
      {localError && <p role="alert" className="text-sm text-destructive">{localError}</p>}
      <div className="flex justify-end gap-2">
        <Button type="button" variant="outline" onClick={onClose}>取消</Button>
        <Button type="submit" disabled={busy}>{busy ? "保存中…" : account ? "保存修改" : "添加并保存"}</Button>
      </div>
    </form>
  )
}

export function LiteExternalIMAPManager({ accounts, loading, mailboxId, onOpenInbox, onOpenShares }: Props) {
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
      toast({ title: editor.account ? "外部邮箱已更新" : "外部邮箱已添加" })
    },
  })
  async function runAction(id: string, action: "test" | "sync" | "toggle" | "delete", account: ExternalImapAccount) {
    if (busyAccountId || save.isPending) return
    if (action === "delete" &&
      !window.confirm(`确定删除「${account.name || account.username}」的接入配置？相关共享链接将不可用，已同步的邮件不会在这里清除。`)) return
    setBusyAccountId(id)
    try {
      if (action === "test") {
        const result = await api.testExternalImapAccount(id)
        toast({ title: `连接成功，检测到 ${result.folders} 个文件夹` })
      } else if (action === "sync") {
        await api.syncExternalImapAccount(id)
        toast({ title: "同步完成" })
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
        toast({ title: account.enabled ? "邮箱已停用" : "邮箱已启用" })
      } else {
        await api.deleteExternalImapAccount(id)
        toast({ title: "邮箱接入配置已删除" })
      }
      await refresh()
    } catch (error) {
      toast({ title: "操作失败", description: errorMessage(error) })
    } finally {
      setBusyAccountId("")
    }
  }

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="mx-auto w-full max-w-5xl space-y-6 px-4 py-6 sm:px-8 sm:py-8">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div className="space-y-1">
            <div className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">EXTERNAL IMAP</div>
            <h1 className="text-2xl font-semibold tracking-tight">外部邮箱接入</h1>
            <p className="text-sm text-muted-foreground">接入并同步 seek.li 或其他 IMAP 邮箱，统一管理车友的邮件共享。</p>
          </div>
          <Button disabled={!mailboxId || save.isPending} onClick={() => setEditor({ open: true, account: null })}>
            <Plus className="mr-2 h-4 w-4" />添加外部邮箱
          </Button>
        </div>
        {!mailboxId && !loading && (
          <div role="alert" className="rounded-lg border p-4 text-sm text-muted-foreground">当前账号没有可用的本地邮箱容器，无法保存外部邮箱接入配置。</div>
        )}
        {loading ? (
          <div className="flex items-center gap-2 rounded-lg border p-8 text-sm text-muted-foreground"><LoaderCircle className="h-4 w-4 animate-spin" />正在读取邮箱...</div>
        ) : accounts.length === 0 ? (
          <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed px-5 py-14 text-center">
            <Inbox className="h-9 w-9 text-muted-foreground" />
            <h2 className="font-medium">还没有外部 IMAP 邮箱</h2>
            <p className="max-w-sm text-sm text-muted-foreground">点击「添加外部邮箱」，设置服务器地址、邮箱用户名和密码后即可同步。</p>
            <Button variant="outline" disabled={!mailboxId} onClick={() => setEditor({ open: true, account: null })}>添加外部邮箱</Button>
          </div>
        ) : (
          <div className="grid gap-4">
            {accounts.map((account) => (
              <section key={account.id} className="rounded-xl border bg-card p-4 shadow-sm sm:p-5">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div className="min-w-0 space-y-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <ServerCog className="h-4 w-4 text-muted-foreground" />
                      <h2 className="break-words font-semibold">{account.name || account.username}</h2>
                      <Badge variant={account.enabled ? "secondary" : "outline"}>{account.enabled ? "已启用" : "已停用"}</Badge>
                      <Badge variant="outline">{account.storageMode === "local" ? "本地同步" : "远端直连"}</Badge>
                    </div>
                    <p className="break-all text-xs text-muted-foreground">{account.username} · {account.host}:{account.port} · {account.tlsMode.toUpperCase()}</p>
                    <div className="flex flex-wrap gap-2 text-xs text-muted-foreground">
                      <span>状态：{account.lastStatus || "尚未同步"}</span>
                      {account.lastSyncAt && <span>最近同步：{formatDateTime(account.lastSyncAt)}</span>}
                    </div>
                    {account.lastError && <p className="max-w-xl break-words text-xs text-destructive">{account.lastError}</p>}
                  </div>
                  <div className="flex flex-wrap gap-2">
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => runAction(account.id, "test", account)}>测试连接</Button>
                    {account.storageMode === "local" && <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => runAction(account.id, "sync", account)}>
                      <RefreshCcw className="mr-1.5 h-3.5 w-3.5" />{busyAccountId === account.id ? "处理中" : "同步"}
                    </Button>}
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => setEditor({ open: true, account })}>编辑</Button>
                    <Button size="sm" variant="outline" disabled={!!busyAccountId} onClick={() => runAction(account.id, "toggle", account)}>{account.enabled ? "停用" : "启用"}</Button>
                    <Button size="sm" variant="destructive" disabled={!!busyAccountId} onClick={() => runAction(account.id, "delete", account)}>
                      <Trash2 className="mr-1.5 h-3.5 w-3.5" />删除
                    </Button>
                  </div>
                </div>
                <div className="mt-4 flex flex-wrap items-center gap-2 border-t pt-3">
                  <Button size="sm" disabled={!account.enabled} onClick={() => onOpenInbox(account)}>
                    <Inbox className="mr-1.5 h-3.5 w-3.5" />查看收件箱
                  </Button>
                  {account.storageMode === "local" && <Button size="sm" variant="outline" onClick={onOpenShares}>
                    <CheckCircle2 className="mr-1.5 h-3.5 w-3.5" />共享管理
                  </Button>}
                </div>
              </section>
            ))}
          </div>
        )}
      </div>
      <Dialog open={editor.open} onOpenChange={(open) => { if (!save.isPending && !open) setEditor({ open: false, account: null }) }}>
        <DialogContent className="max-h-[90dvh] overflow-y-auto sm:max-w-2xl">
          <DialogHeader><DialogTitle>{editor.account ? "编辑外部 IMAP 邮箱" : "添加外部 IMAP 邮箱"}</DialogTitle></DialogHeader>
          {editor.open && <ExternalAccountEditor
            key={editor.account?.id || "new"}
            account={editor.account}
            mailboxId={mailboxId}
            busy={save.isPending}
            onClose={() => setEditor({ open: false, account: null })}
            onSave={async (payload) => { await save.mutateAsync(payload) }}
          />}
        </DialogContent>
      </Dialog>
    </div>
  )
}
