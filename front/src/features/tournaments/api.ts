import { apiClient } from "@/shared/api/client";
import type {
  AuditLog,
  GenericMessageResponse,
  ListResponse,
  TeamDetailsResponse,
  Tournament,
} from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";
import type {
  ManagerFormValues,
  TournamentFormValues,
  TournamentStatusFormValues,
} from "@/features/tournaments/schemas";

function trimOrUndefined(value?: string) {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

function toIsoOrUndefined(value?: string) {
  if (!value) return undefined;

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return undefined;
  }

  return date.toISOString();
}

function sanitizeTournamentPayload(payload: TournamentFormValues) {
  return {
    title: payload.title.trim(),
    discipline: payload.discipline?.trim() ?? "",
    description: trimOrUndefined(payload.description),
    rules: trimOrUndefined(payload.rules),
    location: trimOrUndefined(payload.location),
    max_teams: payload.max_teams ?? 8,
    registration_deadline: toIsoOrUndefined(payload.registration_deadline),
    start_at: toIsoOrUndefined(payload.start_at),
    visibility: payload.visibility,
    registration_mode: payload.registration_mode ?? "team",
  };
}

export const tournamentsApi = {
  async list() {
    const { data } = await apiClient.get<ListResponse<Tournament>>("/tournaments");
    return data;
  },

  async getById(id: string) {
    const { data } = await apiClient.get<Tournament>(`/tournaments/${id}`);
    return data;
  },

  async create(payload: TournamentFormValues) {
    const { data } = await apiClient.post<Tournament>(
      "/tournaments",
      sanitizeTournamentPayload(payload),
    );
    return data;
  },

  async update(id: string, payload: TournamentFormValues) {
    const { data } = await apiClient.patch<Tournament>(
      `/tournaments/${id}`,
      sanitizeTournamentPayload(payload),
    );
    return data;
  },

  async remove(id: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>(`/tournaments/${id}`);
    return data;
  },

  async changeStatus(id: string, payload: TournamentStatusFormValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(
      `/tournaments/${id}/status`,
      payload,
    );
    return data;
  },

  async addManager(id: string, payload: ManagerFormValues) {
    const { data } = await apiClient.post<GenericMessageResponse>(
      `/tournaments/${id}/managers`,
      payload,
    );
    return data;
  },

  async removeManager(id: string, userId: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>(
      `/tournaments/${id}/managers/${userId}`,
    );
    return data;
  },

  async getAudit(id: string) {
    const { data } = await apiClient.get<ListResponse<AuditLog>>(
      `/tournaments/${id}/audit`,
    );
    return data;
  },

  async getParticipants(id: string) {
    const { data } = await apiClient.get<ListResponse<Participant>>(
      `/tournaments/${id}/participants`,
    );
    return data;
  },

  async addParticipant(id: string, name: string) {
    const { data } = await apiClient.post<Participant>(
      `/tournaments/${id}/participants`,
      { name },
    );
    return data;
  },

  async bulkAddParticipants(id: string, names: string[]) {
    const { data } = await apiClient.post<ListResponse<Participant>>(
      `/tournaments/${id}/participants/bulk`,
      { names },
    );
    return data;
  },

  async removeParticipant(id: string, participantId: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>(
      `/tournaments/${id}/participants/${participantId}`,
    );
    return data;
  },

  async shuffleParticipants(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(
      `/tournaments/${id}/participants/shuffle`,
    );
    return data;
  },

  async startBracket(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(
      `/tournaments/${id}/start-bracket`,
    );
    return data;
  },

  async joinIndividual(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(
      `/tournaments/${id}/join`,
    );
    return data;
  },

  async registerTeam(id: string, payload: { team_name: string; members: string[] }) {
    const { data } = await apiClient.post<TeamDetailsResponse>(
      `/tournaments/${id}/register-team`,
      payload,
    );
    return data;
  },
};