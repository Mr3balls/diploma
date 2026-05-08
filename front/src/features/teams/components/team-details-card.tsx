import type { TeamDetailsResponse } from "@/shared/types/api";
import { memberStatusLabel, teamStatusLabel } from "@/shared/lib/enums";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { Button } from "@/shared/ui/button";

function tone(status: string) {
  if (status === "confirmed" || status === "approved") return "success";
  if (status === "declined" || status === "rejected" || status === "removed") return "danger";
  if (status === "pending_confirmation" || status === "ready_for_review") return "warning";
  return "muted";
}

export function TeamDetailsCard({
  data,
  allowAdminActions = false,
  onRemoveMember,
}: {
  data: TeamDetailsResponse;
  allowAdminActions?: boolean;
  onRemoveMember?: (memberId: string) => void;
}) {
  return (
    <Card>
      <CardHeader className="gap-3">
        <div className="flex items-center justify-between gap-4">
          <CardTitle>{data.team.name}</CardTitle>
          <Badge tone={tone(data.team.status)}>{teamStatusLabel[data.team.status]}</Badge>
        </div>
        <p className="text-sm text-slate-500">
          Команда становится готовой к проверке только если капитан подтверждён и подтверждено не менее 4 основных
          игроков.
        </p>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Никнейм</TableHead>
                <TableHead>Роль</TableHead>
                <TableHead>Статус</TableHead>
                {allowAdminActions ? <TableHead /> : null}
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.members.map((member) => (
                <TableRow key={member.id}>
                  <TableCell className="font-medium text-slate-900">
                    {member.nickname || member.display_name || "—"}
                  </TableCell>
                  <TableCell>{member.role}</TableCell>
                  <TableCell>
                    <Badge tone={tone(member.confirmation_status)}>
                      {memberStatusLabel[member.confirmation_status]}
                    </Badge>
                  </TableCell>
                  {allowAdminActions ? (
                    <TableCell>
                      {onRemoveMember ? (
                        <Button variant="outline" size="sm" onClick={() => onRemoveMember(member.id)}>
                          Удалить
                        </Button>
                      ) : null}
                    </TableCell>
                  ) : null}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      </CardContent>
    </Card>
  );
}