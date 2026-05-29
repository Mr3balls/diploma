import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { useLang } from "@/app/providers/lang-provider";
import { registerSchema, type RegisterRequest } from "@/features/auth/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Button } from "@/shared/ui/button";
import { getErrorMessage } from "@/shared/lib/http";

export function RegisterPage() {
  const navigate = useNavigate();
  const { register: registerUser } = useAuth();
  const { t } = useLang();

  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterRequest>({
    resolver: zodResolver(registerSchema),
    defaultValues: { first_name: "", last_name: "", nickname: "", email: "", phone: "", password: "" },
  });

  async function onSubmit(values: RegisterRequest) {
    try {
      await registerUser(values);
      toast.success(t("register.success"));
      navigate("/tournaments");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-8">
      <div className="space-y-1">
        <p className="text-xs font-bold uppercase tracking-[0.3em] text-[#ff5500]">{t("register.newAccount")}</p>
        <h1 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
          {t("register.title")}
        </h1>
        <p className="text-sm text-[#666666]">{t("register.desc")}</p>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label={t("register.firstName")} error={errors.first_name?.message}>
            <Input {...register("first_name")} placeholder={t("register.firstNamePlaceholder")} />
          </FormField>
          <FormField label={t("register.lastName")} error={errors.last_name?.message}>
            <Input {...register("last_name")} placeholder={t("register.lastNamePlaceholder")} />
          </FormField>
        </div>

        <FormField label={t("register.nickname")} error={errors.nickname?.message}>
          <Input {...register("nickname")} placeholder="player123" />
        </FormField>

        <div className="grid gap-4 sm:grid-cols-2">
          <FormField label="Email" error={errors.email?.message}>
            <Input {...register("email")} type="email" placeholder="example@mail.com" />
          </FormField>
          <FormField label={t("register.phone")} error={errors.phone?.message}>
            <Input {...register("phone")} placeholder="+71234567890" />
          </FormField>
        </div>

        <FormField label={t("register.password")} error={errors.password?.message}>
          <Input {...register("password")} type="password" placeholder="••••••••" />
        </FormField>

        <Button type="submit" disabled={isSubmitting} size="lg" className="w-full mt-2">
          {isSubmitting ? t("register.submitting") : t("register.submit")}
        </Button>
      </form>

      <p className="text-center text-sm text-[#666666]">
        {t("register.hasAccount")}{" "}
        <Link to="/login" className="font-semibold text-[#ff5500] hover:text-[#ff7733] transition-colors">
          {t("register.signIn")}
        </Link>
      </p>
    </div>
  );
}
