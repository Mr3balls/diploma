import type { Team } from "@/shared/types/api";
import { teamStatusLabel } from "@/shared/lib/enums";
import { Badge } from "@/shared/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { Button } from "@/shared/ui/button";

function tone(status: Team["status"]) {
  if (status === "approved") return "success";
  if (status === "rejected") return "danger";
  if (status === "ready_for_review") return "warning";
  return "muted";
}

export function TeamsTable({
  teams,
  onOpen,
  withActions = false,
  onApprove,
  onReject,
}: {
  teams: Team[];
  onOpen?: (id: string) => void;
  withActions?: boolean;
  onApprove?: (id: string) => void;
  onReject?: (id: string) => void;
}) {
  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Название</TableHead>
            {withActions ? <TableHead>Статус</TableHead> : null}
            <TableHead>Капитан</TableHead>
            {withActions ? <TableHead>Подтв. игроков</TableHead> : null}
            {withActions ? <TableHead>Дубликаты</TableHead> : null}
            {(withActions || onOpen) ? <TableHead>Действия</TableHead> : null}
          </TableRow>
        </TableHeader>
        <TableBody>
          {teams.map((team) => (
            <TableRow key={team.id}>
              <TableCell className="font-medium text-white">{team.name}</TableCell>
              {withActions ? (
                <TableCell>
                  <Badge tone={tone(team.status)}>{teamStatusLabel[team.status]}</Badge>
                </TableCell>
              ) : null}
              <TableCell>{team.captain_nickname || "—"}</TableCell>
              {withActions ? (
                <TableCell>{team.confirmed_main_players_count ?? "—"}</TableCell>
              ) : null}
              {withActions ? (
                <TableCell>
                  {team.duplicate_conflicts?.length ? (
                    <div className="flex flex-col gap-1">
                      {team.duplicate_conflicts.map((item) => (
                        <span key={item} className="text-xs text-amber-700">
                          {item}
                        </span>
                      ))}
                    </div>
                  ) : (
                    "—"
                  )}
                </TableCell>
              ) : null}
              {(withActions || onOpen) ? (
                <TableCell>
                  <div className="flex flex-wrap gap-2">
                    {onOpen ? (
                      <Button variant="outline" size="sm" onClick={() => onOpen(team.id)}>
                        Состав
                      </Button>
                    ) : null}
                    {withActions && onApprove ? (
                      <Button size="sm" onClick={() => onApprove(team.id)}>
                        Одобрить
                      </Button>
                    ) : null}
                    {withActions && onReject ? (
                      <Button variant="destructive" size="sm" onClick={() => onReject(team.id)}>
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