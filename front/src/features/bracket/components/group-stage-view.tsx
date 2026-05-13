import type { BracketGroup, Match, Team } from "@/shared/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Badge } from "@/shared/ui/badge";
import { matchStatusLabel } from "@/shared/lib/enums";
import { formatDateTime } from "@/shared/lib/date";
import { Button } from "@/shared/ui/button";
import { useAdminSetResult } from "@/features/matches/hooks";
import { useState } from "react";
import { Trophy } from "lucide-react";

function GroupTable({ group, teamsById }: { group: BracketGroup; teamsById: Map<string, Team> }) {
  return (
    <Card className="border-[#0a3575] bg-[#001a4a]">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm text-white">{group.name}</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-[#0a3575] text-[#90afd4]">
              <th className="px-3 py-2 text-left">Команда</th>
              <th className="px-2 py-2 text-center">И</th>
              <th className="px-2 py-2 text-center">В</th>
              <th className="px-2 py-2 text-center">П</th>
              <th className="px-2 py-2 text-center">О</th>
            </tr>
          </thead>
          <tbody>
            {group.members.map((m, idx) => {
              const team = teamsById.get(m.team_id);
              const played = m.wins + m.losses + m.draws;
              return (
                <tr
                  key={m.id}
                  className={`border-b border-[#0a3575]/50 ${idx < 2 ? "text-white" : "text-[#90afd4]"}`}
                >
                  <td className="px-3 py-2">
                    {idx < 2 && <span className="mr-1 text-yellow-400">↑</span>}
                    {team?.name ?? m.team_id}
                  </td>
                  <td className="px-2 py-2 text-center">{played}</td>
                  <td className="px-2 py-2 text-center">{m.wins}</td>
                  <td className="px-2 py-2 text-center">{m.losses}</td>
                  <td className="px-2 py-2 text-center font-semibold">{m.points}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
        <p className="px-3 py-1 text-xs text-[#4a7ab5]">↑ выходят в плей-офф (топ 2)</p>
      </CardContent>
    </Card>
  );
}

function GroupMatchCard({
  match,
  teamsById,
  adminMode,
  tournamentId,
}: {
  match: Match;
  teamsById: Map<string, Team>;
  adminMode: boolean;
  tournamentId?: string;
}) {
  const [picking, setPicking] = useState(false);
  const adminSetResult = useAdminSetResult(tournamentId ?? "");
  const t1 = match.team1_id ? (teamsById.get(match.team1_id)?.name ?? match.team1_id) : "TBD";
  const t2 = match.team2_id ? (teamsById.get(match.team2_id)?.name ?? match.team2_id) : "TBD";
  const isFinished = match.status === "finished";

  function submit(teamId: string) {
    adminSetResult.mutate(
      { matchId: match.id, payload: { winner_team_id: teamId } },
      { onSuccess: () => setPicking(false) },
    );
  }

  return (
    <Card className="border-[#0a3575] bg-[#001a4a]">
      <CardContent className="space-y-1 p-3">
        <div className="flex items-center justify-between">
          <Badge tone={isFinished ? "success" : "muted"}>{matchStatusLabel[match.status] ?? match.status}</Badge>
          {match.scheduled_at && <span className="text-xs text-[#4a7ab5]">{formatDateTime(match.scheduled_at)}</span>}
        </div>
        {([
          [match.team1_id, t1],
          [match.team2_id, t2],
        ] as const).map(([tid, name]) => {
          const won = isFinished && match.winner_team_id === tid;
          return (
            <div
              key={tid ?? name}
              className={`flex items-center justify-between rounded px-2 py-1 text-sm ${
                won ? "bg-[#2255ff]/20 font-semibold text-white" : "text-[#90afd4]"
              }`}
            >
              <span className="truncate">{name}</span>
              {won && <Trophy className="ml-2 h-3 w-3 shrink-0 text-yellow-400" />}
            </div>
          );
        })}
        {match.score_text && <p className="text-center text-xs text-[#90afd4]">{match.score_text}</p>}
        {adminMode && !isFinished && !picking && match.team1_id && match.team2_id && (
          <Button size="sm" variant="secondary" className="mt-1 w-full" onClick={() => setPicking(true)}>
            Указать победителя
          </Button>
        )}
        {picking && (
          <div className="space-y-1">
            {match.team1_id && (
              <Button
                size="sm"
                variant="outline"
                className="w-full justify-start"
                disabled={adminSetResult.isPending}
                onClick={() => submit(match.team1_id!)}
              >
                {t1}
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
                {t2}
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

export function GroupStageView({
  groups,
  matches,
  teams = [],
  adminMode = false,
  tournamentId,
}: {
  groups: BracketGroup[];
  matches: Match[];
  teams?: Team[];
  adminMode?: boolean;
  tournamentId?: string;
}) {
  const teamsById = new Map(teams.map((t) => [t.id, t]));

  // group_id is omitempty in JSON: present as string for group matches, absent for others
  const groupMatches = matches.filter((m) => m.group_id !== undefined && m.group_id !== null);

  const matchesByGroup = new Map<string, Match[]>();
  for (const m of groupMatches) {
    const gid = m.group_id as string;
    if (!matchesByGroup.has(gid)) matchesByGroup.set(gid, []);
    matchesByGroup.get(gid)!.push(m);
  }

  return (
    <div className="space-y-6">
      {/* Standings tables */}
      <div className="grid gap-4 md:grid-cols-2">
        {groups.map((g) => (
          <GroupTable key={g.id} group={g} teamsById={teamsById} />
        ))}
      </div>

      {/* Group matches — hidden after advancing to playoff (group matches get deleted) */}
      {groupMatches.length > 0 && (
        <div className="space-y-4">
          <h3 className="text-sm font-semibold uppercase tracking-wide text-[#90afd4]">Матчи группового этапа</h3>
          {groups.map((g) => {
            const gMatches = matchesByGroup.get(g.id) ?? [];
            if (gMatches.length === 0) return null;
            return (
              <div key={g.id} className="space-y-2">
                <p className="text-xs font-medium text-[#4a7ab5]">{g.name}</p>
                <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                  {gMatches.map((m) => (
                    <GroupMatchCard
                      key={m.id}
                      match={m}
                      teamsById={teamsById}
                      adminMode={adminMode}
                      tournamentId={tournamentId}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
