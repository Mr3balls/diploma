import { z } from "zod";

export const tournamentFormSchema = z.object({
  title: z.string().min(2, "Введите название"),
  discipline: z.string().min(1, "Введите дисциплину"),
  description: z.string().optional(),
  rules: z.string().optional(),
  location: z.string().optional(),
  max_teams: z.coerce.number().min(2, "Минимум 2").max(128, "Максимум 128"),
  registration_deadline: z.string().optional(),
  start_at: z.string().optional(),
  visibility: z.enum(["public", "private"]),
});

export type TournamentFormValues = z.infer<typeof tournamentFormSchema>;

export const tournamentStatusSchema = z.object({
  status: z.enum([
    "draft",
    "registration_open",
    "registration_closed",
    "bracket_generated",
    "in_progress",
    "finished",
    "cancelled",
  ]),
});

export type TournamentStatusFormValues = z.infer<typeof tournamentStatusSchema>;

export const managerSchema = z.object({
  user_id: z.string().min(1, "Введите user_id"),
});

export type ManagerFormValues = z.infer<typeof managerSchema>;