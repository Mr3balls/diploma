import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { registerSchema, type RegisterRequest } from "@/features/auth/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { getErrorMessage } from "@/shared/lib/http";

export function RegisterPage() {
  const navigate = useNavigate();
  const { register: registerUser } = useAuth();

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterRequest>({
    resolver: zodResolver(registerSchema),
    defaultValues: { first_name: "", last_name: "", nickname: "", email: "", phone: "", password: "" },
  });

  async function onSubmit(values: RegisterRequest) {
    try {
      await registerUser(values);
      toast.success("Аккаунт создан");
      navigate("/tournaments");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-8">
      <div className="space-y-1">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-[#ff5500]">Новый аккаунт</p>
        <h1 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
          Регистрация
        </h1>
        <p className="text-sm text-[#666666]">Создайте аккаунт, чтобы участвовать в турнирах.</p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Имя" error={errors.first_name?.message}>
            <Input {...register("first_name")} placeholder="Иван" />
          </FormField>
          <FormField label="Фамилия" error={errors.last_name?.message}>
            <Input {...register("last_name")} placeholder="Иванов" />
          </FormField>
        </div>

        <FormField label="Никнейм" error={errors.nickname?.message}>
          <Input {...register("nickname")} placeholder="player123" />
        </FormField>

        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Email" error={errors.email?.message}>
            <Input {...register("email")} type="email" placeholder="example@mail.com" />
          </FormField>
          <FormField label="Телефон" error={errors.phone?.message}>
            <Input {...register("phone")} placeholder="+71234567890" />
          </FormField>
        </div>

        <FormField label="Пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" placeholder="••••••••" />
        </FormField>

        <Button type="submit" disabled={isSubmitting} size="lg" className="w-full mt-2">
          {isSubmitting ? "Создание..." : "Создать аккаунт"}
        </Button>
      </form>

      <p className="text-center text-sm text-[#666666]">
        Уже есть аккаунт?{" "}
        <Link to="/login" className="font-semibold text-[#ff5500] hover:text-[#ff7733] transition-colors">
          Войти
        </Link>
      </p>
    </div>
  );
}
