import { useEffect } from "react";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { UserCircle, Copy, ShieldCheck } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { profileSchema, type ProfileFormValues } from "@/features/profile/schemas";
import { useDeleteMe, useMe, useUpdateMe } from "@/features/profile/hooks";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { getErrorMessage } from "@/shared/lib/http";

function copyToClipboard(text: string) {
  void navigator.clipboard.writeText(text);
  toast.success("Скопировано");
}

function getInitials(user: { first_name?: string | null; nickname?: string | null }) {
  if (user.first_name) return user.first_name.slice(0, 2).toUpperCase();
  if (user.nickname)   return user.nickname.slice(0, 2).toUpperCase();
  return "??";
}

export function ProfilePage() {
  const { logout } = useAuth();
  const meQuery = useMe();
  const updateMutation = useUpdateMe();
  const deleteMutation = useDeleteMe();

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: { first_name: "", last_name: "", nickname: "", phone: "" },
  });

  useEffect(() => {
    if (meQuery.data) {
      form.reset({
        first_name: meQuery.data.first_name ?? "",
        last_name:  meQuery.data.last_name  ?? "",
        nickname:   meQuery.data.nickname   ?? "",
        phone:      meQuery.data.phone      ?? "",
      });
    }
  }, [form, meQuery.data]);

  if (meQuery.isLoading) return <div className="flex items-center justify-center py-32"><Spinner /></div>;
  if (meQuery.isError || !meQuery.data) return <ErrorState />;

  const me = meQuery.data;
  const displayName = [me.first_name, me.last_name].filter(Boolean).join(" ") || me.nickname || "Пользователь";
  const isPlatformAdmin = me.role === "platform_admin" || me.is_platform_admin;

  async function onSubmit(values: ProfileFormValues) {
    try {
      await updateMutation.mutateAsync(values);
      toast.success("Профиль обновлён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function onDeleteAccount() {
    if (!window.confirm("Удалить аккаунт? Это действие нельзя отменить.")) return;
    try {
      await deleteMutation.mutateAsync();
      toast.success("Аккаунт удалён");
      await logout();
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-0">

      {/* ── Banner ───────────────────────────────────────────────── */}
      <div
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          background: "#111111",
          borderBottom: "1px solid #2d2d2d",
        }}
      >
        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-wrap items-center gap-6">
            {/* avatar */}
            <div className="flex h-20 w-20 shrink-0 items-center justify-center rounded-2xl bg-[#ff5500] text-2xl font-black text-white">
              {getInitials(me)}
            </div>

            <div className="space-y-1.5">
              <div className="flex flex-wrap items-center gap-2">
                <h1
                  className="font-black uppercase text-white"
                  style={{ fontSize: "clamp(1.5rem, 4vw, 2.5rem)", letterSpacing: "-0.03em" }}
                >
                  {displayName}
                </h1>
                {isPlatformAdmin && (
                  <span className="flex items-center gap-1 rounded-full bg-[#ff5500]/20 px-3 py-1 text-xs font-bold text-[#ff5500]">
                    <ShieldCheck className="h-3.5 w-3.5" />
                    Admin
                  </span>
                )}
              </div>
              {me.nickname && (
                <p className="text-sm text-[#666666]">@{me.nickname}</p>
              )}
              <p className="text-sm text-[#9e9e9e]">{me.email}</p>
            </div>
          </div>
        </div>
      </div>

      {/* ── Content ──────────────────────────────────────────────── */}
      <div className="py-8 grid gap-6 lg:grid-cols-[1fr_320px] lg:items-start">

        {/* edit form */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <UserCircle className="h-5 w-5 text-[#ff5500]" />
              Личные данные
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form className="grid gap-5" onSubmit={form.handleSubmit(onSubmit)}>
              <div className="grid gap-4 sm:grid-cols-2">
                <FormField label="Имя" error={form.formState.errors.first_name?.message}>
                  <Input {...form.register("first_name")} placeholder="Иван" />
                </FormField>
                <FormField label="Фамилия" error={form.formState.errors.last_name?.message}>
                  <Input {...form.register("last_name")} placeholder="Иванов" />
                </FormField>
              </div>
              <FormField label="Никнейм" error={form.formState.errors.nickname?.message}>
                <Input {...form.register("nickname")} placeholder="player123" />
              </FormField>
              <FormField label="Телефон" error={form.formState.errors.phone?.message}>
                <Input {...form.register("phone")} placeholder="+71234567890" />
              </FormField>
              <div>
                <Button type="submit" disabled={updateMutation.isPending}>
                  {updateMutation.isPending ? "Сохранение..." : "Сохранить изменения"}
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>

        {/* right sidebar */}
        <div className="grid gap-4">
          {/* account info */}
          <Card>
            <CardHeader>
              <CardTitle className="text-sm">Информация об аккаунте</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 text-sm">
              <div className="space-y-1">
                <p className="text-xs text-[#666666] uppercase tracking-wide">Email</p>
                <p className="text-white">{me.email}</p>
              </div>
              <div className="space-y-1">
                <p className="text-xs text-[#666666] uppercase tracking-wide">ID пользователя</p>
                <div className="flex items-center gap-2">
                  <code className="flex-1 truncate rounded-lg bg-[#111111] px-2 py-1.5 text-xs text-[#9e9e9e] select-all">
                    {me.id}
                  </code>
                  <button
                    type="button"
                    onClick={() => copyToClipboard(me.id)}
                    className="shrink-0 rounded-lg border border-[#2d2d2d] p-1.5 text-[#666666] hover:border-[#ff5500] hover:text-[#ff5500] transition-colors"
                  >
                    <Copy className="h-3.5 w-3.5" />
                  </button>
                </div>
                <p className="text-[10px] text-[#444444]">Используется для добавления менеджеров турнира</p>
              </div>
            </CardContent>
          </Card>

          {/* danger zone */}
          <Card className="border-red-900/40">
            <CardHeader>
              <CardTitle className="text-sm text-red-400">Опасная зона</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3">
              <p className="text-xs text-[#666666]">
                Удаление аккаунта необратимо. Все данные будут утеряны.
              </p>
              <Button
                type="button"
                variant="destructive"
                size="sm"
                disabled={deleteMutation.isPending}
                onClick={() => void onDeleteAccount()}
              >
                {deleteMutation.isPending ? "Удаление..." : "Удалить аккаунт"}
              </Button>
            </CardContent>
          </Card>
        </div>

      </div>
    </div>
  );
}
