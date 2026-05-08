import { Link, NavLink } from "react-router-dom";
import { Bell, Shield, Trophy, UserCircle } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { useUnreadNotifications } from "@/features/notifications/hooks";
import { isPlatformAdmin } from "@/shared/lib/roles";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { cn } from "@/shared/lib/cn";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "rounded-xl px-3 py-2 text-sm font-medium transition-colors",
    isActive ? "bg-slate-900 text-white" : "text-slate-700 hover:bg-slate-100",
  );

export function AppNavbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const { data: unread } = useUnreadNotifications(isAuthenticated);

  return (
    <header className="sticky top-0 z-30 border-b bg-white/90 backdrop-blur">
      <div className="page-shell flex items-center justify-between gap-4 py-4">
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2 text-sm font-semibold">
            <Trophy className="h-5 w-5" />
            <span>Esports MVP</span>
          </Link>
          <nav className="hidden items-center gap-1 md:flex">
            <NavLink to="/" className={navLinkClass}>
              Главная
            </NavLink>
            <NavLink to="/tournaments" className={navLinkClass}>
              Турниры
            </NavLink>
            {isAuthenticated ? (
              <>
                <NavLink to="/notifications" className={navLinkClass}>
                  <span className="inline-flex items-center gap-2">
                    Уведомления
                    {unread?.count ? <Badge>{unread.count}</Badge> : null}
                  </span>
                </NavLink>
                <NavLink to="/profile" className={navLinkClass}>
                  Профиль
                </NavLink>
                {isPlatformAdmin(user) ? (
                  <NavLink to="/platform-admin" className={navLinkClass}>
                    Платформа
                  </NavLink>
                ) : null}
              </>
            ) : null}
          </nav>
        </div>

        <div className="flex items-center gap-2">
          {isAuthenticated ? (
            <>
              <NavLink to="/notifications" className="relative rounded-xl border p-2 hover:bg-slate-50">
                <Bell className="h-4 w-4" />
                {unread?.count ? (
                  <span className="absolute -right-1 -top-1 flex h-5 min-w-5 items-center justify-center rounded-full bg-slate-900 px-1 text-[10px] text-white">
                    {unread.count}
                  </span>
                ) : null}
              </NavLink>
              {isPlatformAdmin(user) ? (
                <NavLink to="/platform-admin" className="rounded-xl border p-2 hover:bg-slate-50">
                  <Shield className="h-4 w-4" />
                </NavLink>
              ) : null}
              <NavLink to="/profile" className="rounded-xl border p-2 hover:bg-slate-50">
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