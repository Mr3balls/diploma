import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().email("Введите корректный email"),
  password: z.string().min(6, "Минимум 6 символов"),
});

export type LoginRequest = z.infer<typeof loginSchema>;

export const registerSchema = z.object({
  first_name: z.string().min(1, "Введите имя"),
  last_name: z.string().min(1, "Введите фамилию"),
  nickname: z.string().min(2, "Введите никнейм"),
  email: z.string().email("Введите корректный email"),
  phone: z
    .string()
    .min(1, "Введите телефон")
    .regex(/^(\+7|8)\d{10}$/, "Формат: +7XXXXXXXXXX или 8XXXXXXXXXX (10 цифр после кода)"),
  password: z.string().min(8, "Минимум 8 символов"),
});

export type RegisterRequest = z.infer<typeof registerSchema>;

export const forgotPasswordSchema = z.object({
  email: z.string().email("Введите корректный email"),
});

export type ForgotPasswordRequest = z.infer<typeof forgotPasswordSchema>;

export const resetPasswordSchema = z.object({
  token: z.string().min(1, "Введите токен"),
  password: z.string().min(8, "Минимум 8 символов"),
});

export type ResetPasswordRequest = z.infer<typeof resetPasswordSchema>;