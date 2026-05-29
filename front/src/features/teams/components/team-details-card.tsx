import type { TeamDetailsResponse } from "@/shared/types/api";
import { useLang } from "@/app/providers/lang-provider";
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
  const { t } = useLang();

  return (
    <Card>
      <CardHeader className="gap-3">
        <div className="flex items-center justify-between gap-4">
          <CardTitle>{data.team.name}</CardTitle>
          <Badge tone={tone(data.team.status)}>{t(`teamStatus.${data.team.status}`)}</Badge>
        </div>
        <p className="text-sm text-[#9e9e9e]">{t("teamCard.readyHint")}</p>
      </CardHeader>
      <CardContent className="grid gap-4">
        <div className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("teamCard.nickname")}</TableHead>
                <TableHead>{t("teamCard.role")}</TableHead>
                <TableHead>{t("teamCard.status")}</TableHead>
                {allowAdminActions ? <TableHead /> : null}
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.members.map((member) => (
                <TableRow key={member.id}>
                  <TableCell className="font-medium text-white">
                    {member.nickname || member.display_name || "—"}
                  </TableCell>
                  <TableCell>{member.role}</TableCell>
                  <TableCell>
                    <Badge tone={tone(member.confirmation_status)}>
                      {t(`memberStatus.${member.confirmation_status}`)}
                    </Badge>
                  </TableCell>
                  {allowAdminActions ? (
                    <TableCell>
                      {onRemoveMember ? (
                        <Button variant="outline" size="sm" onClick={() => onRemoveMember(member.id)}>
                          {t("teamCard.remove")}
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
