import { apiClient } from "@/shared/api/client";
import type {
  AuditLog,
  GenericMessageResponse,
  ListResponse,
  Tournament,
} from "@/shared/types/api";
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
    discipline: payload.discipline.trim(),
    description: trimOrUndefined(payload.description),
    rules: trimOrUndefined(payload.rules),
    location: trimOrUndefined(payload.location),
    max_teams: payload.max_teams,
    registration_deadline: toIsoOrUndefined(payload.registration_deadline),
    start_at: toIsoOrUndefined(payload.start_at),
    visibility: payload.visibility,
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
};