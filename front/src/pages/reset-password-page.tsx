import { Link, useSearchParams } from "react-router-dom";
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

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<ResetPasswordRequest>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { token: params.get("token") ?? "", password: "" },
  });

  async function onSubmit(values: ResetPasswordRequest) {
    try {
      await authApi.resetPassword(values);
      toast.success("Пароль успешно изменён");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-8">
      <div className="space-y-1">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-[#ff5500]">Безопасность</p>
        <h1 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
          Новый пароль
        </h1>
        <p className="text-sm text-[#666666]">Введите токен из письма и задайте новый пароль.</p>
      </div>

      <form className="grid gap-5" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Токен" error={errors.token?.message}>
          <Input {...register("token")} placeholder="Токен из письма" />
        </FormField>
        <FormField label="Новый пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" placeholder="••••••••" />
        </FormField>
        <Button type="submit" disabled={isSubmitting} size="lg" className="w-full">
          {isSubmitting ? "Сохранение..." : "Сменить пароль"}
        </Button>
      </form>

      <p className="text-center text-sm text-[#666666]">
        <Link to="/login" className="font-semibold text-[#ff5500] hover:text-[#ff7733] transition-colors">
          Вернуться ко входу
        </Link>
      </p>
    </div>
  );
}
