import type { Match, Team } from "@/shared/types/api";

export type BracketRound = {
  roundNumber: number;
  matches: Match[];
};

export function buildRounds(matches: Match[]) {
  const groups = new Map<number, Match[]>();

  for (const match of matches) {
    const round = Number(match.round_number ?? 1);
    const existing = groups.get(round) ?? [];
    existing.push(match);
    groups.set(round, existing);
  }

  return Array.from(groups.entries())
    .sort((a, b) => a[0] - b[0])
    .map<BracketRound>(([roundNumber, items]) => ({
      roundNumber,
      matches: [...items].sort((a, b) => Number(a.slot_index ?? 0) - Number(b.slot_index ?? 0)),
    }));
}

export function deriveSeedOrderFromMatches(matches: Match[]) {
  const firstRound = matches
    .filter((match) => Number(match.round_number ?? 1) === 1)
    .sort((a, b) => Number(a.slot_index ?? 0) - Number(b.slot_index ?? 0));

  const order: string[] = [];
  for (const match of firstRound) {
    const first = pickTeamId(match, "home");
    const second = pickTeamId(match, "away");
    if (first) order.push(first);
    if (second) order.push(second);
  }

  return Array.from(new Set(order));
}

export function deriveSeedOrderFromTeams(teams: Team[]) {
  return teams
    .filter((team) => team.status === "approved" || team.status === "ready_for_review")
    .map((team) => team.id);
}

export function pickTeamId(match: Match, side: "home" | "away") {
  const slot = side === "home" ? match.home_team_id : match.away_team_id;
  if (slot) return slot;
  if (side === "home" && match.home_team?.id) return match.home_team.id;
  if (side === "away" && match.away_team?.id) return match.away_team.id;
  return undefined;
}

export function pickTeamName(match: Match, side: "home" | "away") {
  const explicitName = side === "home" ? match.home_team_name : match.away_team_name;
  if (explicitName) return explicitName;

  const nestedName = side === "home" ? match.home_team?.name : match.away_team?.name;
  if (nestedName) return nestedName;

  return "BYE";
}