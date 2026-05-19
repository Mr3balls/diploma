import { apiClient } from "@/shared/api/client";
import type { GenericMessageResponse, ListResponse, MyTournamentEntry, User, UserStats } from "@/shared/types/api";
import type { ProfileFormValues } from "@/features/profile/schemas";

export const profileApi = {
  async getMe() {
    const { data } = await apiClient.get<User>("/me");
    return data;
  },
  async updateMe(payload: ProfileFormValues) {
    const { data } = await apiClient.patch<User>("/me", payload);
    return data;
  },
  async deleteMe() {
    const { data } = await apiClient.delete<GenericMessageResponse>("/me");
    return data;
  },
  async getMyStats() {
    const { data } = await apiClient.get<UserStats>("/me/stats");
    return data;
  },
  async getMyTournaments() {
    const { data } = await apiClient.get<ListResponse<MyTournamentEntry>>("/me/tournaments");
    return data;
  },
};