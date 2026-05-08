import { apiClient } from "@/shared/api/client";
import type { GenericMessageResponse, ListResponse, Tournament, User } from "@/shared/types/api";

export const adminApi = {
  async getUsers() {
    const { data } = await apiClient.get<ListResponse<User>>("/admin/users");
    return data;
  },
  async blockUser(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/admin/users/${id}/block`);
    return data;
  },
  async unblockUser(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/admin/users/${id}/unblock`);
    return data;
  },
  async getTournaments() {
    const { data } = await apiClient.get<ListResponse<Tournament>>("/admin/tournaments");
    return data;
  },
};