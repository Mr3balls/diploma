import { z } from "zod";

export const scheduleMatchSchema = z.object({
  scheduled_at: z.string().min(1, "Выберите дату и время"),
});

export type ScheduleMatchValues = z.infer<typeof scheduleMatchSchema>;

export const requestReasonSchema = z.object({
  reason: z.string().min(1, "Введите причину"),
});

export type RequestReasonValues = z.infer<typeof requestReasonSchema>;

export const submitResultSchema = z.object({
  winner_team_id: z.string().min(1, "Введите winner_team_id"),
  score_text: z.string().min(1, "Введите score_text"),
});

export type SubmitResultValues = z.infer<typeof submitResultSchema>;