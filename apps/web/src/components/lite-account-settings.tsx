import * as React from "react"
import { useMutation, useQueryClient } from "@tanstack/react-query"
import { LogOut } from "lucide-react"
import { api } from "@/lib/api"
import { uiText, useLanguage } from "@/lib/language"
import { errorMessage } from "@/lib/ui-errors"
import { formatDateTime } from "@/lib/utils"
import { useMe } from "@/hooks/use-me"
import { useLogout } from "@/hooks/use-logout"
import { useToast } from "@/hooks/use-toast"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { LitePageLayout, LiteSection } from "@/components/lite-page-layout"

export function LiteAccountSettings() {
  useLanguage()
  const me = useMe()
  const qc = useQueryClient()
  const logout = useLogout()
  const { toast } = useToast()
  const [displayName, setDisplayName] = React.useState("")
  const [currentPassword, setCurrentPassword] = React.useState("")
  const [newPassword, setNewPassword] = React.useState("")
  const [confirmPassword, setConfirmPassword] = React.useState("")

  React.useEffect(() => {
    if (me.data?.user) setDisplayName(me.data.user.displayName)
  }, [me.data?.user?.displayName])

  const saveProfile = useMutation({
    mutationFn: () => api.updateProfile({ displayName: displayName.trim() }),
    onSuccess: (data) => {
      qc.setQueryData(["me"], data)
      toast({ title: uiText("账户资料已保存") })
    },
    onError: (error) => toast({ title: uiText("保存失败"), description: errorMessage(error) }),
  })
  const savePassword = useMutation({
    mutationFn: async () => {
      if (newPassword.length < 8) throw new Error(uiText("新密码至少需要 8 个字符"))
      if (newPassword !== confirmPassword) throw new Error(uiText("两次输入的新密码不一致"))
      return api.changePassword({ currentPassword, newPassword })
    },
    onSuccess: () => {
      setCurrentPassword("")
      setNewPassword("")
      setConfirmPassword("")
      toast({ title: uiText("密码已更新") })
    },
    onError: (error) => toast({ title: uiText("修改失败"), description: errorMessage(error) }),
  })
  const user = me.data?.user

  return (
    <LitePageLayout title={uiText("账户设置")} description={uiText("管理登录账户和安全设置。")}>
      <LiteSection title={uiText("账户信息")}>
        <form className="space-y-4" onSubmit={(event) => { event.preventDefault(); saveProfile.mutate() }}>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="lite-account-email">{uiText("用户名")}</Label>
              <Input id="lite-account-email" value={user?.email || ""} readOnly />
            </div>
            <div className="space-y-2">
              <Label htmlFor="lite-account-name">{uiText("显示名称")}</Label>
              <Input id="lite-account-name" value={displayName} onChange={(e) => setDisplayName(e.target.value)} maxLength={128} required />
            </div>
          </div>
          <div className="flex justify-end">
            <Button type="submit" disabled={!displayName.trim() || saveProfile.isPending}>{saveProfile.isPending ? uiText("保存中...") : uiText("保存资料")}</Button>
          </div>
        </form>
        <div className="mt-4 grid gap-2 border-t pt-4 text-sm">
          <div className="flex items-center justify-between gap-3 rounded-md border px-3 py-2"><span>{uiText("角色")}</span><Badge variant="secondary">{user?.role === "admin" ? uiText("超级管理员") : uiText("普通用户")}</Badge></div>
          <div className="flex items-center justify-between gap-3 rounded-md border px-3 py-2"><span>{uiText("账号状态")}</span><Badge variant="secondary">{user?.disabled ? uiText("已停用") : uiText("正常")}</Badge></div>
          <div className="flex items-center justify-between gap-3 rounded-md border px-3 py-2"><span>{uiText("创建时间")}</span><span className="text-muted-foreground">{user?.createdAt ? formatDateTime(user.createdAt) : "—"}</span></div>
          {user?.twoFactorEnabled && <p className="text-xs text-muted-foreground">{uiText("此账户已启用双因素认证。")}</p>}
        </div>
      </LiteSection>
      <LiteSection title={uiText("修改密码")}>
        <form className="space-y-4" onSubmit={(event) => { event.preventDefault(); savePassword.mutate() }}>
          <div className="space-y-2">
            <Label htmlFor="lite-password-current">{uiText("当前密码")}</Label>
            <Input id="lite-password-current" type="password" autoComplete="current-password" required value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="lite-password-new">{uiText("新密码")}</Label>
              <Input id="lite-password-new" type="password" autoComplete="new-password" minLength={8} required value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="lite-password-confirm">{uiText("确认新密码")}</Label>
              <Input id="lite-password-confirm" type="password" autoComplete="new-password" minLength={8} required value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} />
            </div>
          </div>
          <div className="flex justify-end">
            <Button disabled={savePassword.isPending} type="submit">{savePassword.isPending ? uiText("更新中...") : uiText("更新密码")}</Button>
          </div>
        </form>
      </LiteSection>
      <LiteSection title={uiText("登录状态")} description={uiText("退出当前浏览器的管理员登录。")}>
        <Button variant="outline" onClick={() => void logout()}><LogOut className="mr-2 h-4 w-4" />{uiText("退出登录")}</Button>
      </LiteSection>
    </LitePageLayout>
  )
}
