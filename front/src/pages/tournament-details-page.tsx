import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminAccess } from "@/shared/hooks/use-tournament-admin-access";
import { useTournament, useTournamentAudit } from "@/features/tournaments/hooks";
import { useTournamentTeams } from "@/features/teams/hooks";
import { useTournamentBracket } from "@/features/bracket/hooks";
import { useTournamentMatches } from "@/features/matches/hooks";
import { BracketView } from "@/features/bracket/components/bracket-view";
import { MatchesTable } from "@/features/matches/components/matches-table";
import { TeamsTable } from "@/features/teams/components/teams-table";
import { PageHeader } from "@/shared/ui/page-header";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { Tabs } from "@/shared/ui/tabs";
import { formatDateTime } from "@/shared/lib/date";
import { tournamentStatusLabel, visibilityLabel } from "@/shared/lib/enums";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";

const publicTabs = [
  { value: "overview", label: "Обзор" },
  { value: "teams", label: "Команды" },
  { value: "bracket", label: "Сетка" },
  { value: "matches", label: "Матчи" },
  { value: "rules", label: "Правила / Инфо" },
];

export function TournamentDetailsPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const [tab, setTab] = useState("overview");

  const tournamentQuery = useTournament(id);
  const teamsQuery = useTournamentTeams(id);
  const bracketQuery = useTournamentBracket(id);
  const matchesQuery = useTournamentMatches(id);

  const access = useTournamentAdminAccess(id, tournamentQuery.data);
  const auditQuery = useTournamentAudit(id, access.canAccessAdmin);

  const tabs = useMemo(() => {
    const items = [...publicTabs];
    if (access.canAccessAdmin) {
      items.push({ value: "audit", label: "Audit" });
      items.push({ value: "admin", label: "Admin" });
    }
    return items;
  }, [access.canAccessAdmin]);

  if (tournamentQuery.isLoading) return <Spinner />;
  if (tournamentQuery.isError || !tournamentQuery.data) return <ErrorState />;

  const tournament = tournamentQuery.data;
  const isOwner = Boolean(user?.id && tournament.owner_user_id && user.id === tournament.owner_user_id);

  return (
    <div className="grid gap-6">
      <PageHeader
        title={tournament.title}
        description={tournament.description || "Описание не заполнено"}
        actions={
          <div className="flex flex-wrap gap-2">
            <Badge>{tournamentStatusLabel[tournament.status]}</Badge>
            <Badge tone="muted">{visibilityLabel[tournament.visibility]}</Badge>
            {access.canAccessAdmin ? (
              <Link to={`/tournaments/${id}/admin`}>
                <Button variant="outline">Открыть admin</Button>
              </Link>
            ) : null}
          </div>
        }
      />

      <Card>
        <CardContent className="flex flex-wrap gap-4 pt-5 text-sm text-slate-600">
          <div>Игра: {tournament.game || "—"}</div>
          <div>Макс. команд: {tournament.max_teams ?? "—"}</div>
          <div>Создан: {formatDateTime(tournament.created_at)}</div>
          <div>Owner user_id: {tournament.owner_user_id || "—"}</div>
          {isOwner ? <Badge tone="success">Вы владелец турнира</Badge> : null}
        </CardContent>
      </Card>

      <Tabs value={tab} onValueChange={setTab} tabs={tabs} />

      {tab === "overview" ? (
        <div className="section-grid">
          <Card>
            <CardHeader>
              <CardTitle>Кратко</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-slate-600">
              <div>Статус: {tournamentStatusLabel[tournament.status]}</div>
              <div>Видимость: {visibilityLabel[tournament.visibility]}</div>
              <div>Google Sheets импорт, сетка и матчи доступны из отдельной admin-страницы.</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Бизнес-правила</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-slate-600">
              <div>Команда готова к проверке после подтверждения капитана и минимум 4 основных игроков.</div>
              <div>Менеджер подтверждает финальный результат матча.</div>
              <div>Публичная сетка — только для просмотра.</div>
            </CardContent>
          </Card>
        </div>
      ) : null}

      {tab === "teams" ? (
        teamsQuery.isLoading ? (
          <Spinner />
        ) : teamsQuery.isError ? (
          <ErrorState />
        ) : teamsQuery.data?.items.length ? (
          <TeamsTable teams={teamsQuery.data.items} />
        ) : (
          <EmptyState title="Команд нет" description="После импорта или добавления команд список появится здесь." />
        )
      ) : null}

      {tab === "bracket" ? (
        bracketQuery.isLoading ? (
          <Spinner />
        ) : bracketQuery.isError ? (
          <ErrorState />
        ) : (
          <BracketView matches={bracketQuery.data?.matches ?? []} />
        )
      ) : null}

      {tab === "matches" ? (
        matchesQuery.isLoading ? (
          <Spinner />
        ) : matchesQuery.isError ? (
          <ErrorState />
        ) : matchesQuery.data?.items.length ? (
          <MatchesTable matches={matchesQuery.data.items} />
        ) : (
          <EmptyState title="Матчей пока нет" description="Матчи появятся после генерации сетки." />
        )
      ) : null}

      {tab === "rules" ? (
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Правила</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-slate-600">{tournament.rules || "Не заполнено"}</CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Доп. информация</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-slate-600">{tournament.info || "Не заполнено"}</CardContent>
          </Card>
        </div>
      ) : null}

      {tab === "audit" ? (
        !access.canAccessAdmin ? (
          <EmptyState title="Нет доступа" description="Audit доступен только owner/manager/platform admin." />
        ) : auditQuery.isLoading ? (
          <Spinner />
        ) : auditQuery.isError ? (
          <ErrorState />
        ) : auditQuery.data?.items.length ? (
          <Card>
            <CardContent className="overflow-x-auto pt-5">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Время</TableHead>
                    <TableHead>Действие</TableHead>
                    <TableHead>Actor</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {auditQuery.data.items.map((item) => (
                    <TableRow key={item.id}>
                      <TableCell>{formatDateTime(item.created_at)}</TableCell>
                      <TableCell>{item.action}</TableCell>
                      <TableCell>{item.actor_email || item.actor_user_id || "—"}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        ) : (
          <EmptyState title="Audit пуст" description="Записи аудита пока отсутствуют." />
        )
      ) : null}

      {tab === "admin" ? (
        access.canAccessAdmin ? (
          <Card>
            <CardContent className="flex items-center justify-between gap-4 pt-5">
              <div className="text-sm text-slate-600">
                Для управления турниром используйте отдельную admin-страницу.
              </div>
              <Link to={`/tournaments/${id}/admin`}>
                <Button>Перейти в admin</Button>
              </Link>
            </CardContent>
          </Card>
        ) : (
          <EmptyState title="Нет доступа" description="Admin-вкладка скрывается, если backend не подтверждает доступ." />
        )
      ) : null}
    </div>
  );
}