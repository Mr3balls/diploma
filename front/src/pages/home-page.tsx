import { Link } from "react-router-dom";
import { ArrowRight } from "lucide-react";
import { useTournaments } from "@/features/tournaments/hooks";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { useLang } from "@/app/providers/lang-provider";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { ParallaxCarousel } from "@/shared/ui/parallax-carousel";
import { useAuth } from "@/app/providers/auth-provider";

export function HomePage() {
  const { isAuthenticated } = useAuth();
  const { t } = useLang();
  const tournamentsQuery = useTournaments();
  const active = (tournamentsQuery.data?.items ?? []).filter(
    (t) => t.status === "registration_open" || t.status === "in_progress",
  ).slice(0, 6);
  const latest = active.length
    ? active
    : (tournamentsQuery.data?.items ?? []).slice(0, 6);

  const FEATURES = [
    { title: t("home.step1.title"), desc: t("home.step1.desc") },
    { title: t("home.step2.title"), desc: t("home.step2.desc") },
    { title: t("home.step3.title"), desc: t("home.step3.desc") },
  ];

  return (
    <div className="grid gap-0">

      {/* ── Hero ──────────────────────────────────────────────────── */}
      <section
        className="relative overflow-hidden"
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          background: "#111111",
          borderBottom: "1px solid #2d2d2d",
        }}
      >
        {/* orange glow */}
        <div
          aria-hidden
          style={{
            position: "absolute",
            top: "-120px",
            left: "50%",
            transform: "translateX(-50%)",
            width: "900px",
            height: "500px",
            background: "radial-gradient(ellipse at center, rgba(255,85,0,0.18) 0%, transparent 70%)",
            pointerEvents: "none",
          }}
        />

        <div className="mx-auto w-full max-w-7xl px-4 py-12 sm:py-24 sm:px-6 lg:px-8">
          <div className="grid gap-10 md:grid-cols-[1fr_auto] md:items-center">
            <div className="space-y-7 max-w-2xl">
              <p className="text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">
                {t("home.platform")}
              </p>
              <h1
                className="font-black leading-none text-white uppercase"
                style={{ fontSize: "clamp(2.8rem, 7vw, 5.5rem)", letterSpacing: "-0.03em" }}
              >
                {t("home.hero.titleLine1")}
                <br />
                <span className="text-[#ff5500]">{t("home.hero.titleLine2")}</span>
                <br />
                {t("home.hero.titleLine3")}
              </h1>
              <p className="text-[#9e9e9e] text-base max-w-md leading-relaxed">
                {t("home.hero.desc")}
              </p>
              <div className="flex flex-wrap gap-3">
                <Link to="/tournaments">
                  <Button size="lg" className="gap-2">
                    {t("home.hero.findTournament")} <ArrowRight className="h-4 w-4" />
                  </Button>
                </Link>
                {!isAuthenticated && (
                  <Link to="/register">
                    <Button size="lg" variant="outline">{t("home.hero.createAccount")}</Button>
                  </Link>
                )}
              </div>
            </div>

            {/* side accent block */}
            <div className="hidden md:grid gap-3 w-56">
              {["Single Elimination", "Double Elimination", "Group Stage", "Group + DE"].map((f) => (
                <div
                  key={f}
                  className="flex items-center gap-3 rounded-xl border border-[#2d2d2d] bg-[#1a1a1a] px-4 py-3 text-sm text-white"
                >
                  <span className="h-1.5 w-1.5 rounded-full bg-[#ff5500] shrink-0" />
                  {f}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ── Parallax carousel ─────────────────────────────────────── */}
      <ParallaxCarousel />

      {/* ── How it works ──────────────────────────────────────────── */}
      <section className="py-10 sm:py-20 border-t border-[#2d2d2d]">
        <div className="mb-8 sm:mb-12 text-center">
          <p className="mb-3 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">{t("home.how.label")}</p>
          <h2 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
            {t("home.how.title")}
          </h2>
        </div>
        <div className="grid gap-px md:grid-cols-3 rounded-2xl overflow-hidden border border-[#2d2d2d] bg-[#2d2d2d]">
          {FEATURES.map((f, i) => (
            <div key={f.title} className="bg-[#1a1a1a] p-5 sm:p-8 space-y-4">
              <div className="flex items-center gap-3">
                <span className="text-4xl font-black text-[#ff5500] leading-none">{String(i + 1).padStart(2, "0")}</span>
              </div>
              <h3 className="text-lg font-bold text-white uppercase tracking-wide">{f.title}</h3>
              <p className="text-sm text-[#9e9e9e] leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* ── Active tournaments ────────────────────────────────────── */}
      <section className="pb-12 sm:pb-20 space-y-6 border-t border-[#2d2d2d] pt-10 sm:pt-20">
        <div className="flex flex-wrap items-end justify-between gap-3">
          <div>
            <p className="mb-1 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">{t("home.active.label")}</p>
            <h2 className="text-2xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
              {t("home.active.title")}
            </h2>
          </div>
          <Link to="/tournaments">
            <Button variant="outline" size="sm" className="gap-1.5">
              {t("home.active.allTournaments")} <ArrowRight className="h-3.5 w-3.5" />
            </Button>
          </Link>
        </div>

        {tournamentsQuery.isLoading ? (
          <Spinner />
        ) : latest.length ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {latest.map((tournament) => (
              <TournamentCard key={tournament.id} tournament={tournament} />
            ))}
          </div>
        ) : (
          <EmptyState title={t("home.active.empty")} description={t("home.active.emptyDesc")} />
        )}
      </section>

    </div>
  );
}
