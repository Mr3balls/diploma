import { useEffect } from "react";
import { toast } from "sonner";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAuth } from "@/app/providers/auth-provider";
import { profileSchema, type ProfileFormValues } from "@/features/profile/schemas";
import { useDeleteMe, useMe, useUpdateMe } from "@/features/profile/hooks";
import { PageHeader } from "@/shared/ui/page-header";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { getErrorMessage } from "@/shared/lib/http";

export function ProfilePage() {
  const { logout } = useAuth();
  const meQuery = useMe();
  const updateMutation = useUpdateMe();
  const deleteMutation = useDeleteMe();

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      first_name: "",
      last_name: "",
      nickname: "",
      phone: "",
    },
  });

  useEffect(() => {
    if (meQuery.data) {
      form.reset({
        first_name: meQuery.data.first_name ?? "",
        last_name: meQuery.data.last_name ?? "",
        nickname: meQuery.data.nickname ?? "",
        phone: meQuery.data.phone ?? "",
      });
    }
  }, [form, meQuery.data]);

  if (meQuery.isLoading) return <Spinner />;
  if (meQuery.isError || !meQuery.data) return <ErrorState />;

  async function onSubmit(values: ProfileFormValues) {
    try {
      await updateMutation.mutateAsync(values);
      toast.success("Профиль обновлён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function onDeleteAccount() {
    try {
      await deleteMutation.mutateAsync();
      toast.success("Аккаунт удалён");
      await logout();
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-6">
      <PageHeader title="Профиль" description="Редактирование базовых данных пользователя." />

      <Card>
        <CardContent className="grid gap-4 pt-5">
          <div className="text-sm text-[#90afd4]">Email: {meQuery.data.email}</div>
          <div className="flex items-center gap-2 text-sm text-[#90afd4]">
            <span>ID пользователя:</span>
            <code className="rounded bg-[#001538] px-2 py-0.5 text-xs text-white select-all">{meQuery.data.id}</code>
            <span className="text-xs">(используется для добавления со-организаторов)</span>
          </div>
          <form className="grid gap-4 md:max-w-2xl" onSubmit={form.handleSubmit(onSubmit)}>
            <div className="grid gap-4 md:grid-cols-2">
              <FormField label="Имя" error={form.formState.errors.first_name?.message}>
                <Input {...form.register("first_name")} />
              </FormField>
              <FormField label="Фамилия" error={form.formState.errors.last_name?.message}>
                <Input {...form.register("last_name")} />
              </FormField>
            </div>
            <FormField label="Никнейм" error={form.formState.errors.nickname?.message}>
              <Input {...form.register("nickname")} />
            </FormField>
            <FormField label="Телефон" error={form.formState.errors.phone?.message}>
              <Input {...form.register("phone")} />
            </FormField>
            <div className="flex flex-wrap gap-3">
              <Button type="submit" disabled={updateMutation.isPending}>
                Сохранить
              </Button>
              <Button type="button" variant="destructive" disabled={deleteMutation.isPending} onClick={() => void onDeleteAccount()}>
                Удалить аккаунт
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}