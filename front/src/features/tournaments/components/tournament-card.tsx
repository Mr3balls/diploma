import { Link } from "react-router-dom";
import type { Tournament } from "@/shared/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Badge } from "@/shared/ui/badge";
import { formatDate } from "@/shared/lib/date";
import { tournamentStatusLabel, visibilityLabel } from "@/shared/lib/enums";

function statusTone(status: Tournament["status"]) {
  switch (status) {
    case "finished":
    case "completed":
      return "success";
    case "cancelled":
      return "danger";
    case "in_progress":
      return "warning";
    case "registration_open":
    case "bracket_generated":
    case "ready":
      return "default";
    default:
      return "muted";
  }
}

export function TournamentCard({ tournament }: { tournament: Tournament }) {
  return (
    <Link to={`/tournaments/${tournament.id}`}>
      <Card className="h-full transition-transform hover:-translate-y-0.5">
        <CardHeader className="gap-3">
          <div className="flex flex-wrap items-center gap-2">
            <Badge tone={statusTone(tournament.status)}>{tournamentStatusLabel[tournament.status]}</Badge>
            <Badge tone="muted">{visibilityLabel[tournament.visibility]}</Badge>
          </div>
          <CardTitle>{tournament.title}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <p className="line-clamp-3 text-sm text-[#90afd4]">{tournament.description || "Описание не заполнено."}</p>
          <div className="grid gap-2 text-sm text-[#90afd4]">
            <div>Игра: {tournament.discipline || "—"}</div>
            <div>Макс. команд: {tournament.max_teams ?? "—"}</div>
            <div>Создан: {formatDate(tournament.created_at)}</div>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}