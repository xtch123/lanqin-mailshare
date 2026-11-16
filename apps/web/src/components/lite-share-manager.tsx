import * as React from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Check, Copy, ExternalLink, Link2, LoaderCircle, ShieldOff, Trash2 } from "lucide-react"
import { api, type ExternalImapAccount, type ExternalImapShare, type ExternalImapShareCreate } from "@/lib/api"
import { uiText, useLanguage } from "@/lib/language"
import { errorMessage } from "@/lib/ui-errors"
import { formatDateTime } from "@/lib/utils"
import { useToast } from "@/hooks/use-toast"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { LitePageLayout, LiteSection } from "@/components/lite-page-layout"

type Props = { accounts: ExternalImapAccount[]; onOpenAccounts: () => void }
type Status = "active" | "expired" | "revoked"

const timeOptions = [
  { value: "30", label: "30 天" },
  { value: "7", label: "7 天" },
  { value: "1", label: "1 天" },
  { value: "0", label: "永久有效，直到撤销" },
  { value: "custom", label: "自定义天数" },
]
const rangeOptions = [
  { value: "1440", label: "最近 24 小时" },
  { value: "60", label: "最近 1 小时" },
  { value: "10080", label: "最近 7 天" },
  { value: "0", label: "全部已同步邮件" },
  { value: "custom", label: "自定义分钟数" },
]

function shareStatus(item: ExternalImapShare): Status {
  if (item.revokedAt) return "revoked"
  if (item.expiresAt && Date.parse(item.expiresAt) <= Date.now()) return "expired"
  return "active"
}

function visibleRange(minutes: number): string {
  if (minutes === 0) return uiText("全部已同步邮件")
  if (minutes % 1440 === 0) return uiText("最近 {0} 天", [minutes / 1440])
  if (minutes % 60 === 0) return uiText("最近 {0} 小时", [minutes / 60])
  return uiText("最近 {0} 分钟", [minutes])
}

