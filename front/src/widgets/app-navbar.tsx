import { Link, NavLink } from "react-router-dom";
import { Bell, LayoutList, Shield, Trophy, UserCircle } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { useNotificationStream, useUnreadNotifications } from "@/features/notifications/hooks";
import { isPlatformAdmin } from "@/shared/lib/roles";
import { Button } from "@/shared/ui/button";
import { cn } from "@/shared/lib/cn";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "rounded-xl px-3 py-2 text-sm font-medium transition-colors",
    isActive ? "bg-[#ff5500] text-white" : "text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white",
  );

export function AppNavbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const { data: unread } = useUnreadNotifications(isAuthenticated);
  useNotificationStream(isAuthenticated);

  return (
    <header className="sticky top-0 z-30 border-b border-[#2d2d2d] bg-[#111111]/90 backdrop-blur">
      <div className="page-shell flex items-center justify-between gap-4 py-4">
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2 text-sm font-semibold">
            <Trophy className="h-5 w-5 text-[#ff5500]" />
            <span className="text-white tracking-wide">ACE</span>
          </Link>
          <nav className="hidden items-center gap-1 md:flex">
            <NavLink to="/" className={navLinkClass}>
              Главная
            </NavLink>
            <NavLink to="/tournaments" className={navLinkClass}>
              Турниры
            </NavLink>
          </nav>
        </div>

        <div className="flex items-center gap-2">
          {isAuthenticated ? (
            <>
              <NavLink to="/my-tournaments" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white" title="Мои турниры">
                <LayoutList className="h-4 w-4" />
              </NavLink>
              <NavLink to="/notifications" className="relative rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white">
                <Bell className="h-4 w-4" />
                {unread?.count ? (
                  <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-[#ff5500] px-1 text-[10px] text-white font-bold">
                    {unread.count}
                  </span>
                ) : null}
              </NavLink>
              {isPlatformAdmin(user) ? (
                <NavLink to="/platform-admin" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white">
                  <Shield className="h-4 w-4" />
                </NavLink>
              ) : null}
              <NavLink to="/profile" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white">
                <UserCircle className="h-4 w-4" />
              </NavLink>
              <Button variant="outline" onClick={() => void logout()}>
                Выйти
              </Button>
            </>
          ) : (
            <>
              <Link to="/login"><Button variant="outline">Войти</Button></Link>
              <Link to="/register"><Button>Регистрация</Button></Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}