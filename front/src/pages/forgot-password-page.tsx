import { Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { authApi } from "@/features/auth/api";
import { forgotPasswordSchema, type ForgotPasswordRequest } from "@/features/auth/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { getErrorMessage } from "@/shared/lib/http";

export function ForgotPasswordPage() {
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<ForgotPasswordRequest>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: { email: "" },
  });

  async function onSubmit(values: ForgotPasswordRequest) {
    try {
      await authApi.forgotPassword(values);
      toast.success("Инструкции отправлены на email");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-8">
      <div className="space-y-1">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-[#ff5500]">Доступ</p>
        <h1 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
          Забыли пароль?
        </h1>
        <p className="text-sm text-[#666666]">
          Введите email — мы отправим ссылку для сброса пароля.
        </p>
      </div>

      <form className="grid gap-5" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Email" error={errors.email?.message}>
          <Input {...register("email")} type="email" placeholder="example@mail.com" />
        </FormField>
        <Button type="submit" disabled={isSubmitting} size="lg" className="w-full">
          {isSubmitting ? "Отправка..." : "Отправить инструкции"}
        </Button>
      </form>

      <p className="text-center text-sm text-[#666666]">
        Вспомнили пароль?{" "}
        <Link to="/login" className="font-semibold text-[#ff5500] hover:text-[#ff7733] transition-colors">
          Войти
        </Link>
      </p>
    </div>
  );
}
