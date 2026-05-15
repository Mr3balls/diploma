import { apiClient } from "@/shared/api/client";
import type { GenericMessageResponse, ListResponse, Team, TeamDetailsResponse } from "@/shared/types/api";

export const teamsApi = {
  async getTournamentTeams(tournamentId: string) {
    const { data } = await apiClient.get<ListResponse<Team>>(`/tournaments/${tournamentId}/teams`);
    return data;
  },
  async getAdminTournamentTeams(tournamentId: string) {
    const { data } = await apiClient.get<ListResponse<Team>>(`/tournaments/${tournamentId}/admin/teams`);
    return data;
  },
  async getTeam(teamId: string) {
    const { data } = await apiClient.get<TeamDetailsResponse>(`/teams/${teamId}`);
    return data;
  },
  async updateTeam(teamId: string, payload: Record<string, unknown>) {
    const { data } = await apiClient.patch<TeamDetailsResponse>(`/teams/${teamId}`, payload);
    return data;
  },
  async approveTeam(teamId: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/teams/${teamId}/approve`);
    return data;
  },
  async rejectTeam(teamId: string, reason: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/teams/${teamId}/reject`, { reason });
    return data;
  },
  async removeMember(teamId: string, memberId: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/teams/${teamId}/members/${memberId}/remove`);
    return data;
  },
  async replaceMember(teamId: string, memberId: string, payload: Record<string, unknown>) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/teams/${teamId}/members/${memberId}/replace`, payload);
    return data;
  },
  async adminDeleteTeam(tournamentId: string, teamId: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>(
      `/tournaments/${tournamentId}/admin/teams/${teamId}`,
    );
    return data;
  },
  async adminCreateTeam(tournamentId: string, payload: { team_name: string; members: string[] }) {
    const { data } = await apiClient.post<TeamDetailsResponse>(
      `/tournaments/${tournamentId}/admin/teams`,
      payload,
    );
    return data;
  },
  async getMyTeam(tournamentId: string) {
    const { data } = await apiClient.get<TeamDetailsResponse>(`/tournaments/${tournamentId}/my-team`);
    return data;
  },
  async acceptParticipation(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/team-members/${id}/accept`);
    return data;
  },
  async declineParticipation(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/team-members/${id}/decline`);
    return data;
  },
};