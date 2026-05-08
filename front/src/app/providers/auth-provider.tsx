import {
  createContext,
  PropsWithChildren,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { profileApi } from "@/features/profile/api";
import { authApi } from "@/features/auth/api";
import type { LoginRequest, RegisterRequest } from "@/features/auth/schemas";
import { tokenStorage } from "@/shared/api/token-storage";
import type { TokenPair, User } from "@/shared/types/api";

type AuthContextValue = {
  user: User | null;
  isAuthenticated: boolean;
  isBootstrapping: boolean;
  login: (payload: LoginRequest) => Promise<void>;
  register: (payload: RegisterRequest) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (user: User | null) => void;
  setTokens: (tokens: TokenPair | null) => void;
};

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: PropsWithChildren) {
  const [user, setUser] = useState<User | null>(null);
  const [isBootstrapping, setIsBootstrapping] = useState(true);

  const setTokens = useCallback((tokens: TokenPair | null) => {
    if (tokens) tokenStorage.setTokens(tokens);
    else tokenStorage.clear();
  }, []);

  const bootstrap = useCallback(async () => {
    const accessToken = tokenStorage.getAccessToken();
    if (!accessToken) {
      setIsBootstrapping(false);
      return;
    }

    try {
      const me = await profileApi.getMe();
      setUser(me);
    } catch {
      tokenStorage.clear();
      setUser(null);
    } finally {
      setIsBootstrapping(false);
    }
  }, []);

  useEffect(() => {
    void bootstrap();
  }, [bootstrap]);

  useEffect(() => {
    const handleForcedLogout = () => {
      setUser(null);
      tokenStorage.clear();
    };

    window.addEventListener("auth:logout", handleForcedLogout);
    return () => window.removeEventListener("auth:logout", handleForcedLogout);
  }, []);

  const login = useCallback(async (payload: LoginRequest) => {
    const response = await authApi.login(payload);
    tokenStorage.setTokens(response.tokens);
    setUser(response.user);
  }, []);

  const register = useCallback(async (payload: RegisterRequest) => {
    const response = await authApi.register(payload);
    tokenStorage.setTokens(response.tokens);
    setUser(response.user);
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } finally {
      tokenStorage.clear();
      setUser(null);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      isAuthenticated: Boolean(user && tokenStorage.getAccessToken()),
      isBootstrapping,
      login,
      register,
      logout,
      setUser,
      setTokens,
    }),
    [isBootstrapping, login, logout, register, setTokens, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (!value) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return value;
}