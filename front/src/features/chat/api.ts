import { apiClient } from "@/shared/api/client";
import type { ListResponse, TournamentMessage } from "@/shared/types/api";

export const chatApi = {
  async list(tournamentId: string, limit = 50, before?: string) {
    const { data } = await apiClient.get<ListResponse<TournamentMessage>>(
      `/tournaments/${tournamentId}/chat`,
      { params: { limit, ...(before ? { before } : {}) } },
    );
    return data;
  },

  async send(tournamentId: string, content: string) {
    const { data } = await apiClient.post<TournamentMessage>(
      `/tournaments/${tournamentId}/chat`,
      { content },
    );
    return data;
  },
};
