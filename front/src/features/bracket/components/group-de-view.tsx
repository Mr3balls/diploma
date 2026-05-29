import type { BracketGroup, Match, Team } from "@/shared/types/api";
import { BracketView } from "./bracket-view";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { useLang } from "@/app/providers/lang-provider";

const QP_COLOR: Record<number, string> = {
  1: "text-yellow-400",
  2: "text-[#ff7733]",
  3: "text-[#ff7733]",
};

function GroupSeedTable({ group, teamsById }: { group: BracketGroup; teamsById: Map<string, Team> }) {
  const { t } = useLang();

  const qpLabel: Record<number, string> = {
    1: t("bracket.toSemiFinal"),
    2: t("bracket.toQuarterFinal"),
    3: t("bracket.toQuarterFinal"),
  };

  return (
    <Card className="border-[#2d2d2d] bg-[#1a1a1a]">
      <CardHeader className="pb-2">
        <CardTitle className="text-sm text-white">{group.name}</CardTitle>
      </CardHeader>
      <CardContent className="p-0">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-[#2d2d2d] text-[#9e9e9e]">
              <th className="px-3 py-2 text-left">{t("teamsTable.name")}</th>
              <th className="px-3 py-2 text-right">{t("teamsTable.status")}</th>
            </tr>
          </thead>
          <tbody>
            {group.members.map((m) => {
              const team = teamsById.get(m.team_id);
              const qp = m.qualified_position ?? null;
              return (
                <tr key={m.id} className="border-b border-[#2d2d2d]/50">
                  <td className="px-3 py-2 text-white">{team?.name ?? m.team_id}</td>
                  <td className={`px-3 py-2 text-right font-semibold ${qp ? QP_COLOR[qp] : "text-[#666666]"}`}>
                    {qp ? qpLabel[qp] : "—"}
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
            <h3 className="text-xs font-semibold uppercase tracking-widest text-[#666666]">
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
