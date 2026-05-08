import { apiClient } from "@/shared/api/client";
import type { TournamentBracketResponse } from "@/shared/types/api";

export const bracketApi = {
  async getByTournamentId(tournamentId: string) {
    const { data } = await apiClient.get<TournamentBracketResponse>(`/tournaments/${tournamentId}/bracket`);
    return data;
  },
  async generate(tournamentId: string) {
    const { data } = await apiClient.post<TournamentBracketResponse>(`/tournaments/${tournamentId}/bracket/generate`);
    return data;
  },
  async regenerate(tournamentId: string) {
    const { data } = await apiClient.post<TournamentBracketResponse>(`/tournaments/${tournamentId}/bracket/regenerate`);
    return data;
  },
  async reseed(tournamentId: string, payload: { ordered_team_ids: string[] }) {
    const { data } = await apiClient.post<TournamentBracketResponse>(`/tournaments/${tournamentId}/bracket/reseed`, payload);
    return data;
  },
};