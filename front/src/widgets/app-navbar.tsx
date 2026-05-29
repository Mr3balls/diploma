import { Link, NavLink } from "react-router-dom";
import { Bell, LayoutList, Shield, Trophy, UserCircle } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { useLang } from "@/app/providers/lang-provider";
import { useNotificationStream, useUnreadNotifications } from "@/features/notifications/hooks";
import { isPlatformAdmin } from "@/shared/lib/roles";
import { Button } from "@/shared/ui/button";
import { cn } from "@/shared/lib/cn";
import type { Lang } from "@/shared/lib/i18n";

const navLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "rounded-xl px-3 py-2 text-sm font-medium transition-colors",
    isActive ? "bg-[#ff5500] text-white" : "text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white",
  );

const LANGS: { code: Lang; label: string }[] = [
  { code: "ru", label: "RU" },
  { code: "en", label: "EN" },
  { code: "kk", label: "ҚАЗ" },
];

export function AppNavbar() {
  const { user, isAuthenticated, logout } = useAuth();
  const { lang, setLang, t } = useLang();
  const { data: unread } = useUnreadNotifications(isAuthenticated);
  useNotificationStream(isAuthenticated);

  function cycleLang() {
    const idx = LANGS.findIndex((l) => l.code === lang);
    const next = LANGS[(idx + 1) % LANGS.length];
    setLang(next.code);
  }

  return (
    <header className="sticky top-0 z-30 border-b border-[#2d2d2d] bg-[#111111]/90 backdrop-blur">
      <div className="page-shell flex items-center justify-between gap-4 py-4">
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2 text-sm font-semibold">
            <Trophy className="h-5 w-5 text-[#ff5500]" />
            <span className="text-white tracking-wide">ACE</span>
          </Link>
          <nav className="hidden items-center gap-1 md:flex">
            <NavLink to="/" end className={navLinkClass}>
              {t("nav.home")}
            </NavLink>
            <NavLink to="/tournaments" className={navLinkClass}>
              {t("nav.tournaments")}
            </NavLink>
          </nav>
        </div>

        <div className="flex items-center gap-2">
          {/* Language switcher */}
          <button
            onClick={cycleLang}
            className="rounded-xl border border-[#2d2d2d] px-3 py-2 text-xs font-bold text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white transition-colors min-w-[3rem] text-center"
            title="Switch language"
          >
            {LANGS.find((l) => l.code === lang)?.label}
          </button>

          {isAuthenticated ? (
            <>
              <NavLink to="/my-tournaments" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white" title={t("nav.myTournaments")}>
                <LayoutList className="h-4 w-4" />
              </NavLink>
              <NavLink to="/notifications" className="relative rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white" title={t("nav.notifications")}>
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
              <NavLink to="/profile" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white" title={t("nav.profile")}>
                <UserCircle className="h-4 w-4" />
              </NavLink>
              <Button variant="outline" onClick={() => void logout()}>
                {t("nav.signOut")}
              </Button>
            </>
          ) : (
            <>
              <Link to="/login"><Button variant="outline">{t("nav.signIn")}</Button></Link>
              <Link to="/register"><Button>{t("nav.register")}</Button></Link>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
