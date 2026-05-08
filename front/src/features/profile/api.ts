import { apiClient } from "@/shared/api/client";
import type { GenericMessageResponse, User } from "@/shared/types/api";
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
};