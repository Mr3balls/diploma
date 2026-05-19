import { useEffect, useState } from "react";
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
import { MapPicker } from "@/shared/ui/map-picker";

export function CreateTournamentForm({
  defaultValues,
  onSubmit,
  submitLabel,
  isSubmitting,
  showAdvanced = false,
}: {
  defaultValues?: Partial<TournamentFormValues>;
  onSubmit: (values: TournamentFormValues) => void;
  submitLabel: string;
  isSubmitting?: boolean;
  showAdvanced?: boolean;
}) {
  const [advanced, setAdvanced] = useState(showAdvanced);

  const form = useForm<TournamentFormValues>({
    resolver: zodResolver(tournamentFormSchema),
    defaultValues: {
      title: defaultValues?.title ?? "",
      discipline: defaultValues?.discipline ?? "",
      description: defaultValues?.description ?? "",
      rules: defaultValues?.rules ?? "",
      location: defaultValues?.location ?? "",
      latitude: defaultValues?.latitude ?? undefined,
      longitude: defaultValues?.longitude ?? undefined,
      max_teams: defaultValues?.max_teams ?? 8,
      format: (defaultValues?.format as TournamentFormValues["format"]) ?? "single_elimination",
      group_count: defaultValues?.group_count ?? undefined,
      registration_deadline: defaultValues?.registration_deadline ?? "",
      start_at: defaultValues?.start_at ?? "",
      visibility: defaultValues?.visibility ?? "public",
      registration_mode: (defaultValues?.registration_mode as TournamentFormValues["registration_mode"]) ?? "team",
    },
  });

  const { register, handleSubmit, formState, control, setValue } = form;
  const { errors } = formState;
  const selectedFormat = useWatch({ control, name: "format" });

  useEffect(() => {
    if (selectedFormat === "group_stage" || selectedFormat === "group_de") {
      const current = form.getValues("group_count");
      if (!current || ![2, 3, 4].includes(current)) {
        setValue("group_count", 2);
      }
    }
  }, [selectedFormat, setValue, form]);

  return (
    <form className="grid gap-4" onSubmit={handleSubmit(onSubmit)}>
      <FormField label="Название" error={errors.title?.message}>
        <Input {...register("title")} placeholder="Например: Летний кубок ACE" />
      </FormField>

      <FormField label="Тип регистрации" error={errors.registration_mode?.message}>
        <Select {...register("registration_mode")}>
          <option value="team">Командный (5v5, 2v2 и т.д.)</option>
          <option value="individual">Индивидуальный (1v1, по именам)</option>
        </Select>
      </FormField>

      <div className="grid gap-4 md:grid-cols-2">
        <FormField label="Формат сетки" error={errors.format?.message}>
          <Select {...register("format")}>
            <option value="single_elimination">Single Elimination</option>
            <option value="double_elimination">Double Elimination</option>
            <option value="group_stage">Групповой этап + Плей-офф</option>
            <option value="group_de">Групп. DE + Плей-офф</option>
          </Select>
        </FormField>

        <FormField label="Видимость" error={errors.visibility?.message}>
          <Select {...register("visibility")}>
            <option value="public">Публичный</option>
            <option value="private">Приватный</option>
          </Select>
        </FormField>
      </div>

      {(selectedFormat === "group_stage" || selectedFormat === "group_de") && (
        <FormField
          label="Количество групп"
          error={errors.group_count?.message}
          hint={
            selectedFormat === "group_de"
              ? "Топ 3 из каждой группы (WB-финал победитель/проигравший + LB-финал победитель) выходят в плей-офф"
              : "Топ 2 из каждой группы выходят в плей-офф"
          }
        >
          <Select {...register("group_count", { valueAsNumber: true })}>
            <option value="2">2 группы</option>
            <option value="3">3 группы</option>
            <option value="4">4 группы</option>
          </Select>
        </FormField>
      )}

      <button
        type="button"
        onClick={() => setAdvanced((v) => !v)}
        className="text-left text-xs text-[#666666] hover:text-white transition-colors"
      >
        {advanced ? "▲ Скрыть дополнительные параметры" : "▼ Дополнительные параметры"}
      </button>

      {advanced && (
        <>
          <FormField label="Дисциплина" error={errors.discipline?.message}>
            <Input {...register("discipline")} placeholder="CS2, Dota 2, Valorant…" />
          </FormField>

          <FormField label="Макс. команд / участников" error={errors.max_teams?.message}>
            <Input type="number" min={2} max={128} {...register("max_teams", { valueAsNumber: true })} />
          </FormField>

          <FormField label="Место проведения (адрес)" error={errors.location?.message}>
            <Input {...register("location")} placeholder="Например: Алматы, ул. Абая 1" />
          </FormField>

          <FormField label="Место проведения на карте">
            <MapPicker
              value={
                form.getValues("latitude") != null && form.getValues("longitude") != null
                  ? { lat: form.getValues("latitude")!, lng: form.getValues("longitude")! }
                  : null
              }
              onChange={(coords) => {
                setValue("latitude", coords?.lat ?? undefined);
                setValue("longitude", coords?.lng ?? undefined);
              }}
            />
          </FormField>

          <FormField label="Описание" error={errors.description?.message}>
            <Textarea {...register("description")} />
          </FormField>

          <FormField label="Правила" error={errors.rules?.message}>
            <Textarea {...register("rules")} />
          </FormField>
        </>
      )}

      <div>
        <Button type="submit" disabled={isSubmitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  );
}
