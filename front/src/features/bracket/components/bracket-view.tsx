import type { Match } from "@/shared/types/api";
import { buildRounds, pickTeamName } from "@/shared/lib/bracket";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { formatDateTime } from "@/shared/lib/date";

function winnerClass(match: Match, side: "home" | "away") {
  const teamId = side === "home" ? match.home_team_id || match.home_team?.id : match.away_team_id || match.away_team?.id;
  return match.winner_team_id && match.winner_team_id === teamId ? "font-semibold text-slate-900" : "text-slate-600";
}

export function BracketView({ matches, adminMode = false }: { matches: Match[]; adminMode?: boolean }) {
  const rounds = buildRounds(matches);

  if (!rounds.length) {
    return (
      <Card>
        <CardContent className="py-8 text-sm text-slate-500">
          Сетка пока не создана.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="overflow-x-auto pb-2">
      <div className="flex min-w-max gap-4">
        {rounds.map((round) => (
          <div key={round.roundNumber} className="w-[280px] shrink-0 space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold">Раунд {round.roundNumber}</h3>
              {adminMode ? <Badge tone="muted">Admin</Badge> : null}
            </div>
            {round.matches.map((match) => (
              <Card key={match.id}>
                <CardHeader className="gap-2 pb-3">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-sm">Матч #{match.slot_index ?? "—"}</CardTitle>
                    {match.is_bye ? <Badge tone="warning">BYE</Badge> : null}
                  </div>
                  <p className="text-xs text-slate-500">{formatDateTime(match.scheduled_at)}</p>
                </CardHeader>
                <CardContent className="grid gap-3">
                  <div className={winnerClass(match, "home")}>{pickTeamName(match, "home")}</div>
                  <div className={winnerClass(match, "away")}>{pickTeamName(match, "away")}</div>
                  <div className="flex items-center justify-between text-xs text-slate-500">
                    <span>{match.status}</span>
                    <span>{match.score_text || "Без счёта"}</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}