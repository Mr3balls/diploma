import { useSearchParams } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { authApi } from "@/features/auth/api";
import { resetPasswordSchema, type ResetPasswordRequest } from "@/features/auth/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { getErrorMessage } from "@/shared/lib/http";

export function ResetPasswordPage() {
  const [params] = useSearchParams();

  const form = useForm<ResetPasswordRequest>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: {
      token: params.get("token") ?? "",
      password: "",
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form;

  async function onSubmit(values: ResetPasswordRequest) {
    try {
      const response = await authApi.resetPassword(values);
      toast.success("Демо-сброс выполнен");
      console.info("Reset password demo response", response);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-6 p-6 md:p-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Сброс пароля</h1>
        <p className="text-sm text-slate-500">
          Страница подключена к demo stub backend endpoint. Здесь нет выдуманной почтовой логики.
        </p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Токен" error={errors.token?.message}>
          <Input {...register("token")} />
        </FormField>
        <FormField label="Новый пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" />
        </FormField>
        <Button type="submit" disabled={isSubmitting}>
          Сменить пароль
        </Button>
      </form>
    </div>
  );
}