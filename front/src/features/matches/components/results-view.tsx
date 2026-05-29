import type { Match, Team } from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";
import { Card, CardContent } from "@/shared/ui/card";
import { Trophy } from "lucide-react";
import { formatDateTime } from "@/shared/lib/date";
import { useLang } from "@/app/providers/lang-provider";

function resolveName(
  id: string | null | undefined,
  teamsById: Map<string, Team>,
  participantsById: Map<string, Participant>,
): string {
  if (!id) return "BYE";
  const team = teamsById.get(id);
  if (team) return team.name;
  const p = participantsById.get(id);
  if (p) return p.name;
  return id.slice(0, 8) + "…";
}

function ResultCard({
  match,
  teamsById,
  participantsById,
  t,
}: {
  match: Match;
  teamsById: Map<string, Team>;
  participantsById: Map<string, Participant>;
  t: (key: string, vars?: Record<string, string | number>) => string;
}) {
  const side1Id = match.team1_id ?? match.participant1_id;
  const side2Id = match.team2_id ?? match.participant2_id;
  const winnerId = match.winner_team_id ?? match.winner_participant_id;

  const name1 = resolveName(side1Id, teamsById, participantsById);
  const name2 = resolveName(side2Id, teamsById, participantsById);

  const win1 = !!winnerId && winnerId === side1Id;
  const win2 = !!winnerId && winnerId === side2Id;

  return (
    <Card className="border-[#2d2d2d] bg-[#1a1a1a]">
      <CardContent className="p-4 space-y-2">
        {match.scheduled_at && (
          <p className="text-xs text-[#666666]">{formatDateTime(match.scheduled_at)}</p>
        )}

        <div className="space-y-1">
          {[
            { name: name1, won: win1 },
            { name: name2, won: win2 },
          ].map(({ name, won }) => (
            <div
              key={name}
              className={`flex items-center justify-between rounded px-3 py-1.5 text-sm ${
                won
                  ? "bg-[#ff5500]/20 font-semibold text-white"
                  : "text-[#9e9e9e]"
              }`}
            >
              <span className="truncate">{name}</span>
              {won && <Trophy className="ml-2 h-3.5 w-3.5 shrink-0 text-yellow-400" />}
            </div>
          ))}
        </div>

        {match.score_text && (
          <p className="text-center text-xs font-medium text-[#9e9e9e]">{match.score_text}</p>
        )}

        {match.round_number != null && (
          <p className="text-xs text-[#666666]">{t("results.round", { n: match.round_number })}</p>
        )}
      </CardContent>
    </Card>
  );
}

export function ResultsView({
  matches,
  teams = [],
  participants = [],
}: {
  matches: Match[];
  teams?: Team[];
  participants?: Participant[];
}) {
  const { t } = useLang();
  const finished = matches.filter((m) => m.status === "finished" || m.status === "confirmed");

  const teamsById = new Map(teams.map((t) => [t.id, t]));
  const participantsById = new Map(participants.map((p) => [p.id, p]));

  if (finished.length === 0) {
    return (
      <div className="rounded-xl border border-[#2d2d2d] px-6 py-10 text-center text-sm text-[#666666]">
        {t("results.empty")}
      </div>
    );
  }

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {finished.map((m) => (
        <ResultCard key={m.id} match={m} teamsById={teamsById} participantsById={participantsById} t={t} />
      ))}
    </div>
  );
}
