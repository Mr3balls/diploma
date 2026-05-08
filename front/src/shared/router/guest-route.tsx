import * as React from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/app/providers/auth-provider";
import { Spinner } from "@/shared/ui/spinner";

export function GuestRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isBootstrapping } = useAuth();

  if (isBootstrapping) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (isAuthenticated) {
    return <Navigate to="/tournaments" replace />;
  }

  return <>{children}</>;
}