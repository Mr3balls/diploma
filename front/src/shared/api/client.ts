import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import type { ApiErrorResponse, TokenPair } from "@/shared/types/api";
import { tokenStorage } from "@/shared/api/token-storage";

const baseURL = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

const refreshClient = axios.create({ baseURL });
let refreshPromise: Promise<TokenPair> | null = null;

function redirectToLogin() {
  window.dispatchEvent(new Event("auth:logout"));
  if (!window.location.pathname.startsWith("/login")) {
    window.location.href = "/login";
  }
}

async function refreshTokens() {
  const refreshToken = tokenStorage.getRefreshToken();
  if (!refreshToken) {
    throw new Error("No refresh token");
  }

  const response = await refreshClient.post<{ tokens: TokenPair }>("/auth/refresh", {
    refresh_token: refreshToken,
  });

  tokenStorage.setTokens(response.data.tokens);
  return response.data.tokens;
}

function attachToken(config: InternalAxiosRequestConfig) {
  const accessToken = tokenStorage.getAccessToken();
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
}

export const apiClient = axios.create({
  baseURL,
});

apiClient.interceptors.request.use(attachToken);

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiErrorResponse>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (error.response?.status !== 401 || originalRequest?._retry) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      if (!refreshPromise) {
        refreshPromise = refreshTokens().finally(() => {
          refreshPromise = null;
        });
      }

      const tokens = await refreshPromise;
      originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`;
      return apiClient(originalRequest);
    } catch {
      tokenStorage.clear();
      redirectToLogin();
      return Promise.reject(error);
    }
  },
);