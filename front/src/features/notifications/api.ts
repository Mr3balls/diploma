import { apiClient } from "@/shared/api/client";
import type {
  GenericMessageResponse,
  ListResponse,
  Notification,
  UnreadCountResponse,
} from "@/shared/types/api";

export const notificationsApi = {
  async list() {
    const { data } = await apiClient.get<ListResponse<Notification>>("/notifications");
    return data;
  },
  async unreadCount() {
    const { data } = await apiClient.get<UnreadCountResponse>("/notifications/unread-count");
    return data;
  },
  async markRead(id: string) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/notifications/${id}/read`);
    return data;
  },
  async readAll() {
    const { data } = await apiClient.post<GenericMessageResponse>("/notifications/read-all");
    return data;
  },
  async action(id: string, payload?: Record<string, unknown>) {
    const { data } = await apiClient.post<GenericMessageResponse>(`/notifications/${id}/action`, payload ?? {});
    return data;
  },
};