import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { googleSheetSchema, type GoogleSheetFormValues } from "@/features/import/schemas";
import { Button } from "@/shared/ui/button";
import { FormField } from "@/shared/ui/form-field";
import { Input } from "@/shared/ui/input";

export function GoogleSheetForm({
  onConnect,
  onValidate,
  onPreview,
  defaultValues,
  isBusy,
}: {
  onConnect: (values: GoogleSheetFormValues) => void;
  onValidate: (values: GoogleSheetFormValues) => void;
  onPreview: (values: GoogleSheetFormValues) => void;
  defaultValues?: Partial<GoogleSheetFormValues>;
  isBusy?: boolean;
}) {
  const form = useForm<GoogleSheetFormValues>({
    resolver: zodResolver(googleSheetSchema),
    defaultValues: {
      sheet_url: defaultValues?.sheet_url ?? "",
      worksheet_name: defaultValues?.worksheet_name ?? "",
    },
  });

  const values = form.watch();
  const { register, handleSubmit, formState } = form;
  const { errors } = formState;

  return (
    <form className="grid gap-4" onSubmit={handleSubmit(onPreview)}>
      <FormField
        label="Ссылка на публичную таблицу"
        hint="Вставьте public/edit/published URL. Авторизация Google в фронте не используется."
        error={errors.sheet_url?.message}
      >
        <Input {...register("sheet_url")} placeholder="https://docs.google.com/spreadsheets/..." />
      </FormField>
      <FormField label="Имя листа (worksheet_name)" error={errors.worksheet_name?.message}>
        <Input {...register("worksheet_name")} placeholder="Sheet1" />
      </FormField>
      <div className="flex flex-wrap gap-2">
        <Button type="button" variant="outline" disabled={isBusy} onClick={handleSubmit(onConnect)}>
          Сохранить связь
        </Button>
        <Button type="button" variant="secondary" disabled={isBusy} onClick={handleSubmit(onValidate)}>
          Проверить
        </Button>
        <Button type="submit" disabled={isBusy || !values.sheet_url || !values.worksheet_name}>
          Превью импорта
        </Button>
      </div>
    </form>
  );
}