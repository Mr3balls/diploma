import { useForm, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  tournamentFormSchema,
  type TournamentFormValues,
} from "@/features/tournaments/schemas";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";
import { Textarea } from "@/shared/ui/textarea";
import { Select } from "@/shared/ui/select";
import { Button } from "@/shared/ui/button";

function toDateTimeLocal(value?: string) {
  if (!value) return "";

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";

  const offset = date.getTimezoneOffset();
  const localDate = new Date(date.getTime() - offset * 60_000);

  return localDate.toISOString().slice(0, 16);
}

export function CreateTournamentForm({
  defaultValues,
  onSubmit,
  submitLabel,
  isSubmitting,
}: {
  defaultValues?: Partial<TournamentFormValues>;
  onSubmit: (values: TournamentFormValues) => void;
  submitLabel: string;
  isSubmitting?: boolean;
}) {
  const form = useForm<TournamentFormValues>({
    resolver: zodResolver(tournamentFormSchema),
    defaultValues: {
      title: defaultValues?.title ?? "",
      discipline: defaultValues?.discipline ?? "",
      description: defaultValues?.description ?? "",
      rules: defaultValues?.rules ?? "",
      location: defaultValues?.location ?? "",
      max_teams: defaultValues?.max_teams ?? 8,
      format: (defaultValues?.format as TournamentFormValues["format"]) ?? "single_elimination",
      group_count: defaultValues?.group_count ?? undefined,
      registration_deadline: toDateTimeLocal(defaultValues?.registration_deadline),
      start_at: toDateTimeLocal(defaultValues?.start_at),
      visibility: defaultValues?.visibility ?? "public",
    },
  });

  const { register, handleSubmit, formState, control } = form;
  const { errors } = formState;
  const selectedFormat = useWatch({ control, name: "format" });

  return (
    <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
      <FormField label="Название" error={errors.title?.message}>
        <Input {...register("title")} />
      </FormField>

      <FormField label="Дисциплина" error={errors.discipline?.message}>
        <Input
          {...register("discipline")}
          placeholder="Например, CS2, Dota 2, Valorant"
        />
      </FormField>

      <FormField label="Локация" error={errors.location?.message}>
        <Input
          {...register("location")}
          placeholder="Например, Astana IT University"
        />
      </FormField>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField label="Максимум команд" error={errors.max_teams?.message} hint="От 2 до 16">
          <Input
            type="number"
            min={2}
            max={16}
            {...register("max_teams", { valueAsNumber: true })}
          />
        </FormField>

        <FormField label="Видимость" error={errors.visibility?.message}>
          <Select {...register("visibility")}>
            <option value="public">Публичный</option>
            <option value="private">Приватный</option>
          </Select>
        </FormField>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField label="Формат сетки" error={errors.format?.message}>
          <Select {...register("format")}>
            <option value="single_elimination">Single Elimination</option>
            <option value="double_elimination">Double Elimination</option>
            <option value="group_stage">Групповой этап + Плей-офф</option>
          </Select>
        </FormField>

        {selectedFormat === "group_stage" && (
          <FormField label="Количество групп" error={errors.group_count?.message}>
            <Select {...register("group_count", { valueAsNumber: true })}>
              <option value="">— выберите —</option>
              <option value="2">2 группы (по 4 команды)</option>
              <option value="3">3 группы (по 4 команды)</option>
              <option value="4">4 группы (по 4 команды)</option>
            </Select>
          </FormField>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField
          label="Дедлайн регистрации"
          error={errors.registration_deadline?.message}
        >
          <Input type="datetime-local" {...register("registration_deadline")} />
        </FormField>

        <FormField label="Дата начала" error={errors.start_at?.message}>
          <Input type="datetime-local" {...register("start_at")} />
        </FormField>
      </div>

      <FormField label="Описание" error={errors.description?.message}>
        <Textarea {...register("description")} />
      </FormField>

      <FormField label="Правила" error={errors.rules?.message}>
        <Textarea {...register("rules")} />
      </FormField>

      <div>
        <Button type="submit" disabled={isSubmitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  );
}