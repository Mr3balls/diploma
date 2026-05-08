import { useNavigate } from "react-router-dom";
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

  const form = useForm<LoginRequest>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form;

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
    <div className="grid gap-6 p-6 md:p-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Вход</h1>
        <p className="text-sm text-slate-500">Используйте email и пароль вашей учётной записи.</p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Email" error={errors.email?.message}>
          <Input {...register("email")} type="email" />
        </FormField>
        <FormField label="Пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" />
        </FormField>
        <div className="flex flex-wrap gap-3">
          <Button type="submit" disabled={isSubmitting}>
            Войти
          </Button>
          <Button type="button" variant="link" onClick={() => navigate("/forgot-password")}>
            Забыли пароль?
          </Button>
        </div>
      </form>
    </div>
  );
}