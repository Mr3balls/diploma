import { useState } from "react";
import { Trophy } from "lucide-react";
import type { Match, Team } from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";
import { buildRounds, buildTeamsById, pickSideName, isMatchWinner } from "@/shared/lib/bracket";
import { matchStatusLabel } from "@/shared/lib/enums";
import { useAdminSetResult } from "@/features/matches/hooks";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { formatDateTime } from "@/shared/lib/date";

function MatchCard({
  match,
  teamsById,
  participantsById,
  adminMode,
  tournamentId,
}: {
  match: Match;
  teamsById: Map<string, Team>;
  participantsById: Map<string, Participant>;
  adminMode: boolean;
  tournamentId?: string;
}) {
  const [picking, setPicking] = useState(false);
  const adminSetResult = useAdminSetResult(tournamentId ?? "");

  const p1name = pickSideName(match, "1", teamsById, participantsById);
  const p2name = pickSideName(match, "2", teamsById, participantsById);
  const won1 = isMatchWinner(match, "1");
  const won2 = isMatchWinner(match, "2");
  const isFinished = match.status === "finished";

  const canSubmit =
    adminMode &&
    Boolean(tournamentId) &&
    !isFinished &&
    !match.is_bye &&
    Boolean(match.team1_id) &&
    Boolean(match.team2_id);

  function submit(teamId: string) {
    adminSetResult.mutate(
      { matchId: match.id, payload: { winner_team_id: teamId } },
      { onSuccess: () => setPicking(false) },
    );
  }

  return (
    <Card className="border-[#0a3575] bg-[#001a4a]">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xs text-[#90afd4]">
            {match.bracket_section} · #{match.slot_index ?? "—"}
          </CardTitle>
          <Badge tone={isFinished ? "success" : "muted"}>
            {matchStatusLabel[match.status] ?? match.status}
          </Badge>
        </div>
        {match.scheduled_at && (
          <p className="text-xs text-[#4a7ab5]">{formatDateTime(match.scheduled_at)}</p>
        )}
      </CardHeader>
      <CardContent className="space-y-1 pt-0">
        {(
          [
            ["1", p1name, won1],
            ["2", p2name, won2],
          ] as const
        ).map(([side, name, won]) => (
          <div
            key={side}
            className={`flex items-center justify-between rounded-lg px-2 py-1 text-sm ${
              won ? "bg-[#2255ff]/20 font-semibold text-white" : "text-[#90afd4]"
            }`}
          >
            <span className="truncate">{name}</span>
            {won && <Trophy className="ml-2 h-3 w-3 shrink-0 text-yellow-400" />}
          </div>
        ))}
        {match.score_text && (
          <p className="pt-1 text-center text-xs text-[#90afd4]">{match.score_text}</p>
        )}
        {canSubmit && !picking && (
          <Button
            size="sm"
            variant="secondary"
            className="mt-2 w-full"
            onClick={() => setPicking(true)}
          >
            Указать победителя
          </Button>
        )}
        {picking && (
          <div className="mt-2 space-y-1">
            <p className="text-xs text-[#90afd4]">Победитель:</p>
            {match.team1_id && (
              <Button
                size="sm"
                variant="outline"
                className="w-full justify-start"
                disabled={adminSetResult.isPending}
                onClick={() => submit(match.team1_id!)}
              >
                {p1name}
              </Button>
            )}
            {match.team2_id && (
              <Button
                size="sm"
                variant="outline"
                className="w-full justify-start"
                disabled={adminSetResult.isPending}
                onClick={() => submit(match.team2_id!)}
              >
                {p2name}
              </Button>
            )}
            <Button size="sm" variant="ghost" className="w-full" onClick={() => setPicking(false)}>
              Отмена
            </Button>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export function BracketView({
  matches,
  teams = [],
  participants = [],
  adminMode = false,
  tournamentId,
}: {
  matches: Match[];
  teams?: Team[];
  participants?: Participant[];
  adminMode?: boolean;
  tournamentId?: string;
}) {
  const visibleMatches = matches.filter((m) => !m.is_bye);
  const rounds = buildRounds(visibleMatches);
  const teamsById = buildTeamsById(teams);
  const participantsById = new Map(participants.map((p) => [p.id, p]));

  if (!rounds.length) {
    return (
      <Card className="border-[#0a3575]">
        <CardContent className="py-8 text-sm text-[#90afd4]">
          Сетка пока не создана.
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="overflow-x-auto pb-4">
      <div className="flex min-w-max gap-4">
        {rounds.map((round) => (
          <div
            key={`${round.section}-${round.roundNumber}`}
            className="w-[230px] shrink-0 space-y-3"
          >
            <div className="flex items-center justify-between">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-[#90afd4]">
                {round.section !== "WB" ? `${round.section} · ` : ""}Раунд {round.roundNumber}
              </h3>
              {adminMode && <Badge tone="muted">Admin</Badge>}
            </div>
            {round.matches.map((match) => (
              <MatchCard
                key={match.id}
                match={match}
                teamsById={teamsById}
                participantsById={participantsById}
                adminMode={adminMode}
                tournamentId={tournamentId}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}
