import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { useCreateChallonge } from "@/features/challonge/hooks";

const schema = z.object({
  name: z.string().min(2, "Минимум 2 символа").max(200),
  format: z.enum(["single_elimination", "double_elimination"]),
  privacy: z.enum(["public", "private"]),
  max_participants: z.coerce.number().min(0).max(256).optional(),
  slug: z.string().regex(/^[a-z0-9-]*$/, "Только строчные буквы, цифры и дефис").optional().or(z.literal("")),
});

type FormValues = z.infer<typeof schema>;

export function ChallongePage() {
  const navigate = useNavigate();
  const create = useCreateChallonge();
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      format: "single_elimination",
      privacy: "public",
      max_participants: 0,
      slug: "",
    },
  });

  async function onSubmit(values: FormValues) {
    setError(null);
    try {
      const t = await create.mutateAsync({
        name: values.name,
        format: values.format,
        privacy: values.privacy,
        max_participants: values.max_participants || 0,
        slug: values.slug || undefined,
      });
      const slug = t.slug ?? t.id;
      navigate(`/challonge/${slug}`);
    } catch (e: unknown) {
      const msg =
        (e as { response?: { data?: { error?: { message?: string } } } })?.response?.data?.error
          ?.message ?? "Ошибка при создании";
      setError(msg);
    }
  }

  return (
    <div className="page-shell space-y-8 py-8">
      <div>
        <h1 className="text-2xl font-semibold">Новый турнир</h1>
        <p className="mt-1 text-sm text-[#9e9e9e]">
          Создайте сетку, добавьте участников и запустите турнир в один клик
        </p>
      </div>

      <Card className="mx-auto max-w-lg border-[#2d2d2d] bg-[#1a1a1a]">
        <CardHeader>
          <CardTitle>Параметры турнира</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <FormField label="Название" error={errors.name?.message}>
              <Input placeholder="Например: ACE Summer Cup" {...register("name")} />
            </FormField>

            <FormField label="Формат" error={errors.format?.message}>
              <select
                {...register("format")}
                className="w-full rounded-xl border border-[#2d2d2d] bg-[#111111] px-3 py-2 text-sm text-white"
              >
                <option value="single_elimination">Single Elimination</option>
                <option value="double_elimination">Double Elimination</option>
              </select>
            </FormField>

            <FormField label="Приватность" error={errors.privacy?.message}>
              <select
                {...register("privacy")}
                className="w-full rounded-xl border border-[#2d2d2d] bg-[#111111] px-3 py-2 text-sm text-white"
              >
                <option value="public">Публичный</option>
                <option value="private">Приватный</option>
              </select>
            </FormField>

            <FormField
              label="Макс. участников (0 = без ограничений)"
              error={errors.max_participants?.message}
            >
              <Input type="number" min={0} max={256} {...register("max_participants")} />
            </FormField>

            <FormField
              label="URL-слаг (необязательно)"
              error={errors.slug?.message}
              hint="Оставьте пустым для автогенерации. Только строчные буквы, цифры, дефис."
            >
              <Input placeholder="my-tournament" {...register("slug")} />
            </FormField>

            {error && (
              <p className="rounded-xl bg-red-900/30 px-3 py-2 text-sm text-red-400">{error}</p>
            )}

            <Button type="submit" className="w-full" disabled={isSubmitting || create.isPending}>
              {create.isPending ? "Создание..." : "Создать и добавить участников →"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
