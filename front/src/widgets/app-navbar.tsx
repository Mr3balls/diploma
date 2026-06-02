import { useEffect, useState } from "react";
import { Link, NavLink, useLocation } from "react-router-dom";
import { Bell, LayoutList, Shield, Trophy, UserCircle, Menu, X } from "lucide-react";
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

const drawerLinkClass = ({ isActive }: { isActive: boolean }) =>
  cn(
    "flex items-center gap-3 rounded-xl px-4 py-3 text-base font-medium transition-colors",
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
  const location = useLocation();
  const [menuOpen, setMenuOpen] = useState(false);
  useNotificationStream(isAuthenticated);

  // Close drawer on route change
  useEffect(() => { setMenuOpen(false); }, [location.pathname]);

  function cycleLang() {
    const idx = LANGS.findIndex((l) => l.code === lang);
    const next = LANGS[(idx + 1) % LANGS.length];
    setLang(next.code);
  }

  function closeMenu() { setMenuOpen(false); }

  return (
    <>
      <header className="sticky top-0 z-30 border-b border-[#2d2d2d] bg-[#111111]/90 backdrop-blur">
        <div className="page-shell flex items-center justify-between gap-4 py-4">
          {/* Logo + desktop nav */}
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2 text-sm font-semibold" onClick={closeMenu}>
              <Trophy className="h-5 w-5 text-[#ff5500]" />
              <span className="text-white tracking-wide">ACE</span>
            </Link>
            <nav className="hidden items-center gap-1 md:flex">
              <NavLink to="/" end className={navLinkClass}>{t("nav.home")}</NavLink>
              <NavLink to="/tournaments" className={navLinkClass}>{t("nav.tournaments")}</NavLink>
            </nav>
          </div>

          {/* Desktop right side */}
          <div className="hidden md:flex items-center gap-2">
            <button
              onClick={cycleLang}
              className="rounded-xl border border-[#2d2d2d] px-3 py-2 text-xs font-bold text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white transition-colors min-w-[3rem] text-center"
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
                {isPlatformAdmin(user) && (
                  <NavLink to="/platform-admin" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white">
                    <Shield className="h-4 w-4" />
                  </NavLink>
                )}
                <NavLink to="/profile" className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white" title={t("nav.profile")}>
                  <UserCircle className="h-4 w-4" />
                </NavLink>
                <Button variant="outline" onClick={() => void logout()}>{t("nav.signOut")}</Button>
              </>
            ) : (
              <>
                <Link to="/login"><Button variant="outline">{t("nav.signIn")}</Button></Link>
                <Link to="/register"><Button>{t("nav.register")}</Button></Link>
              </>
            )}
          </div>

          {/* Mobile right: bell (if auth) + hamburger */}
          <div className="flex items-center gap-2 md:hidden">
            {isAuthenticated && (
              <NavLink to="/notifications" className="relative rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e]">
                <Bell className="h-4 w-4" />
                {unread?.count ? (
                  <span className="absolute -right-1 -top-1 flex h-4 min-w-4 items-center justify-center rounded-full bg-[#ff5500] px-1 text-[9px] text-white font-bold">
                    {unread.count}
                  </span>
                ) : null}
              </NavLink>
            )}
            <button
              onClick={() => setMenuOpen((v) => !v)}
              className="rounded-xl border border-[#2d2d2d] p-2 text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white transition-colors"
              aria-label="Menu"
            >
              {menuOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
          </div>
        </div>
      </header>

      {/* Mobile drawer overlay */}
      {menuOpen && (
        <div
          className="fixed inset-0 z-20 bg-black/60 md:hidden"
          onClick={closeMenu}
        />
      )}

      {/* Mobile drawer panel */}
      <div
        className={cn(
          "fixed left-0 right-0 z-20 border-b border-[#2d2d2d] bg-[#111111] md:hidden transition-all duration-200 overflow-y-auto overscroll-contain",
          menuOpen ? "translate-y-0 opacity-100 pointer-events-auto" : "-translate-y-2 opacity-0 pointer-events-none",
        )}
        style={{ top: "var(--navbar-h)", maxHeight: "calc(100dvh - var(--navbar-h))" }}
      >
        <div className="page-shell py-4 grid gap-1">
          {/* Nav links */}
          <NavLink to="/" end className={drawerLinkClass} onClick={closeMenu}>
            {t("nav.home")}
          </NavLink>
          <NavLink to="/tournaments" className={drawerLinkClass} onClick={closeMenu}>
            {t("nav.tournaments")}
          </NavLink>

          {isAuthenticated ? (
            <>
              <NavLink to="/my-tournaments" className={drawerLinkClass} onClick={closeMenu}>
                <LayoutList className="h-4 w-4" />
                {t("nav.myTournaments")}
              </NavLink>
              <NavLink to="/profile" className={drawerLinkClass} onClick={closeMenu}>
                <UserCircle className="h-4 w-4" />
                {t("nav.profile")}
              </NavLink>
              {isPlatformAdmin(user) && (
                <NavLink to="/platform-admin" className={drawerLinkClass} onClick={closeMenu}>
                  <Shield className="h-4 w-4" />
                  Admin
                </NavLink>
              )}
              <div className="mt-2 border-t border-[#2d2d2d] pt-3 flex items-center justify-between gap-3">
                <button
                  onClick={cycleLang}
                  className="rounded-xl border border-[#2d2d2d] px-4 py-2.5 text-sm font-bold text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white transition-colors"
                >
                  {LANGS.find((l) => l.code === lang)?.label}
                </button>
                <Button variant="outline" className="flex-1" onClick={() => { void logout(); closeMenu(); }}>
                  {t("nav.signOut")}
                </Button>
              </div>
            </>
          ) : (
            <div className="mt-2 border-t border-[#2d2d2d] pt-3 grid gap-2">
              <div className="flex items-center gap-2">
                <button
                  onClick={cycleLang}
                  className="rounded-xl border border-[#2d2d2d] px-4 py-2.5 text-sm font-bold text-[#9e9e9e] hover:bg-[#2a2a2a] hover:text-white transition-colors"
                >
                  {LANGS.find((l) => l.code === lang)?.label}
                </button>
                <Link to="/login" className="flex-1" onClick={closeMenu}>
                  <Button variant="outline" className="w-full">{t("nav.signIn")}</Button>
                </Link>
                <Link to="/register" className="flex-1" onClick={closeMenu}>
                  <Button className="w-full">{t("nav.register")}</Button>
                </Link>
              </div>
            </div>
          )}
        </div>
      </div>
    </>
  );
}
