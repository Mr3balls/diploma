import * as React from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "@/app/providers/auth-provider";
import { Spinner } from "@/shared/ui/spinner";

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isBootstrapping } = useAuth();
  const location = useLocation();

  if (isBootstrapping) {
    return (
      <div className="py-10">
        <Spinner />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }

  return <>{children}</>;
}