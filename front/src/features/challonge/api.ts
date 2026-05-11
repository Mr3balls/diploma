import { apiClient } from "@/shared/api/client";
import type {
  ChallongeBracketResponse,
  ChallongeTournament,
  Participant,
  Standing,
} from "./types";

export type CreateChallongePayload = {
  name: string;
  format: "single_elimination" | "double_elimination";
  privacy: "public" | "private";
  max_participants?: number;
  description?: string;
  slug?: string;
};

export const challongeApi = {
  async create(payload: CreateChallongePayload): Promise<ChallongeTournament> {
    const { data } = await apiClient.post<ChallongeTournament>("/c", {
      ...payload,
      game: "",
    });
    return data;
  },

  async getBracket(slug: string): Promise<ChallongeBracketResponse> {
    const { data } = await apiClient.get<ChallongeBracketResponse>(`/c/${slug}`);
    return data;
  },

  async addParticipant(slug: string, name: string): Promise<Participant> {
    const { data } = await apiClient.post<Participant>(`/c/${slug}/participants`, { name });
    return data;
  },

  async bulkAdd(slug: string, names: string[]): Promise<{ participants: Participant[] }> {
    const { data } = await apiClient.post<{ participants: Participant[] }>(
      `/c/${slug}/participants/bulk`,
      { names },
    );
    return data;
  },

  async removeParticipant(slug: string, participantID: string): Promise<void> {
    await apiClient.delete(`/c/${slug}/participants/${participantID}`);
  },

  async shuffle(slug: string): Promise<void> {
    await apiClient.post(`/c/${slug}/participants/shuffle`);
  },

  async start(slug: string): Promise<void> {
    await apiClient.post(`/c/${slug}/start`);
  },

  async reset(slug: string): Promise<void> {
    await apiClient.post(`/c/${slug}/reset`);
  },

  async submitResult(
    slug: string,
    matchID: string,
    winnerID: string,
    score1 = 0,
    score2 = 0,
  ): Promise<void> {
    await apiClient.post(`/c/${slug}/matches/${matchID}/result`, {
      winner_id: winnerID,
      score1,
      score2,
    });
  },

  async resetMatch(slug: string, matchID: string): Promise<void> {
    await apiClient.post(`/c/${slug}/matches/${matchID}/reset`);
  },

  async getStandings(slug: string): Promise<{ standings: Standing[] }> {
    const { data } = await apiClient.get<{ standings: Standing[] }>(`/c/${slug}/standings`);
    return data;
  },

  async join(slug: string): Promise<void> {
    await apiClient.post(`/c/${slug}/join`);
  },
};
