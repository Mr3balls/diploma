import { Link } from "react-router-dom";
import { ArrowRight } from "lucide-react";
import { useTournaments } from "@/features/tournaments/hooks";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { ParallaxCarousel } from "@/shared/ui/parallax-carousel";
import { useAuth } from "@/app/providers/auth-provider";

const FEATURES = [
  {
    title: "Регистрируйся",
    desc: "Создай аккаунт за 30 секунд и записывайся на открытые турниры — соло или с командой.",
  },
  {
    title: "Играй",
    desc: "Получай расписание матчей, уведомления о результатах и следи за сеткой в реальном времени.",
  },
  {
    title: "Побеждай",
    desc: "Капитан управляет составом, администратор фиксирует результаты, победитель попадает в историю.",
  },
];

export function HomePage() {
  const { isAuthenticated } = useAuth();
  const tournamentsQuery = useTournaments();
  const active = (tournamentsQuery.data?.items ?? []).filter(
    (t) => t.status === "registration_open" || t.status === "in_progress",
  ).slice(0, 6);
  const latest = active.length
    ? active
    : (tournamentsQuery.data?.items ?? []).slice(0, 6);

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

        <div className="mx-auto w-full max-w-7xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="grid gap-10 md:grid-cols-[1fr_auto] md:items-center">
            <div className="space-y-7 max-w-2xl">
              <p className="text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">
                Esports Platform
              </p>
              <h1
                className="font-black leading-none text-white uppercase"
                style={{ fontSize: "clamp(2.8rem, 7vw, 5.5rem)", letterSpacing: "-0.03em" }}
              >
                Твой путь
                <br />
                <span className="text-[#ff5500]">к победе</span>
                <br />
                начинается здесь
              </h1>
              <p className="text-[#9e9e9e] text-base max-w-md leading-relaxed">
                Регистрируй команду, участвуй в турнирах, следи за сеткой и получай уведомления о каждом матче.
              </p>
              <div className="flex flex-wrap gap-3">
                <Link to="/tournaments">
                  <Button size="lg" className="gap-2">
                    Найти турнир <ArrowRight className="h-4 w-4" />
                  </Button>
                </Link>
                {!isAuthenticated && (
                  <Link to="/register">
                    <Button size="lg" variant="outline">Создать аккаунт</Button>
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
      <section className="py-20">
        <div className="mb-12 text-center">
          <p className="mb-3 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">Как это работает</p>
          <h2 className="text-3xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
            Три шага до победы
          </h2>
        </div>
        <div className="grid gap-px md:grid-cols-3 rounded-2xl overflow-hidden border border-[#2d2d2d]">
          {FEATURES.map((f, i) => (
            <div key={f.title} className="bg-[#1a1a1a] p-8 space-y-4">
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
      <section className="pb-20 space-y-6">
        <div className="flex items-end justify-between">
          <div>
            <p className="mb-1 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">Прямо сейчас</p>
            <h2 className="text-2xl font-black uppercase text-white" style={{ letterSpacing: "-0.02em" }}>
              Актуальные турниры
            </h2>
          </div>
          <Link to="/tournaments">
            <Button variant="outline" size="sm" className="gap-1.5">
              Все турниры <ArrowRight className="h-3.5 w-3.5" />
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
          <EmptyState title="Турниров пока нет" description="После создания турниры будут отображаться здесь." />
        )}
      </section>

    </div>
  );
}
