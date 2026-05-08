import { apiClient } from "@/shared/api/client";
import type { AuthResponse, GenericMessageResponse } from "@/shared/types/api";
import type {
  ForgotPasswordRequest,
  LoginRequest,
  RegisterRequest,
  ResetPasswordRequest,
} from "@/features/auth/schemas";

export const authApi = {
  async register(payload: RegisterRequest) {
    const { data } = await apiClient.post<AuthResponse>("/auth/register", payload);
    return data;
  },
  async login(payload: LoginRequest) {
    const { data } = await apiClient.post<AuthResponse>("/auth/login", payload);
    return data;
  },
  async logout() {
    const { data } = await apiClient.post<GenericMessageResponse>("/auth/logout");
    return data;
  },
  async forgotPassword(payload: ForgotPasswordRequest) {
    const { data } = await apiClient.post<Record<string, unknown>>("/auth/forgot-password", payload);
    return data;
  },
  async resetPassword(payload: ResetPasswordRequest) {
    const { data } = await apiClient.post<Record<string, unknown>>("/auth/reset-password", payload);
    return data;
  },
};