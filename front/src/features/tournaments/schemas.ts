import { z } from "zod";

export const tournamentFormSchema = z
  .object({
    title: z.string().min(2, "Введите название"),
    discipline: z.string().optional(),
    description: z.string().optional(),
    rules: z.string().optional(),
    location: z.string().optional(),
    max_teams: z.coerce.number().min(2, "Минимум 2").max(128, "Максимум 128").optional(),
    format: z.enum(["single_elimination", "double_elimination", "group_stage", "group_de"]),
    group_count: z.coerce.number().min(2).max(4).optional(),
    registration_deadline: z.string().optional(),
    start_at: z.string().optional(),
    visibility: z.enum(["public", "private"]),
    registration_mode: z.enum(["team", "individual"]).default("team"),
  })
  .refine(
    (data) => {
      if (data.format === "group_stage" || data.format === "group_de") {
        return data.group_count && [2, 3, 4].includes(data.group_count);
      }
      return true;
    },
    { message: "Выберите количество групп: 2, 3 или 4", path: ["group_count"] },
  );

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
    "ready",
    "completed",
  ]),
});

export type TournamentStatusFormValues = z.infer<typeof tournamentStatusSchema>;

export const managerSchema = z.object({
  user_id: z.string().min(1, "Введите user_id"),
});

export type ManagerFormValues = z.infer<typeof managerSchema>;