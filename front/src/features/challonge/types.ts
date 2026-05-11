export type ChallongeTournament = {
  id: string;
  title: string;
  discipline?: string;
  description?: string;
  format: string;
  status: "draft" | "ready" | "in_progress" | "completed" | string;
  visibility: string;
  slug?: string;
  max_participants?: number;
  owner_user_id: string;
  created_at: string;
  updated_at?: string;
};

export type Participant = {
  id: string;
  tournament_id: string;
  user_id?: string | null;
  name: string;
  seed: number;
  status: "active" | "eliminated" | "champion";
  final_rank?: number | null;
  created_at?: string;
};

export type ChallongeMatch = {
  id: string;
  tournament_id?: string;
  bracket_id?: string;
  bracket_section?: string;
  round_number?: number | null;
  slot_index?: number | null;
  global_number?: number;
  participant1_id?: string | null;
  participant2_id?: string | null;
  winner_participant_id?: string | null;
  status: string;
  score_text?: string | null;
  is_bye?: boolean;
  next_match_id?: string | null;
  loser_next_match_id?: string | null;
  created_at?: string;
  updated_at?: string;
};

export type ChallongeBracketResponse = {
  tournament: ChallongeTournament;
  matches: ChallongeMatch[];
  participants: Participant[];
  current_user_role: string;
};

export type Standing = {
  rank: number;
  tied: boolean;
  seed: number;
  wins: number;
  losses: number;
};
