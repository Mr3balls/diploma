import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { toast } from "sonner";
import { Calendar, MapPin, Shield, Swords, Users, Settings } from "lucide-react";
import { useAuth } from "@/app/providers/auth-provider";
import { useTournamentAdminAccess } from "@/shared/hooks/use-tournament-admin-access";
import {
  useTournament,
  useTournamentParticipants,
  useJoinIndividualTournament,
  useRegisterTeam,
} from "@/features/tournaments/hooks";
import { useTournamentTeams, useTeam, useMyTeam, useReplaceMember } from "@/features/teams/hooks";
import { useTournamentBracket, useTournamentPlacements } from "@/features/bracket/hooks";
import { useTournamentMatches } from "@/features/matches/hooks";
import { BracketView } from "@/features/bracket/components/bracket-view";
import { PlacementsView } from "@/features/bracket/components/placements-view";
import { GroupStageView } from "@/features/bracket/components/group-stage-view";
import { GroupDEView } from "@/features/bracket/components/group-de-view";
import { MatchesTable } from "@/features/matches/components/matches-table";
import { ResultsView } from "@/features/matches/components/results-view";
import { TeamsTable } from "@/features/teams/components/teams-table";
import { TeamDetailsCard } from "@/features/teams/components/team-details-card";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Input } from "@/shared/ui/input";
import { Spinner } from "@/shared/ui/spinner";
import { Tabs } from "@/shared/ui/tabs";
import { formatDateTime, formatDate } from "@/shared/lib/date";
import { tournamentFormatLabel, tournamentStatusLabel } from "@/shared/lib/enums";
import { getErrorMessage } from "@/shared/lib/http";
import { cn } from "@/shared/lib/cn";
import type { Tournament } from "@/shared/types/api";

// ── Status helpers ────────────────────────────────────────────────────────────
function statusColor(status: Tournament["status"]) {
  switch (status) {
    case "registration_open": return "#ff5500";
    case "in_progress":       return "#f59e0b";
    case "finished":
    case "completed":         return "#22c55e";
    case "cancelled":         return "#ef4444";
    default:                  return "#666666";
  }
}

// ── Team captain helpers ──────────────────────────────────────────────────────
const TEAM_STATUS_LABEL: Record<string, string> = {
  pending: "На рассмотрении",
  awaiting_confirmation: "Ожидает подтверждений",
  ready_for_review: "Готова к проверке",
  approved: "Одобрена",
  rejected: "Отклонена",
};
const TEAM_STATUS_TONE: Record<string, "default" | "success" | "danger" | "muted" | "warning"> = {
  pending: "muted",
  awaiting_confirmation: "warning",
  ready_for_review: "default",
  approved: "success",
  rejected: "danger",
};
function TeamStatusBadge({ status }: { status: string }) {
  return <Badge tone={TEAM_STATUS_TONE[status] ?? "muted"}>{TEAM_STATUS_LABEL[status] ?? status}</Badge>;
}

const MEMBER_STATUS_TONE: Record<string, "default" | "success" | "danger" | "muted" | "warning"> = {
  confirmed: "success",
  pending_confirmation: "warning",
  declined: "danger",
  removed: "muted",
};
const MEMBER_STATUS_LABEL: Record<string, string> = {
  confirmed: "Подтверждён",
  pending_confirmation: "Ожидает",
  declined: "Отказался",
  removed: "Удалён",
};
function MemberStatusBadge({ status }: { status: string }) {
  return <Badge tone={MEMBER_STATUS_TONE[status] ?? "muted"}>{MEMBER_STATUS_LABEL[status] ?? status}</Badge>;
}

