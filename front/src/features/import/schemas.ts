import { z } from "zod";

export const googleSheetSchema = z.object({
  sheet_url: z.string().url("Введите корректную ссылку"),
  worksheet_name: z.string().min(1, "Введите имя листа"),
});

export type GoogleSheetFormValues = z.infer<typeof googleSheetSchema>;