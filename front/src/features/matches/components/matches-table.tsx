import type { Match, Team } from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";
import { formatDateTime } from "@/shared/lib/date";
import { matchStatusLabel, matchTeamConfirmationLabel } from "@/shared/lib/enums";
import { buildTeamsById, pickTeamName } from "@/shared/lib/bracket";
import { Badge } from "@/shared/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { Button } from "@/shared/ui/button";

function tone(status: string) {
  if (status === "finished" || status === "confirmed") return "success";
  if (status === "issue_reported" || status === "cancelled") return "danger";
  if (status === "awaiting_confirmation" || status === "reschedule_requested") return "warning";
  return "muted";
}

function pickSideName(
  match: Match,
  side: "1" | "2",
  teamsById: Map<string, Team>,
  participantsById: Map<string, Participant>,
): string {
  const teamId = side === "1" ? match.team1_id : match.team2_id;
  if (teamId) {
    const team = teamsById.get(teamId);
    if (team?.name) return team.name;
    return teamId.slice(0, 8) + "…";
  }
  const participantId = side === "1" ? match.participant1_id : match.participant2_id;
  if (participantId) {
    const p = participantsById.get(participantId);
    if (p?.name) return p.name;
    return participantId.slice(0, 8) + "…";
  }
  return "BYE";
}

export function MatchesTable({
  matches,
  teams = [],
  participants = [],
  adminMode = false,
  onSchedule,
  onConfirmReady,
  onReschedule,
  onIssue,
  onSubmitResult,
  onApprove,
  onReject,
}: {
  matches: Match[];
  teams?: Team[];
  participants?: Participant[];
  adminMode?: boolean;
  onSchedule?: (match: Match) => void;
  onConfirmReady?: (match: Match) => void;
  onReschedule?: (match: Match) => void;
  onIssue?: (match: Match) => void;
  onSubmitResult?: (match: Match) => void;
  onApprove?: (match: Match) => void;
  onReject?: (match: Match) => void;
}) {
  const teamsById = buildTeamsById(teams);
  const participantsById = new Map(participants.map((p) => [p.id, p]));
  const visibleMatches = matches.filter((m) => !m.is_bye);

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Матч</TableHead>
            <TableHead>Статус</TableHead>
            <TableHead>Готовность</TableHead>
            <TableHead>Время</TableHead>
            <TableHead>Счёт</TableHead>
            {adminMode ? <TableHead>Действия</TableHead> : null}
          </TableRow>
        </TableHeader>
        <TableBody>
          {visibleMatches.map((match) => (
            <TableRow key={match.id}>
              <TableCell>
                <div className="font-medium text-white">
                  {pickSideName(match, "1", teamsById, participantsById)} vs {pickSideName(match, "2", teamsById, participantsById)}
                </div>
                <div className="text-xs text-[#90afd4]">
                  {match.bracket_section ? `${match.bracket_section} · ` : ""}Раунд {match.round_number ?? "—"} · Слот {match.slot_index ?? "—"}
                </div>
              </TableCell>
              <TableCell>
                <Badge tone={tone(match.status)}>{matchStatusLabel[match.status]}</Badge>
              </TableCell>
              <TableCell>
                <div className="space-y-1 text-xs">
                  <div>
                    A:{" "}
                    {match.team1_confirmation_status
                      ? matchTeamConfirmationLabel[match.team1_confirmation_status]
                      : "—"}
                  </div>
                  <div>
                    B:{" "}
                    {match.team2_confirmation_status
                      ? matchTeamConfirmationLabel[match.team2_confirmation_status]
                      : "—"}
                  </div>
                </div>
              </TableCell>
              <TableCell>{formatDateTime(match.scheduled_at)}</TableCell>
              <TableCell>{match.score_text || "—"}</TableCell>
              {adminMode ? (
                <TableCell>
                  <div className="flex flex-wrap gap-2">
                    {onSchedule ? (
                      <Button variant="outline" size="sm" onClick={() => onSchedule(match)}>
                        Время
                      </Button>
                    ) : null}
                    {onConfirmReady ? (
                      <Button variant="outline" size="sm" onClick={() => onConfirmReady(match)}>
                        Готов
                      </Button>
                    ) : null}
                    {onReschedule ? (
                      <Button variant="outline" size="sm" onClick={() => onReschedule(match)}>
                        Перенос
                      </Button>
                    ) : null}
                    {onIssue ? (
                      <Button variant="outline" size="sm" onClick={() => onIssue(match)}>
                        Проблема
                      </Button>
                    ) : null}
                    {onSubmitResult ? (
                      <Button size="sm" onClick={() => onSubmitResult(match)}>
                        Результат
                      </Button>
                    ) : null}
                    {onApprove ? (
                      <Button size="sm" onClick={() => onApprove(match)}>
                        Принять
                      </Button>
                    ) : null}
                    {onReject ? (
                      <Button variant="destructive" size="sm" onClick={() => onReject(match)}>
                        Отклонить
                      </Button>
                    ) : null}
                  </div>
                </TableCell>
              ) : null}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
