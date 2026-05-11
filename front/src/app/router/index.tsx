import { createBrowserRouter } from "react-router-dom";
import { AppShellLayout } from "@/layouts/app-shell-layout";
import { AuthLayout } from "@/layouts/auth-layout";
import { GuestRoute } from "@/shared/router/guest-route";
import { PlatformAdminRoute } from "@/shared/router/platform-admin-route";
import { ProtectedRoute } from "@/shared/router/protected-route";
import { HomePage } from "@/pages/home-page";
import { LoginPage } from "@/pages/login-page";
import { RegisterPage } from "@/pages/register-page";
import { ForgotPasswordPage } from "@/pages/forgot-password-page";
import { ResetPasswordPage } from "@/pages/reset-password-page";
import { ProfilePage } from "@/pages/profile-page";
import { NotificationsPage } from "@/pages/notifications-page";
import { TournamentsListPage } from "@/pages/tournaments-list-page";
import { TournamentDetailsPage } from "@/pages/tournament-details-page";
import { TournamentAdminPage } from "@/pages/tournament-admin-page";
import { PlatformAdminPage } from "@/pages/platform-admin-page";
import { NotFoundPage } from "@/pages/not-found-page";
import { ChallongePage } from "@/pages/challonge-page";
import { ChallongeTournamentPage } from "@/pages/challonge-tournament-page";

export const router = createBrowserRouter([
  {
    element: <AppShellLayout />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/tournaments", element: <TournamentsListPage /> },
      { path: "/tournaments/:id", element: <TournamentDetailsPage /> },
      { path: "/challonge", element: <ChallongePage /> },
      { path: "/challonge/:slug", element: <ChallongeTournamentPage /> },
      {
        path: "/profile",
        element: (
          <ProtectedRoute>
            <ProfilePage />
          </ProtectedRoute>
        ),
      },
      {
        path: "/notifications",
        element: (
          <ProtectedRoute>
            <NotificationsPage />
          </ProtectedRoute>
        ),
      },
      {
        path: "/tournaments/:id/admin",
        element: (
          <ProtectedRoute>
            <TournamentAdminPage />
          </ProtectedRoute>
        ),
      },
      {
        path: "/platform-admin",
        element: (
          <ProtectedRoute>
            <PlatformAdminRoute>
              <PlatformAdminPage />
            </PlatformAdminRoute>
          </ProtectedRoute>
        ),
      },
    ],
  },
  {
    element: (
      <GuestRoute>
        <AuthLayout />
      </GuestRoute>
    ),
    children: [
      { path: "/login", element: <LoginPage /> },
      { path: "/register", element: <RegisterPage /> },
      { path: "/forgot-password", element: <ForgotPasswordPage /> },
      { path: "/reset-password", element: <ResetPasswordPage /> },
    ],
  },
  { path: "*", element: <NotFoundPage /> },
]);