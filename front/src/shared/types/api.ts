export type TournamentStatus =
  | "draft"
  | "registration_open"
  | "registration_closed"
  | "bracket_generated"
  | "in_progress"
  | "finished"
  | "cancelled";

export type TournamentVisibility = "public" | "private";

export type TeamStatus =
  | "pending"
  | "awaiting_confirmation"
  | "ready_for_review"
  | "approved"
  | "rejected";

export type MemberConfirmationStatus =
  | "found"
  | "not_found"
  | "pending_confirmation"
  | "confirmed"
  | "declined"
  | "removed";

export type ImportBatchStatus = "pending" | "parsing" | "preview_ready" | "confirmed" | "failed";

export type ImportRowStatus =
  | "new"
  | "valid"
  | "needs_review"
  | "duplicate"
  | "rejected"
  | "confirmed";

export type MatchStatus =
  | "scheduled"
  | "awaiting_confirmation"
  | "confirmed"
  | "reschedule_requested"
  | "issue_reported"
  | "in_progress"
  | "finished"
  | "cancelled";

export type MatchTeamConfirmationStatus =
  | "pending"
  | "ready_confirmed"
  | "reschedule_requested"
  | "issue_reported";

export type NotificationType =
  | "added_to_team"
  | "team_participation_confirmed"
  | "team_participation_declined"
  | "match_assigned"
  | "match_time_changed"
  | "match_rescheduled"
  | "match_cancelled"
  | "result_submitted"
  | "result_confirmed"
  | "tournament_finished";

export type GenericMessageResponse = {
  message: string;
};

export type UnreadCountResponse = {
  count: number;
};

export type ApiErrorResponse = {
  error: {
    code: string;
    message: string;
    details?: unknown;
  };
};

export type TokenPair = {
  access_token: string;
  refresh_token: string;
};

export type User = {
  id: string;
  email: string;
  nickname?: string | null;
  first_name?: string | null;
  last_name?: string | null;
  phone?: string | null;
  role?: "player" | "platform_admin" | string;
  is_blocked?: boolean;
  is_platform_admin?: boolean;
  created_at?: string;
  updated_at?: string;
};

export type Tournament = {
  id: string;
  title: string;
  discipline: string;
  description?: string;
  rules?: string;
  location?: string;
  max_teams: number;
  registration_deadline?: string;
  start_at?: string;
  status: string;
  visibility: "public" | "private";
  owner_user_id: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
};

export type GoogleSheetLink = {
  id?: string;
  tournament_id?: string;
  sheet_url: string;
  worksheet_name: string;
  created_at?: string;
};

export type ImportBatch = {
  id: string;
  tournament_id?: string;
  status: ImportBatchStatus;
  spreadsheet_id?: string | null;
  worksheet_name?: string | null;
  row_count?: number;
  created_at?: string;
  updated_at?: string;
};

export type ImportRow = {
  id: string;
  batch_id?: string;
  status: ImportRowStatus;
  team_name?: string | null;
  captain_nickname?: string | null;
  discipline?: string | null;
  player_nicknames?: string[];
  duplicate_conflicts?: string[];
  validation_errors?: string[];
};

export type TeamMember = {
  id: string;
  team_id?: string;
  user_id?: string | null;
  nickname?: string | null;
  role: "captain" | "player" | "substitute" | string;
  confirmation_status: MemberConfirmationStatus;
  display_name?: string | null;
};

export type Team = {
  id: string;
  tournament_id?: string;
  name: string;
  status: TeamStatus;
  captain_user_id?: string | null;
  captain_nickname?: string | null;
  duplicate_conflicts?: string[];
  member_count?: number;
  confirmed_main_players_count?: number;
  created_at?: string;
};

export type Bracket = {
  id?: string;
  tournament_id?: string;
  type?: "single_elimination" | string;
  generated_at?: string | null;
};

export type Match = {
  id: string;
  tournament_id?: string;
  round_number?: number | null;
  slot_index?: number | null;
  status: MatchStatus;
  home_team_id?: string | null;
  away_team_id?: string | null;
  home_team_name?: string | null;
  away_team_name?: string | null;
  home_team?: Team | null;
  away_team?: Team | null;
  home_team_confirmation_status?: MatchTeamConfirmationStatus | null;
  away_team_confirmation_status?: MatchTeamConfirmationStatus | null;
  winner_team_id?: string | null;
  score_text?: string | null;
  scheduled_at?: string | null;
  is_bye?: boolean;
  created_at?: string;
  updated_at?: string;
};

export type Notification = {
  id: string;
  type: NotificationType;
  title?: string | null;
  message: string;
  is_read?: boolean;
  created_at?: string;
  team_member_id?: string | null;
  match_id?: string | null;
  tournament_id?: string | null;
  payload?: Record<string, unknown> | null;
};

export type AuditLog = {
  id: string;
  action: string;
  actor_user_id?: string | null;
  actor_email?: string | null;
  created_at?: string;
  details?: Record<string, unknown> | null;
};

export type AuthResponse = {
  user: User;
  tokens: TokenPair;
};

export type ListResponse<T> = {
  items: T[];
};

export type TournamentBracketResponse = {
  bracket: Bracket;
  matches: Match[];
};

export type ImportPreviewResponse = {
  batch: ImportBatch;
  rows: ImportRow[];
};

export type ImportBatchDetailsResponse = {
  batch: ImportBatch;
  rows: ImportRow[];
};

export type TeamDetailsResponse = {
  team: Team;
  members: TeamMember[];
};

export type ValidateGoogleSheetResponse = {
  status: string;
  spreadsheet_id: string;
  worksheet_name: string;
  row_count: number;
  sample_row: string[];
};