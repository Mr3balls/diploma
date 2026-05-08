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
  const form = useForm<ForgotPasswordRequest>({
    resolver: zodResolver(forgotPasswordSchema),
    defaultValues: {
      email: "",
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = form;

  async function onSubmit(values: ForgotPasswordRequest) {
    try {
      const response = await authApi.forgotPassword(values);
      toast.success("Демо-ответ получен");
      console.info("Forgot password demo response", response);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-6 p-6 md:p-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Восстановление пароля</h1>
        <p className="text-sm text-[#90afd4]">
          Backend возвращает демонстрационный ответ. Реальный email-flow в текущем MVP не эмулируется.
        </p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <FormField label="Email" error={errors.email?.message}>
          <Input {...register("email")} type="email" />
        </FormField>
        <Button type="submit" disabled={isSubmitting}>
          Отправить запрос
        </Button>
      </form>
    </div>
  );
}