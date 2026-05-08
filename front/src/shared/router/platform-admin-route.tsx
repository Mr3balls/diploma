import * as React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/app/providers/auth-provider";
import { isPlatformAdmin } from "@/shared/lib/roles";

export function PlatformAdminRoute({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();

  if (!isPlatformAdmin(user)) {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
}