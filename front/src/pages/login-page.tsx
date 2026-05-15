import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { loginSchema, type LoginRequest } from "@/features/auth/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { getErrorMessage } from "@/shared/lib/http";

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginRequest>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "" },
  });

  async function onSubmit(values: LoginRequest) {
    try {
      await login(values);
      toast.success("Вход выполнен");
      navigate("/tournaments");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-8">
      <div className="space-y-1">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-[#ff5500]">Добро пожаловать</p>
        <h1 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
          Войти
        </h1>
        <p className="text-sm text-[#666666]">Введите email и пароль вашей учётной записи.</p>
      </div>

      <form className="grid gap-5" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Email" error={errors.email?.message}>
          <Input {...register("email")} type="email" placeholder="example@mail.com" />
        </FormField>
        <FormField label="Пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" placeholder="••••••••" />
        </FormField>

        <div className="flex items-center justify-between">
          <Link to="/forgot-password" className="text-xs text-[#666666] hover:text-[#ff5500] transition-colors">
            Забыли пароль?
          </Link>
        </div>

        <Button type="submit" disabled={isSubmitting} size="lg" className="w-full">
          {isSubmitting ? "Вход..." : "Войти"}
        </Button>
      </form>

      <p className="text-center text-sm text-[#666666]">
        Нет аккаунта?{" "}
        <Link to="/register" className="font-semibold text-[#ff5500] hover:text-[#ff7733] transition-colors">
          Зарегистрироваться
        </Link>
      </p>
    </div>
  );
}
