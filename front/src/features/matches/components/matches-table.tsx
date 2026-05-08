import type { Match } from "@/shared/types/api";
import { formatDateTime } from "@/shared/lib/date";
import { matchStatusLabel, matchTeamConfirmationLabel } from "@/shared/lib/enums";
import { pickTeamName } from "@/shared/lib/bracket";
import { Badge } from "@/shared/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { Button } from "@/shared/ui/button";

function tone(status: string) {
  if (status === "finished" || status === "confirmed") return "success";
  if (status === "issue_reported" || status === "cancelled") return "danger";
  if (status === "awaiting_confirmation" || status === "reschedule_requested") return "warning";
  return "muted";
}

export function MatchesTable({
  matches,
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
  adminMode?: boolean;
  onSchedule?: (match: Match) => void;
  onConfirmReady?: (match: Match) => void;
  onReschedule?: (match: Match) => void;
  onIssue?: (match: Match) => void;
  onSubmitResult?: (match: Match) => void;
  onApprove?: (match: Match) => void;
  onReject?: (match: Match) => void;
}) {
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
          {matches.map((match) => (
            <TableRow key={match.id}>
              <TableCell>
                <div className="font-medium text-white">
                  {pickTeamName(match, "home")} vs {pickTeamName(match, "away")}
                </div>
                <div className="text-xs text-[#90afd4]">
                  Раунд {match.round_number ?? "—"} · Слот {match.slot_index ?? "—"}
                </div>
              </TableCell>
              <TableCell>
                <Badge tone={tone(match.status)}>{matchStatusLabel[match.status]}</Badge>
              </TableCell>
              <TableCell>
                <div className="space-y-1 text-xs">
                  <div>
                    A:{" "}
                    {match.home_team_confirmation_status
                      ? matchTeamConfirmationLabel[match.home_team_confirmation_status]
                      : "—"}
                  </div>
                  <div>
                    B:{" "}
                    {match.away_team_confirmation_status
                      ? matchTeamConfirmationLabel[match.away_team_confirmation_status]
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