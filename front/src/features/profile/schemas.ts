import { z } from "zod";

export const profileSchema = z.object({
  first_name: z.string().min(1, "Введите имя"),
  last_name: z.string().min(1, "Введите фамилию"),
  nickname: z.string().min(2, "Введите никнейм"),
  phone: z.string().min(5, "Введите телефон"),
});

export type ProfileFormValues = z.infer<typeof profileSchema>;