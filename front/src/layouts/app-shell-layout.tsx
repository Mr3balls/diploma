import { Outlet } from "react-router-dom";
import { AppNavbar } from "@/widgets/app-navbar";

export function AppShellLayout() {
  return (
    <div className="min-h-screen bg-background">
      <AppNavbar />
      <main className="page-shell">
        <Outlet />
      </main>
    </div>
  );
}