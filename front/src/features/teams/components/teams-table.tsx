import type { Team } from "@/shared/types/api";
import { useLang } from "@/app/providers/lang-provider";
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
  onDelete,
}: {
  teams: Team[];
  onOpen?: (id: string) => void;
  withActions?: boolean;
  onApprove?: (id: string) => void;
  onReject?: (id: string) => void;
  onDelete?: (id: string) => void;
}) {
  const { t } = useLang();

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("teamsTable.name")}</TableHead>
            {withActions ? <TableHead>{t("teamsTable.status")}</TableHead> : null}
            <TableHead>{t("teamsTable.captain")}</TableHead>
            {withActions ? <TableHead>{t("teamsTable.confirmedPlayers")}</TableHead> : null}
            {withActions ? <TableHead>{t("teamsTable.duplicates")}</TableHead> : null}
            {(withActions || onOpen) ? <TableHead>{t("teamsTable.actions")}</TableHead> : null}
          </TableRow>
        </TableHeader>
        <TableBody>
          {teams.map((team) => (
            <TableRow key={team.id}>
              <TableCell className="font-medium text-white">{team.name}</TableCell>
              {withActions ? (
                <TableCell>
                  <Badge tone={tone(team.status)}>{t(`teamStatus.${team.status}`)}</Badge>
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
                        {t("teamsTable.roster")}
                      </Button>
                    ) : null}
                    {withActions && onApprove ? (
                      <Button size="sm" onClick={() => onApprove(team.id)}>
                        {t("teamsTable.approve")}
                      </Button>
                    ) : null}
                    {withActions && onReject ? (
                      <Button variant="destructive" size="sm" onClick={() => onReject(team.id)}>
                        {t("teamsTable.reject")}
                      </Button>
                    ) : null}
                    {withActions && onDelete ? (
                      <Button variant="destructive" size="sm" onClick={() => onDelete(team.id)}>
                        {t("teamsTable.delete")}
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