// ── Page ─────────────────────────────────────────────────────────────────────
export function TournamentDetailsPage() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const [tab, setTab] = useState("bracket");
  const [showRegisterForm, setShowRegisterForm] = useState(false);
  const [detailsTeamId, setDetailsTeamId] = useState<string | null>(null);
  const [teamName, setTeamName] = useState("");
  const [members, setMembers] = useState(["", "", "", "", ""]);
  const [replaceTargetId, setReplaceTargetId] = useState<string | null>(null);
  const [replaceEmail, setReplaceEmail] = useState("");

  const tournamentQuery = useTournament(id);
  const access = useTournamentAdminAccess(id, tournamentQuery.data);
  const isIndividual = tournamentQuery.data?.registration_mode === "individual";

  const teamsQuery = useTournamentTeams(id);
  const bracketQuery = useTournamentBracket(id);
  const placementsQuery = useTournamentPlacements(id);
  const matchesQuery = useTournamentMatches(id);
  const participantsQuery = useTournamentParticipants(isIndividual ? id : undefined);
  const joinMutation = useJoinIndividualTournament(id);
  const registerTeamMutation = useRegisterTeam(id);
  const teamDetailsQuery = useTeam(detailsTeamId ?? undefined, Boolean(detailsTeamId));
  const myTeamQuery = useMyTeam(
    !isIndividual && !access.canAccessAdmin && Boolean(user) ? id : undefined,
    user?.id,
  );
  const replaceMemberMutation = useReplaceMember(id);

  const isFinished = tournamentQuery.data?.status === "finished" || tournamentQuery.data?.status === "completed";

  const hasPlacements =
    tournamentQuery.data?.status === "bracket_generated" ||
    tournamentQuery.data?.status === "in_progress" ||
    tournamentQuery.data?.status === "finished" ||
    tournamentQuery.data?.status === "completed";

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
    if (hasPlacements) {
      items.push({ value: "places", label: "Места" });
    }
    items.push({ value: "rules", label: "Правила" });
    return items;
  }, [isIndividual, isFinished, hasPlacements]);

  if (tournamentQuery.isLoading) return <div className="flex items-center justify-center py-32"><Spinner /></div>;
  if (tournamentQuery.isError || !tournamentQuery.data) return <ErrorState />;

  const tournament = tournamentQuery.data;
  const color = statusColor(tournament.status);

  const canJoinIndividual =
    user && !access.canAccessAdmin && isIndividual &&
    (tournament.status === "draft" || tournament.status === "registration_open");

  const canRegisterTeam =
    user && !access.canAccessAdmin && !isIndividual &&
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
    if (!teamName.trim()) { toast.error("Введите название команды"); return; }
    try {
      const filteredMembers = members.map((m) => m.trim()).filter(Boolean);
      await registerTeamMutation.mutateAsync({ team_name: teamName.trim(), members: filteredMembers });
      toast.success("Команда зарегистрирована! Ожидайте подтверждения.");
      setShowRegisterForm(false);
      setTeamName("");
      setMembers(["", "", "", "", ""]);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleReplaceMember(teamId: string, memberId: string) {
    if (!replaceEmail.trim()) return;
    try {
      await replaceMemberMutation.mutateAsync({ teamId, memberId, email: replaceEmail.trim() });
      toast.success("Игрок заменён. Ему отправлено приглашение.");
      setReplaceTargetId(null);
      setReplaceEmail("");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  function updateMember(index: number, value: string) {
    setMembers((prev) => prev.map((m, i) => (i === index ? value : m)));
  }

  return (
    <div className="grid gap-0">

      {/* ── Banner header ─────────────────────────────────────────────────── */}
      <div
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          background: "#111111",
          borderBottom: "1px solid #2d2d2d",
          position: "relative",
          overflow: "hidden",
        }}
      >
        {/* color glow based on status */}
        <div
          aria-hidden
          style={{
            position: "absolute",
            bottom: 0,
            left: 0,
            right: 0,
            height: 2,
            background: color,
          }}
        />

        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="grid gap-6">
            {/* status + title */}
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <span
                  className="h-2 w-2 rounded-full shrink-0"
                  style={{ background: color }}
                />
                <span className="text-xs font-semibold uppercase tracking-widest" style={{ color }}>
                  {tournamentStatusLabel[tournament.status]}
                </span>
              </div>
              <h1
                className="font-black uppercase text-white"
                style={{ fontSize: "clamp(1.8rem, 5vw, 3.5rem)", letterSpacing: "-0.03em", lineHeight: 1.1 }}
              >
                {tournament.title}
              </h1>
              {tournament.description && (
                <p className="max-w-2xl text-sm text-[#9e9e9e] leading-relaxed">
                  {tournament.description}
                </p>
              )}
            </div>

            {/* meta chips */}
            <div className="flex flex-wrap gap-2">
              {tournament.discipline && (
                <span className="flex items-center gap-1.5 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-1.5 text-xs text-[#9e9e9e]">
                  <Swords className="h-3.5 w-3.5" />
                  {tournament.discipline}
                </span>
              )}
              <span className="flex items-center gap-1.5 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-1.5 text-xs text-[#9e9e9e]">
                <Shield className="h-3.5 w-3.5" />
                {tournamentFormatLabel[tournament.format] ?? tournament.format}
              </span>
              {tournament.max_teams && (
                <span className="flex items-center gap-1.5 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-1.5 text-xs text-[#9e9e9e]">
                  <Users className="h-3.5 w-3.5" />
                  до {tournament.max_teams} участников
                </span>
              )}
              {tournament.start_at && (
                <span className="flex items-center gap-1.5 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-1.5 text-xs text-[#9e9e9e]">
                  <Calendar className="h-3.5 w-3.5" />
                  {formatDate(tournament.start_at)}
                </span>
              )}
              {tournament.location && (
                <span className="flex items-center gap-1.5 rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-1.5 text-xs text-[#9e9e9e]">
                  <MapPin className="h-3.5 w-3.5" />
                  {tournament.location}
                </span>
              )}
            </div>

            {/* actions */}
            <div className="flex flex-wrap gap-2">
              {canJoinIndividual && (
                <Button disabled={joinMutation.isPending} onClick={handleJoinIndividual}>
                  {joinMutation.isPending ? "Запись..." : "Записаться"}
                </Button>
              )}
              {canRegisterTeam && (
                <Button
                  variant={showRegisterForm ? "outline" : "default"}
                  onClick={() => setShowRegisterForm((v) => !v)}
                >
                  {showRegisterForm ? "Отмена" : "Записаться с командой"}
                </Button>
              )}
              {access.canAccessAdmin && (
                <Link to={`/tournaments/${id}/admin`}>
                  <Button variant="outline" className="gap-2">
                    <Settings className="h-4 w-4" /> Управление
                  </Button>
                </Link>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* ── Content ──────────────────────────────────────────────────────── */}
      <div className="py-8 grid gap-6">

        {/* register team form */}
        {showRegisterForm && canRegisterTeam && (
          <Card>
            <CardHeader><CardTitle>Регистрация команды</CardTitle></CardHeader>
            <CardContent className="grid gap-4">
              <div className="grid gap-1">
                <label className="text-sm text-[#9e9e9e]">Название команды</label>
                <Input placeholder="Введите название" value={teamName}
                  onChange={(e) => setTeamName(e.target.value)} className="md:max-w-sm" />
              </div>
              <div className="grid gap-2">
                <label className="text-sm text-[#9e9e9e]">Игроки (email адреса)</label>
                <p className="text-xs text-[#666666]">Вы будете капитаном. Добавьте других участников по email — они получат приглашение.</p>
                {members.map((member, index) => (
                  <div key={index} className="flex gap-2 md:max-w-sm">
                    <Input placeholder={`игрок${index + 2}@email.com`} value={member}
                      onChange={(e) => updateMember(index, e.target.value)} />
                    {members.length > 1 && (
                      <Button type="button" variant="outline" size="sm"
                        onClick={() => setMembers((p) => p.filter((_, i) => i !== index))}
                        className="shrink-0">×</Button>
                    )}
                  </div>
                ))}
                {members.length < 9 && (
                  <Button type="button" variant="outline" size="sm"
                    onClick={() => setMembers((p) => [...p, ""])} className="md:max-w-sm">
                    + Добавить игрока
                  </Button>
                )}
              </div>
              <div className="flex gap-3">
                <Button disabled={registerTeamMutation.isPending} onClick={() => void handleRegisterTeam()}>
                  {registerTeamMutation.isPending ? "Отправка..." : "Зарегистрировать команду"}
                </Button>
                <Button variant="outline" onClick={() => setShowRegisterForm(false)}>Отмена</Button>
              </div>
            </CardContent>
          </Card>
        )}

        {/* captain my-team panel */}
        {myTeamQuery.data && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>Моя команда — {myTeamQuery.data.team.name}</span>
                <TeamStatusBadge status={myTeamQuery.data.team.status} />
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3">
              <ul className="divide-y divide-[#2d2d2d]">
                {myTeamQuery.data.members.map((m) => (
                  <li key={m.id} className="flex items-center justify-between gap-3 py-2.5 text-sm">
                    <div className="flex items-center gap-2 min-w-0">
                      <span className="text-white truncate">{m.nickname}</span>
                      {m.is_captain && (
                        <span className="text-[10px] text-[#ff5500] uppercase font-bold">капитан</span>
                      )}
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      <MemberStatusBadge status={m.confirmation_status} />
                      {m.confirmation_status === "declined" && myTeamQuery.data.team.status !== "approved" && (
                        replaceTargetId === m.id ? (
                          <div className="flex items-center gap-1">
                            <Input className="h-7 w-40 text-xs" placeholder="email нового игрока"
                              value={replaceEmail} onChange={(e) => setReplaceEmail(e.target.value)} />
                            <Button size="sm" className="h-7 text-xs"
                              disabled={replaceMemberMutation.isPending}
                              onClick={() => void handleReplaceMember(myTeamQuery.data!.team.id, m.id)}>
                              ОК
                            </Button>
                            <Button size="sm" variant="ghost" className="h-7 text-xs"
                              onClick={() => { setReplaceTargetId(null); setReplaceEmail(""); }}>×</Button>
                          </div>
                        ) : (
                          <Button size="sm" variant="outline" className="h-7 text-xs"
                            onClick={() => { setReplaceTargetId(m.id); setReplaceEmail(""); }}>
                            Заменить
                          </Button>
                        )
                      )}
                    </div>
                  </li>
                ))}
              </ul>
              <p className="text-xs text-[#666666]">
                Для перехода в статус «Готова к проверке» нужно подтверждение минимум 4 игроков, включая капитана.
              </p>
            </CardContent>
          </Card>
        )}

        {/* tabs */}
        <Tabs value={tab} onValueChange={setTab} tabs={tabs} variant="underline" />

        {/* bracket */}
        {tab === "bracket" && (
          isIndividual ? (
            participantsQuery.isLoading || bracketQuery.isLoading ? <Spinner /> :
            (bracketQuery.data?.matches ?? []).length > 0 ? (
              <BracketView matches={bracketQuery.data?.matches ?? []} participants={participantsQuery.data?.items ?? []} />
            ) : (
              <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
            )
          ) : bracketQuery.isLoading ? <Spinner /> :
            bracketQuery.isError ? <ErrorState /> :
            bracketQuery.data?.bracket?.format === "group_stage" ? (
              <div className="space-y-6">
                {bracketQuery.data.bracket?.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold uppercase tracking-wide text-[#9e9e9e]">Плей-офф</h3>
                    <BracketView matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)} teams={teamsQuery.data?.items ?? []} />
                  </div>
                )}
                {(bracketQuery.data.groups ?? []).length > 0 ? (
                  <GroupStageView groups={bracketQuery.data.groups ?? []} matches={bracketQuery.data.matches ?? []} teams={teamsQuery.data?.items ?? []} />
                ) : (
                  <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
                )}
              </div>
            ) : bracketQuery.data?.bracket?.format === "group_de" ? (
              <div className="space-y-6">
                {bracketQuery.data.bracket?.status === "playoff" && (bracketQuery.data.matches ?? []).filter((m) => !m.group_id).length > 0 && (
                  <div className="space-y-2">
                    <h3 className="text-sm font-semibold uppercase tracking-wide text-[#9e9e9e]">Плей-офф</h3>
                    <BracketView matches={(bracketQuery.data.matches ?? []).filter((m) => !m.group_id)} teams={teamsQuery.data?.items ?? []} />
                  </div>
                )}
                {(bracketQuery.data.groups ?? []).length > 0 ? (
                  <GroupDEView groups={bracketQuery.data.groups ?? []} matches={bracketQuery.data.matches ?? []} teams={teamsQuery.data?.items ?? []} />
                ) : (
                  <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
                )}
              </div>
            ) : (bracketQuery.data?.matches ?? []).length > 0 ? (
              <BracketView matches={bracketQuery.data?.matches ?? []} teams={teamsQuery.data?.items ?? []} participants={participantsQuery.data?.items ?? []} />
            ) : (
              <EmptyState title="Сетка не сгенерирована" description="Администратор ещё не запустил сетку." />
            )
        )}

        {/* teams / participants */}
        {tab === "teams" && (
          isIndividual ? (
            participantsQuery.isLoading ? <Spinner /> :
            (participantsQuery.data?.items ?? []).length > 0 ? (
              <Card>
                <CardContent className="pt-5">
                  <ul className="divide-y divide-[#2d2d2d]">
                    {[...(participantsQuery.data?.items ?? [])].sort((a, b) => a.seed - b.seed).map((p) => (
                      <li key={p.id} className="flex items-center gap-3 py-2.5">
                        <span className="w-6 text-right text-xs text-[#666666]">{p.seed}</span>
                        <span className="text-sm text-white">{p.name}</span>
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            ) : (
              <EmptyState title="Участников пока нет" description="Никто ещё не записался на турнир." />
            )
          ) : teamsQuery.isLoading ? <Spinner /> :
            teamsQuery.isError ? <ErrorState /> :
            teamsQuery.data?.items.length ? (
              <>
                <TeamsTable teams={teamsQuery.data.items} onOpen={setDetailsTeamId} />
                {detailsTeamId && teamDetailsQuery.isLoading ? <Spinner /> : null}
                {teamDetailsQuery.data && detailsTeamId ? (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <p className="text-xs text-[#666666]">Состав команды</p>
                      <Button variant="ghost" size="sm" className="text-xs" onClick={() => setDetailsTeamId(null)}>
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
        )}

        {/* matches */}
        {tab === "matches" && !isFinished && (
          matchesQuery.isLoading ? <Spinner /> :
          matchesQuery.isError ? <ErrorState /> :
          matchesQuery.data?.items.length ? (
            <MatchesTable matches={matchesQuery.data.items} teams={teamsQuery.data?.items ?? []} participants={participantsQuery.data?.items ?? []} />
          ) : (
            <EmptyState title="Матчей пока нет" description="Матчи появятся после генерации сетки." />
          )
        )}

        {/* results */}
        {(tab === "results" || (tab === "matches" && isFinished)) && (
          matchesQuery.isLoading ? <Spinner /> : (
            <ResultsView
              matches={matchesQuery.data?.items ?? []}
              teams={teamsQuery.data?.items ?? []}
              participants={participantsQuery.data?.items ?? []}
            />
          )
        )}

        {/* placements */}
        {tab === "places" && (
          placementsQuery.isLoading ? <Spinner /> :
          (placementsQuery.data?.placements ?? []).length > 0 ? (
            <PlacementsView placements={placementsQuery.data!.placements} />
          ) : (
            <EmptyState title="Места ещё не определены" description="Распределение мест обновляется по ходу турнира." />
          )
        )}

        {/* rules */}
        {tab === "rules" && (
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader><CardTitle>Правила</CardTitle></CardHeader>
              <CardContent className="text-sm text-[#9e9e9e] leading-relaxed">
                {tournament.rules || "Не заполнено"}
              </CardContent>
            </Card>
            {tournament.location && (
              <Card>
                <CardHeader><CardTitle>Место проведения</CardTitle></CardHeader>
                <CardContent className="text-sm text-[#9e9e9e]">{tournament.location}</CardContent>
              </Card>
            )}
          </div>
        )}

      </div>
    </div>
  );
}
