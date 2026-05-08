import { apiClient } from "@/shared/api/client";
import type { GenericMessageResponse, ListResponse, Match } from "@/shared/types/api";
import type {
  RequestReasonValues,
  ScheduleMatchValues,
  SubmitResultValues,
} from "@/features/matches/schemas";

export const matchesApi = {
  async getPublicByTournamentId(tournamentId: string) {
    const { data } = await apiClient.get<ListResponse<Match>>(`/tournaments/${tournamentId}/matches`);
    return data;
  },
  async getAdminByTournamentId(tournamentId: string) {
    const { data } = await apiClient.get<ListResponse<Match>>(`/tournaments/${tournamentId}/admin/matches`);
    return data;
  },
  async schedule(matchId: string, payload: ScheduleMatchValues) {
    const { data } = await apiClient.patch<GenericMessageResponse>(`/matches/${matchId}/schedule`, payload);
    return data;
  },
  async confirmReady(matchId: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/confirm-ready`);
    return data;
  },
  async requestReschedule(matchId: string, payload: RequestReasonValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/request-reschedule`, payload);
    return data;
  },
  async reportIssue(matchId: string, payload: RequestReasonValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/report-issue`, payload);
    return data;
  },
  async submitResult(matchId: string, payload: SubmitResultValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/submit-result`, payload);
    return data;
  },
  async approveResult(matchId: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/approve-result`);
    return data;
  },
  async rejectResult(matchId: string, payload?: RequestReasonValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/matches/${matchId}/reject-result`, payload ?? {});
    return data;
  },
};