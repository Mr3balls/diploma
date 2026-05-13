import type { BracketGroup, Match, Team } from "@/shared/types/api";
import { BracketView } from "./bracket-view";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";

const QP_LABEL: Record<number, string> = {
  1: "→ Полуфинал",
  2: "→ Четвертьфинал",
  3: "→ Четвертьфинал",
};
const QP_COLOR: Record<number, string> = {
  1: "text-yellow-400",
  2: "text-[#4a9eff]",
  3: "text-[#4a9eff]",
};

function GroupSeedTable({ group, teamsById }: { group: BracketGroup; teamsById: Map<string, Team> }) {
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
              <th className="px-3 py-2 text-right">Статус</th>
            </tr>
          </thead>
          <tbody>
            {group.members.map((m) => {
              const team = teamsById.get(m.team_id);
              const qp = m.qualified_position ?? null;
              return (
                <tr key={m.id} className="border-b border-[#0a3575]/50">
                  <td className="px-3 py-2 text-white">{team?.name ?? m.team_id}</td>
                  <td className={`px-3 py-2 text-right font-semibold ${qp ? QP_COLOR[qp] : "text-[#4a7ab5]"}`}>
                    {qp ? QP_LABEL[qp] : "—"}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </CardContent>
    </Card>
  );
}

export function GroupDEView({
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

  const matchesByGroup = new Map<string, Match[]>();
  for (const m of matches) {
    if (!m.group_id) continue;
    if (!matchesByGroup.has(m.group_id)) matchesByGroup.set(m.group_id, []);
    matchesByGroup.get(m.group_id)!.push(m);
  }

  return (
    <div className="space-y-8">
      {/* Qualification status per group */}
      <div className="grid gap-4 md:grid-cols-2">
        {groups.map((g) => (
          <GroupSeedTable key={g.id} group={g} teamsById={teamsById} />
        ))}
      </div>

      {/* Per-group DE brackets */}
      {groups.map((g) => {
        const gMatches = matchesByGroup.get(g.id) ?? [];
        if (!gMatches.length) return null;
        return (
          <div key={g.id} className="space-y-2">
            <h3 className="text-xs font-semibold uppercase tracking-widest text-[#4a7ab5]">
              {g.name}
            </h3>
            <BracketView
              matches={gMatches}
              teams={teams}
              adminMode={adminMode}
              tournamentId={tournamentId}
            />
          </div>
        );
      })}
    </div>
  );
}
