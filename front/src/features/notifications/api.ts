import { apiClient } from "@/shared/api/client";
import type {
  GenericMessageResponse,
  ListResponse,
  Notification,
  NotificationPreferences,
  UnreadCountResponse,
} from "@/shared/types/api";

export const notificationsApi = {
  async list(limit = 20, offset = 0) {
    const { data } = await apiClient.get<ListResponse<Notification>>("/notifications", {
      params: { limit, offset },
    });
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
  async delete(id: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>(`/notifications/${id}`);
    return data;
  },
  async deleteAll() {
    const { data } = await apiClient.delete<GenericMessageResponse>("/notifications");
    return data;
  },
  async getPreferences() {
    const { data } = await apiClient.get<NotificationPreferences>("/notifications/preferences");
    return data;
  },
  async setPreferences(disabled: string[]) {
    const { data } = await apiClient.patch<GenericMessageResponse>("/notifications/preferences", { disabled });
    return data;
  },
  async getVAPIDPublicKey() {
    const { data } = await apiClient.get<{ public_key: string }>("/notifications/vapid-public-key");
    return data.public_key;
  },
  async registerPush(endpoint: string, p256dh: string, auth: string) {
    const { data } = await apiClient.post<GenericMessageResponse>("/notifications/push", { endpoint, p256dh, auth });
    return data;
  },
  async unregisterPush(endpoint: string) {
    const { data } = await apiClient.delete<GenericMessageResponse>("/notifications/push", { data: { endpoint } });
    return data;
  },
};
