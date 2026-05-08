import type {
  ImportBatchStatus,
  ImportRowStatus,
  MatchStatus,
  MatchTeamConfirmationStatus,
  MemberConfirmationStatus,
  NotificationType,
  TeamStatus,
  TournamentFormat,
  TournamentStatus,
  TournamentVisibility,
} from "@/shared/types/api";

export const tournamentStatusLabel: Record<TournamentStatus, string> = {
  draft: "Черновик",
  registration_open: "Регистрация открыта",
  registration_closed: "Регистрация закрыта",
  bracket_generated: "Сетка создана",
  in_progress: "В процессе",
  finished: "Завершён",
  cancelled: "Отменён",
};

export const visibilityLabel: Record<TournamentVisibility, string> = {
  public: "Публичный",
  private: "Приватный",
};

export const tournamentFormatLabel: Record<TournamentFormat, string> = {
  single_elimination: "Single Elimination",
  double_elimination: "Double Elimination",
  group_stage: "Групповой этап + Плей-офф",
};

export const teamStatusLabel: Record<TeamStatus, string> = {
  pending: "Ожидает",
  awaiting_confirmation: "Ожидает подтверждений",
  ready_for_review: "Готова к проверке",
  approved: "Одобрена",
  rejected: "Отклонена",
};

export const memberStatusLabel: Record<MemberConfirmationStatus, string> = {
  found: "Аккаунт найден",
  not_found: "Не найден",
  pending_confirmation: "Ожидает подтверждения",
  confirmed: "Подтверждён",
  declined: "Отклонён",
  removed: "Удалён",
};

export const importBatchStatusLabel: Record<ImportBatchStatus, string> = {
  pending: "Ожидает",
  parsing: "Парсинг",
  preview_ready: "Превью готово",
  confirmed: "Подтверждён",
  failed: "Ошибка",
};

export const importRowStatusLabel: Record<ImportRowStatus, string> = {
  new: "Новая",
  valid: "Валидна",
  needs_review: "Нужно проверить",
  duplicate: "Дубликат",
  rejected: "Отклонена",
  confirmed: "Подтверждена",
};

export const matchStatusLabel: Record<MatchStatus, string> = {
  scheduled: "Запланирован",
  awaiting_confirmation: "Ожидает подтверждения",
  confirmed: "Подтверждён",
  reschedule_requested: "Запрошен перенос",
  issue_reported: "Сообщена проблема",
  in_progress: "В процессе",
  finished: "Завершён",
  cancelled: "Отменён",
};

export const matchTeamConfirmationLabel: Record<MatchTeamConfirmationStatus, string> = {
  pending: "Ожидание",
  ready_confirmed: "Готовность подтверждена",
  reschedule_requested: "Запрошен перенос",
  issue_reported: "Сообщена проблема",
};

export const notificationTypeLabel: Record<NotificationType, string> = {
  added_to_team: "Добавление в команду",
  team_participation_confirmed: "Участие подтверждено",
  team_participation_declined: "Участие отклонено",
  match_assigned: "Назначен матч",
  match_time_changed: "Время матча изменено",
  match_rescheduled: "Матч перенесён",
  match_cancelled: "Матч отменён",
  result_submitted: "Результат отправлен",
  result_confirmed: "Результат подтверждён",
  tournament_finished: "Турнир завершён",
};