import { useNavigate } from "react-router-dom";
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

  const form = useForm<RegisterRequest>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      first_name: "",
      last_name: "",
      nickname: "",
      email: "",
      phone: "",
      password: "",
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form;

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
    <div className="grid gap-6 p-6 md:p-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Регистрация</h1>
        <p className="text-sm text-slate-500">После регистрации пользователь сможет создавать турниры и участвовать в них.</p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 md:grid-cols-2">
          <FormField label="Имя" error={errors.first_name?.message}>
            <Input {...register("first_name")} />
          </FormField>
          <FormField label="Фамилия" error={errors.last_name?.message}>
            <Input {...register("last_name")} />
          </FormField>
        </div>
        <FormField label="Никнейм" error={errors.nickname?.message}>
          <Input {...register("nickname")} />
        </FormField>
        <div className="grid gap-4 md:grid-cols-2">
          <FormField label="Email" error={errors.email?.message}>
            <Input {...register("email")} type="email" placeholder="example@mail.com" />
            <p className="text-xs text-slate-500 mt-1">Обязательно содержит символ @</p>
          </FormField>
          <FormField label="Телефон" error={errors.phone?.message}>
            <Input {...register("phone")} placeholder="+71234567890" />
            <p className="text-xs text-slate-500 mt-1">Формат: +7XXXXXXXXXX или 8XXXXXXXXXX</p>
          </FormField>
        </div>
        <FormField label="Пароль" error={errors.password?.message}>
          <Input {...register("password")} type="password" />
        </FormField>
        <div className="flex flex-wrap gap-3">
          <Button type="submit" disabled={isSubmitting}>
            Создать аккаунт
          </Button>
          <Button type="button" variant="link" onClick={() => navigate("/login")}>
            Уже есть аккаунт?
          </Button>
        </div>
      </form>
    </div>
  );
}