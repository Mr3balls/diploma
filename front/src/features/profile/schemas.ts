import { z } from "zod";

export const profileSchema = z.object({
  first_name: z.string().max(100).optional().default(""),
  last_name: z.string().max(100).optional().default(""),
  nickname: z.string().min(2, "Никнейм не может быть пустым").max(50).optional().default(""),
  phone: z
    .string()
    .regex(/^(\+7|8)\d{10}$/, "Формат: +7XXXXXXXXXX или 8XXXXXXXXXX")
    .or(z.literal(""))
    .optional()
    .default(""),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;