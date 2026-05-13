import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminAccess } from "@/shared/hooks/use-tournament-admin-access";
import {
  useTournament,
  useTournamentParticipants,
  useJoinIndividualTournament,
  useRegisterTeam,
} from "@/features/tournaments/hooks";
import { useTournamentTeams, useTeam } from "@/features/teams/hooks";
import { useTournamentBracket } from "@/features/bracket/hooks";
import { useTournamentMatches } from "@/features/matches/hooks";
import { BracketView } from "@/features/bracket/components/bracket-view";
import { GroupStageView } from "@/features/bracket/components/group-stage-view";
import { GroupDEView } from "@/features/bracket/components/group-de-view";
import { MatchesTable } from "@/features/matches/components/matches-table";
import { ResultsView } from "@/features/matches/components/results-view";
import { TeamsTable } from "@/features/teams/components/teams-table";
import { TeamDetailsCard } from "@/features/teams/components/team-details-card";
import { PageHeader } from "@/shared/ui/page-header";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Input } from "@/shared/ui/input";
import { Spinner } from "@/shared/ui/spinner";
import { Tabs } from "@/shared/ui/tabs";
import { formatDateTime } from "@/shared/lib/date";
import { tournamentStatusLabel, visibilityLabel } from "@/shared/lib/enums";
import { getErrorMessage } from "@/shared/lib/http";

const STATUS_TONE: Record<string, "default" | "success" | "danger" | "muted" | "warning"> = {
  draft: "muted",
  registration_open: "default",
  registration_closed: "muted",
  bracket_generated: "default",
  in_progress: "warning",
  finished: "success",
  cancelled: "danger",
  ready: "default",
  completed: "success",
};

