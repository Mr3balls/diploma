import { Link } from "react-router-dom";
import { useTournaments } from "@/features/tournaments/hooks";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { PageHeader } from "@/shared/ui/page-header";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { Card, CardContent } from "@/shared/ui/card";

export function HomePage() {
  const tournamentsQuery = useTournaments();
  const latest = tournamentsQuery.data?.items.slice(0, 6) ?? [];

  return (
    <div className="grid gap-8">
      <section className="grid gap-6 rounded-3xl bg-slate-900 px-6 py-10 text-white md:grid-cols-[1.2fr_0.8fr]">
        <div className="space-y-4">
          <p className="text-sm uppercase tracking-[0.28em] text-slate-300">Esports Diploma MVP</p>
          <h1 className="text-4xl font-semibold leading-tight text-white">
            Управление турнирами,
            <br />
            импортом команд и сеткой
            <br />
            без лишних страниц
          </h1>
          <p className="max-w-2xl text-sm text-slate-300">
            Авторизация, список турниров, публичная страница турнира, отдельная admin-панель, импорт Google Sheets,
            подтверждения участия, уведомления и управление матчами.
          </p>
          <div className="flex flex-wrap gap-3">
            <Button onClick={() => (window.location.href = "/tournaments")}>Перейти к турнирам</Button>
            <Button variant="secondary" onClick={() => (window.location.href = "/register")}>
              Создать аккаунт
            </Button>
          </div>
        </div>

        <Card className="border-slate-700 bg-slate-800">
          <CardContent className="grid gap-4 pt-5 text-sm text-slate-200">
            <div>• Только реальные backend endpoints</div>
            <div>• Single elimination</div>
            <div>• Раздельные public/admin режимы сетки</div>
            <div>• Inline-действия в уведомлениях там, где это поддержано</div>
            <div>• Менеджеры добавляются только по user_id</div>
          </CardContent>
        </Card>
      </section>

      <section className="grid gap-4">
        <PageHeader
          title="Последние турниры"
          description="Публичный список турниров. Доступен без авторизации."
          actions={
            <Link to="/tournaments">
              <Button variant="outline">Все турниры</Button>
            </Link>
          }
        />

        {tournamentsQuery.isLoading ? <Spinner /> : null}
        {tournamentsQuery.isError ? <ErrorState /> : null}
        {!tournamentsQuery.isLoading && !tournamentsQuery.isError && !latest.length ? (
          <EmptyState title="Турниров пока нет" description="После создания турниры будут отображаться здесь." />
        ) : null}

        {latest.length ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {latest.map((tournament) => (
              <TournamentCard key={tournament.id} tournament={tournament} />
            ))}
          </div>
        ) : null}
      </section>
    </div>
  );
}