export function LiteShareManager({ accounts, onOpenAccounts }: Props) {
  useLanguage()
  const qc = useQueryClient()
  const { toast } = useToast()
  const localAccounts = React.useMemo(() => accounts.filter((a) => a.enabled && a.storageMode === "local"), [accounts])
  const [accountId, setAccountId] = React.useState("")
  const [filterId, setFilterId] = React.useState("all")
  const [label, setLabel] = React.useState("")
  const [days, setDays] = React.useState("30")
  const [customDays, setCustomDays] = React.useState("90")
  const [visible, setVisible] = React.useState("1440")
  const [customVisible, setCustomVisible] = React.useState("30")
  const [link, setLink] = React.useState("")
  const [linkCopied, setLinkCopied] = React.useState(false)
  const [busyId, setBusyId] = React.useState("")

  React.useEffect(() => {
    if (!localAccounts.some((a) => a.id === accountId)) {
      setAccountId(localAccounts[0]?.id || "")
    }
  }, [accountId, localAccounts])

  React.useEffect(() => {
    if (filterId !== "all" && !accounts.some((a) => a.id === filterId)) setFilterId("all")
  }, [accounts, filterId])

  const shares = useQuery({
    queryKey: ["lite-shares", filterId],
    queryFn: () => api.externalImapShares(filterId === "all" ? undefined : filterId),
    refetchInterval: 15000,
  })
  const refresh = () => qc.invalidateQueries({ queryKey: ["lite-shares"] })
  const create = useMutation({
    mutationFn: (payload: ExternalImapShareCreate) => api.createExternalImapShare(payload),
    onSuccess: async ({ url }) => {
      setLink(url)
      setLinkCopied(false)
      setLabel("")
      await refresh()
      toast({ title: uiText("共享链接已创建") })
    },
    onError: (error) => toast({ title: uiText("创建失败"), description: errorMessage(error) }),
  })

  function createShare(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!accountId) return
    const duration = days === "custom" ? Number(customDays) : Number(days)
    const minutes = visible === "custom" ? Number(customVisible) : Number(visible)
    if (!Number.isInteger(duration) || duration < 0 || duration > 3650 ||
      (days === "custom" && duration === 0) ||
      !Number.isInteger(minutes) || minutes < 0 || minutes > 525600) {
      toast({ title: uiText("时间范围不合法") })
      return
    }
    const payload: ExternalImapShareCreate = { accountId, label: label.trim(), visibilityMinutes: minutes }
    if (duration > 0) payload.expiresAt = new Date(Date.now() + duration * 86400000).toISOString()
    create.mutate(payload)
  }

  async function copyLink(url: string) {
    setLink(url)
    setLinkCopied(false)
    try {
      await navigator.clipboard.writeText(url)
      setLinkCopied(true)
      toast({ title: uiText("共享链接已复制") })
    } catch {
      toast({ title: uiText("请从下方手动复制链接") })
    }
  }

  async function runAction(item: ExternalImapShare, type: "copy" | "rotate" | "revoke" | "delete") {
    if (busyId || create.isPending) return
    if (type === "rotate" && !window.confirm(uiText("重置后旧链接立即失效，确定继续吗？"))) return
    if (type === "revoke" && !window.confirm(uiText("确定撤销？该车友将立即无法继续访问。"))) return
    if (type === "delete" && !window.confirm(uiText("永久删除这条已撤销的记录？不会删除邮箱或邮件。"))) return
    setBusyId(item.id)
    try {
      if (type === "copy") {
        const value = await api.externalImapShareLink(item.id)
        await copyLink(value.url)
      } else if (type === "rotate") {
        const value = await api.rotateExternalImapShare(item.id)
        await copyLink(value.url)
        await refresh()
      } else if (type === "revoke") {
        await api.revokeExternalImapShare(item.id)
        setLink("")
        await refresh()
        toast({ title: uiText("共享已撤销") })
      } else {
        await api.deleteRevokedExternalImapShare(item.id)
        await refresh()
        toast({ title: uiText("共享记录已删除") })
      }
    } catch (error) {
      toast({ title: uiText("操作失败"), description: errorMessage(error) })
    } finally {
      setBusyId("")
    }
  }

  return (
    <LitePageLayout title={uiText("共享管理")} description={uiText("为车友创建只读链接，设置有效期与可查看的邮件范围。")}>
      <div className="grid items-start gap-4 lg:grid-cols-[minmax(0,1.4fr)_minmax(320px,1fr)]">
        <LiteSection title={uiText("创建共享链接")} description={uiText("每位车友使用独立链接，方便单独撤销权限。")}>
          <form onSubmit={createShare} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="lite-share-account">{uiText("外部邮箱")}</Label>
              <Select value={accountId || undefined} onValueChange={setAccountId}>
                <SelectTrigger id="lite-share-account"><SelectValue placeholder={uiText("选择外部邮箱")} /></SelectTrigger>
                <SelectContent>{localAccounts.map((account) => <SelectItem key={account.id} value={account.id}>{account.name || account.username} ({account.username})</SelectItem>)}</SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="lite-share-label">{uiText("车友备注（选填）")}</Label>
              <Input id="lite-share-label" maxLength={64} value={label} onChange={(event) => setLabel(event.target.value)} placeholder={uiText("例如：一号车 · 成员 A")} />
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-2">
                <Label>{uiText("共享链接有效期")}</Label>
                <Select value={days} onValueChange={setDays}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>{timeOptions.map((item) => <SelectItem value={item.value} key={item.value}>{uiText(item.label)}</SelectItem>)}</SelectContent>
                </Select>
                {days === "custom" && <Input type="number" min={1} max={3650} value={customDays} onChange={(event) => setCustomDays(event.target.value)} aria-label={uiText("自定义有效期（天）")} />}
              </div>
              <div className="space-y-2">
                <Label>{uiText("历史邮件可见范围")}</Label>
                <Select value={visible} onValueChange={setVisible}>
                  <SelectTrigger><SelectValue /></SelectTrigger>
                  <SelectContent>{rangeOptions.map((item) => <SelectItem value={item.value} key={item.value}>{uiText(item.label)}</SelectItem>)}</SelectContent>
                </Select>
                {visible === "custom" && <Input type="number" min={0} max={525600} value={customVisible} onChange={(event) => setCustomVisible(event.target.value)} aria-label={uiText("自定义可见范围（分钟）")} />}
              </div>
            </div>
            <div className="flex flex-wrap items-center justify-between gap-2 pt-1">
              <Button type="submit" disabled={!accountId || create.isPending}>
                {create.isPending ? <LoaderCircle className="mr-2 h-4 w-4 animate-spin" /> : <Link2 className="mr-2 h-4 w-4" />}
                {uiText("生成专属链接")}
              </Button>
              {!localAccounts.length && <Button variant="outline" type="button" onClick={onOpenAccounts}>{uiText("先添加外部邮箱")}</Button>}
            </div>
          </form>
          {link && <div className="mt-5 space-y-3 rounded-md border bg-muted/20 p-3">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-medium">{uiText("共享链接已准备好")}</p>
              <div className="flex gap-2">
                <Button type="button" size="sm" variant="outline" onClick={() => void copyLink(link)}>
                  {linkCopied ? <Check className="mr-1 h-3.5 w-3.5" /> : <Copy className="mr-1 h-3.5 w-3.5" />}{linkCopied ? uiText("已复制") : uiText("复制链接")}
                </Button>
                <Button type="button" size="sm" variant="outline" asChild><a href={link} target="_blank" rel="noopener noreferrer"><ExternalLink className="mr-1 h-3.5 w-3.5" />{uiText("预览")}</a></Button>
              </div>
            </div>
            <Input readOnly value={link} onFocus={(e) => e.currentTarget.select()} aria-label={uiText("共享链接")} />
            <p className="text-xs text-muted-foreground">{uiText("如复制失败，可选中上方链接手动复制。")}</p>
          </div>}
        </LiteSection>

        <LiteSection title={uiText("已有共享")} description={uiText("按邮箱筛选、复制链接或撤销访问。")}>
          <div className="mb-4 space-y-2">
            <Label>{uiText("按邮箱筛选")}</Label>
            <Select value={filterId} onValueChange={setFilterId}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">{uiText("全部外部邮箱")}</SelectItem>
                {accounts.map((account) => <SelectItem value={account.id} key={account.id}>{account.name || account.username}</SelectItem>)}
              </SelectContent>
            </Select>
          </div>
          {shares.isLoading ? <div className="flex items-center gap-2 text-sm text-muted-foreground"><LoaderCircle className="h-4 w-4 animate-spin" />{uiText("正在读取...")}</div>
            : shares.isError ? <div role="alert" className="text-sm text-destructive">{errorMessage(shares.error)}</div>
              : !shares.data?.items.length ? <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">{uiText("尚无共享链接")}</div>
                : <div className="divide-y">{shares.data.items.map((item) => {
                    const status = shareStatus(item)
                    return <article key={item.id} className="space-y-3 py-4 first:pt-0 last:pb-0">
                      <div className="flex flex-wrap items-center gap-2">
                        <span className="min-w-0 flex-1 break-words text-sm font-semibold">{item.accountName || item.accountEmail}</span>
                        <Badge variant="outline">{uiText(status === "active" ? "有效" : status === "expired" ? "已过期" : "已撤销")}</Badge>
                      </div>
                      <div className="break-all text-xs text-muted-foreground">{item.accountEmail}</div>
                      <div className="text-sm">{uiText("车友：")}{item.label || uiText("未命名")}</div>
                      <p className="text-xs leading-relaxed text-muted-foreground">
                        {visibleRange(item.visibilityMinutes)} · {item.expiresAt ? `${uiText("到期")} ${formatDateTime(item.expiresAt)}` : uiText("永久有效")}
                      </p>
                      <div className="flex flex-wrap gap-2">
                        {status === "active" && <Button size="sm" variant="outline" disabled={!!busyId} onClick={() => void runAction(item, item.canCopy ? "copy" : "rotate")}>
                          <Copy className="mr-1 h-3.5 w-3.5" />{uiText(item.canCopy ? "复制链接" : "重置链接")}
                        </Button>}
                        {status !== "revoked" && <Button size="sm" variant="outline" disabled={!!busyId} onClick={() => void runAction(item, "revoke")}>
                          <ShieldOff className="mr-1 h-3.5 w-3.5" />{uiText("撤销")}
                        </Button>}
                        {status === "revoked" && <Button size="sm" variant="destructive" disabled={!!busyId} onClick={() => void runAction(item, "delete")}>
                          <Trash2 className="mr-1 h-3.5 w-3.5" />{uiText("删除记录")}
                        </Button>}
                      </div>
                    </article>
                  })}</div>}
        </LiteSection>
      </div>
    </LitePageLayout>
  )
}