export function TournamentDetailsPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const [tab, setTab] = useState("bracket");
  const [showRegisterForm, setShowRegisterForm] = useState(false);
  const [detailsTeamId, setDetailsTeamId] = useState<string | null>(null);
  const [teamName, setTeamName] = useState("");
  const [members, setMembers] = useState(["", "", "", "", ""]);

  const tournamentQuery = useTournament(id);
  const access = useTournamentAdminAccess(id, tournamentQuery.data);

  const isIndividual = tournamentQuery.data?.registration_mode === "individual";

  const teamsQuery = useTournamentTeams(id);
  const bracketQuery = useTournamentBracket(id);
  const matchesQuery = useTournamentMatches(id);
  const participantsQuery = useTournamentParticipants(isIndividual ? id : undefined);
  const joinMutation = useJoinIndividualTournament(id);
  const registerTeamMutation = useRegisterTeam(id);
  const teamDetailsQuery = useTeam(detailsTeamId ?? undefined, Boolean(detailsTeamId));

  const isFinished = tournamentQuery.data?.status === "finished" || tournamentQuery.data?.status === "completed";

  const tabs = useMemo(() => {
    const items = [
      { value: "bracket", label: "Сетка" },
      { value: "teams", label: isIndividual ? "Участники" : "Команды" },
    ];
    if (isFinished) {
      items.push({ value: "results", label: "Результаты" });
    } else {
      items.push({ value: "matches", label: "Матчи" });
    }
    items.push({ value: "rules", label: "Правила" });
    if (access.canAccessAdmin) {
      items.push({ value: "admin", label: "Управление" });
    }
    return items;
  }, [access.canAccessAdmin, isIndividual, isFinished]);

  if (tournamentQuery.isLoading) return <Spinner />;
  if (tournamentQuery.isError || !tournamentQuery.data) return <ErrorState />;

  const tournament = tournamentQuery.data;
  const canJoinIndividual =
    user &&
    !access.canAccessAdmin &&
    isIndividual &&
    (tournament.status === "draft" || tournament.status === "registration_open");

  const canRegisterTeam =
    user &&
    !access.canAccessAdmin &&
    !isIndividual &&
    tournament.status === "registration_open";

  async function handleJoinIndividual() {
    try {
      await joinMutation.mutateAsync();
      toast.success("Вы записаны на турнир!");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleRegisterTeam() {
    if (!teamName.trim()) {
      toast.error("Введите название команды");
      return;
    }
    try {
      const filteredMembers = members.map((m) => m.trim()).filter(Boolean);
      await registerTeamMutation.mutateAsync({ team_name: teamName.trim(), members: filteredMembers });
      toast.success("Команда зарегистрирована! Ожидайте подтверждения.");
      setShowRegisterForm(false);
      setTeamName("");
      setMembers(["", "", "", ""]);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  function updateMember(index: number, value: string) {
    setMembers((prev) => prev.map((m, i) => (i === index ? value : m)));
  }

  function addMemberField() {
    setMembers((prev) => [...prev, ""]);
  }

  function removeMemberField(index: number) {
    setMembers((prev) => prev.filter((_, i) => i !== index));
  }

  const winnerName = tournament.winner_team_id
    ? (teamsQuery.data?.items.find((t) => t.id === tournament.winner_team_id)?.name ?? null)
    : tournament.winner_participant_id
    ? (participantsQuery.data?.items.find((p) => p.id === tournament.winner_participant_id)?.name ?? null)
    : null;

  return (
    <div className="grid gap-6">
      {winnerName && (
        <div className="flex items-center gap-3 rounded-xl border border-yellow-500/30 bg-yellow-500/10 px-5 py-4">
          <span className="text-2xl">🏆</span>
          <div>
            <p className="text-xs font-medium uppercase tracking-wide text-yellow-400">Победитель турнира</p>
            <p className="text-lg font-semibold text-white">{winnerName}</p>
          </div>
        </div>
      )}
      <PageHeader
        title={tournament.title}
        description={tournament.description || ""}
        actions={
          <div className="flex flex-wrap gap-2">
            <Badge tone={STATUS_TONE[tournament.status] ?? "muted"}>
              {tournamentStatusLabel[tournament.status]}
            </Badge>
            {tournament.visibility === "private" && (
              <Badge tone="muted">{visibilityLabel[tournament.visibility]}</Badge>
            )}
            {canJoinIndividual && (
              <Button size="sm" disabled={joinMutation.isPending} onClick={handleJoinIndividual}>
                {joinMutation.isPending ? "Запись..." : "Записаться"}
              </Button>
            )}
            {canRegisterTeam && (
              <Button
                size="sm"
                variant={showRegisterForm ? "outline" : "default"}
                onClick={() => setShowRegisterForm((v) => !v)}
              >
                {showRegisterForm ? "Отмена" : "Записаться с командой"}
              </Button>
            )}
            {access.canAccessAdmin && (
              <Link to={`/tournaments/${id}/admin`}>
                <Button variant="outline" size="sm">Управление</Button>
              </Link>
            )}
          </div>
        }
      />

      {showRegisterForm && canRegisterTeam && (
        <Card>
          <CardHeader>
            <CardTitle>Регистрация команды</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="grid gap-1">
              <label className="text-sm text-[#90afd4]">Название команды</label>
              <Input
                placeholder="Введите название команды"
                value={teamName}
                onChange={(e) => setTeamName(e.target.value)}
                className="md:max-w-sm"
              />
            </div>
            <div className="grid gap-2">
              <label className="text-sm text-[#90afd4]">Игроки (никнеймы)</label>
              <p className="text-xs text-[#90afd4]">Вы будете капитаном. Добавьте других участников по никнейму.</p>
              {members.map((member, index) => (
                <div key={index} className="flex gap-2 md:max-w-sm">
                  <Input
                    placeholder={`Игрок ${index + 2}`}
                    value={member}
                    onChange={(e) => updateMember(index, e.target.value)}
                  />
                  {members.length > 1 && (
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => removeMemberField(index)}
                      className="shrink-0"
                    >
                      ×
                    </Button>
                  )}
                </div>
              ))}
              {members.length < 9 && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={addMemberField}
                  className="md:max-w-sm"
                >
                  + Добавить игрока
                </Button>
              )}
            </div>
            <div className="flex gap-3">
              <Button
                disabled={registerTeamMutation.isPending}
                onClick={() => void handleRegisterTeam()}
              >
                {registerTeamMutation.isPending ? "Отправка..." : "Зарегистрировать команду"}
              </Button>
              <Button variant="outline" onClick={() => setShowRegisterForm(false)}>
                Отмена
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="flex flex-wrap gap-6 pt-5 text-sm text-[#90afd4]">
          {tournament.discipline && <div>Игра: <span className="text-white">{tournament.discipline}</span></div>}
          {tournament.max_teams && <div>Макс. участников: <span className="text-white">{tournament.max_teams}</span></div>}
          <div>Создан: <span className="text-white">{formatDateTime(tournament.created_at)}</span></div>
        </CardContent>
      </Card>

      <Tabs value={tab} onValueChange={setTab} tabs={tabs} />

      {tab === "bracket" ? (
        isIndividual ? (
          participantsQuery.isLoading || bracketQuery.isLoading ? (
            <Spinner />
          ) : (bracketQuery.data?.matches ?? []).length > 0 ? (
            <BracketView
              matches={bracketQuery.data?.matches ?? []}
              participants={participantsQuery.data?.items ?? []}
            />
          ) : (
            <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
          )
        ) : bracketQuery.isLoading ? (
          <Spinner />
        ) : bracketQuery.isError ? (
          <ErrorState />
        ) : bracketQuery.data?.bracket?.format === "group_stage" ? (
          <div className="space-y-6">
            {bracketQuery.data.bracket?.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
              <div className="space-y-2">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-[#90afd4]">Плей-офф</h3>
                <BracketView
                  matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)}
                  teams={teamsQuery.data?.items ?? []}
                  participants={participantsQuery.data?.items ?? []}
                />
              </div>
            )}
            {(bracketQuery.data.groups ?? []).length > 0 ? (
              <GroupStageView
                groups={bracketQuery.data.groups ?? []}
                matches={bracketQuery.data.matches ?? []}
                teams={teamsQuery.data?.items ?? []}
              />
            ) : (
              <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
            )}
          </div>
        ) : bracketQuery.data?.bracket?.format === "group_de" ? (
          <div className="space-y-6">
            {/* Playoff bracket — shown when all groups are done */}
            {bracketQuery.data.bracket?.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
              <div className="space-y-2">
                <h3 className="text-sm font-semibold uppercase tracking-wide text-[#90afd4]">Плей-офф</h3>
                <BracketView
                  matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)}
                  teams={teamsQuery.data?.items ?? []}
                  participants={participantsQuery.data?.items ?? []}
                />
              </div>
            )}
            {(bracketQuery.data.groups ?? []).length > 0 ? (
              <GroupDEView
                groups={bracketQuery.data.groups ?? []}
                matches={bracketQuery.data.matches ?? []}
                teams={teamsQuery.data?.items ?? []}
              />
            ) : (
              <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
            )}
          </div>
        ) : (bracketQuery.data?.matches ?? []).length > 0 ? (
          <BracketView matches={bracketQuery.data?.matches ?? []} teams={teamsQuery.data?.items ?? []} participants={participantsQuery.data?.items ?? []} />
        ) : (
          <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
        )
      ) : null}

      {tab === "teams" ? (
        isIndividual ? (
          participantsQuery.isLoading ? (
            <Spinner />
          ) : (participantsQuery.data?.items ?? []).length > 0 ? (
            <Card>
              <CardContent className="pt-5">
                <ul className="divide-y divide-[#0a3575]">
                  {[...(participantsQuery.data?.items ?? [])].sort((a, b) => a.seed - b.seed).map((p) => (
                    <li key={p.id} className="flex items-center gap-3 py-2">
                      <span className="w-6 text-right text-xs text-[#4a7ab5]">{p.seed}</span>
                      <span className="text-sm text-white">{p.name}</span>
                    </li>
                  ))}
                </ul>
              </CardContent>
            </Card>
          ) : (
            <EmptyState title="Участников пока нет" description="Никто ещё не записался на турнир." />
          )
        ) : teamsQuery.isLoading ? (
          <Spinner />
        ) : teamsQuery.isError ? (
          <ErrorState />
        ) : teamsQuery.data?.items.length ? (
          <>
            <TeamsTable
              teams={teamsQuery.data.items}
              onOpen={setDetailsTeamId}
            />
            {detailsTeamId && teamDetailsQuery.isLoading ? <Spinner /> : null}
            {teamDetailsQuery.data && detailsTeamId ? (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <p className="text-xs text-[#4a7ab5]">Состав команды</p>
                  <Button variant="ghost" size="sm" className="text-xs text-[#4a7ab5]" onClick={() => setDetailsTeamId(null)}>
                    Закрыть
                  </Button>
                </div>
                <TeamDetailsCard data={teamDetailsQuery.data} />
              </div>
            ) : null}
          </>
        ) : (
          <EmptyState title="Команд пока нет" description="Регистрация ещё не началась." />
        )
      ) : null}

      {tab === "matches" && !isFinished ? (
        matchesQuery.isLoading ? (
          <Spinner />
        ) : matchesQuery.isError ? (
          <ErrorState />
        ) : matchesQuery.data?.items.length ? (
          <MatchesTable matches={matchesQuery.data.items} teams={teamsQuery.data?.items ?? []} participants={participantsQuery.data?.items ?? []} />
        ) : (
          <EmptyState title="Матчей пока нет" description="Матчи появятся после генерации сетки." />
        )
      ) : null}

      {(tab === "results" || (tab === "matches" && isFinished)) ? (
        matchesQuery.isLoading ? (
          <Spinner />
        ) : (
          <ResultsView
            matches={matchesQuery.data?.items ?? []}
            teams={teamsQuery.data?.items ?? []}
            participants={participantsQuery.data?.items ?? []}
          />
        )
      ) : null}

      {tab === "rules" ? (
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Правила</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-[#90afd4]">
              {tournament.rules || "Не заполнено"}
            </CardContent>
          </Card>
          {tournament.location && (
            <Card>
              <CardHeader>
                <CardTitle>Место проведения</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-[#90afd4]">{tournament.location}</CardContent>
            </Card>
          )}
        </div>
      ) : null}

      {tab === "admin" && access.canAccessAdmin ? (
        <Card>
          <CardContent className="flex items-center justify-between gap-4 pt-5">
            <p className="text-sm text-[#90afd4]">
              Управление участниками, сеткой и матчами.
            </p>
            <Link to={`/tournaments/${id}/admin`}>
              <Button>Перейти в управление</Button>
            </Link>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
