import type { Match, Team } from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";

export type BracketRound = {
  roundNumber: number;
  section: string;
  matches: Match[];
};

export function buildRounds(matches: Match[]) {
  const groups = new Map<string, Match[]>();

  for (const match of matches) {
    const section = match.bracket_section ?? "WB";
    const round = Number(match.round_number ?? 1);
    const key = `${section}-${round}`;
    const existing = groups.get(key) ?? [];
    existing.push(match);
    groups.set(key, existing);
  }

  const sectionOrder: Record<string, number> = { WB: 0, LB: 1, GF: 2 };

  return Array.from(groups.entries())
    .sort((a, b) => {
      const [sA, rA] = a[0].split("-");
      const [sB, rB] = b[0].split("-");
      const sectionDiff = (sectionOrder[sA] ?? 99) - (sectionOrder[sB] ?? 99);
      if (sectionDiff !== 0) return sectionDiff;
      return Number(rA) - Number(rB);
    })
    .map<BracketRound>(([key, items]) => {
      const [section, roundStr] = key.split("-");
      return {
        roundNumber: Number(roundStr),
        section,
        matches: [...items].sort((a, b) => Number(a.slot_index ?? 0) - Number(b.slot_index ?? 0)),
      };
    });
}

export function deriveSeedOrderFromMatches(matches: Match[]) {
  const firstRound = matches
    .filter((match) => (match.bracket_section ?? "WB") === "WB" && Number(match.round_number ?? 1) === 1)
    .sort((a, b) => Number(a.slot_index ?? 0) - Number(b.slot_index ?? 0));

  const order: string[] = [];
  for (const match of firstRound) {
    const first = match.team1_id;
    const second = match.team2_id;
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

export function pickTeamId(match: Match, side: "team1" | "team2") {
  return side === "team1" ? match.team1_id ?? undefined : match.team2_id ?? undefined;
}

export function pickTeamName(match: Match, side: "team1" | "team2", teamsById?: Map<string, Team>) {
  const teamId = side === "team1" ? match.team1_id : match.team2_id;
  if (teamId && teamsById) {
    const team = teamsById.get(teamId);
    if (team?.name) return team.name;
  }
  if (teamId) return teamId.slice(0, 8) + "…";
  return "BYE";
}

export function buildTeamsById(teams: Team[]): Map<string, Team> {
  return new Map(teams.map((t) => [t.id, t]));
}

export function pickSideName(
  match: Match,
  side: "1" | "2",
  teamsById: Map<string, Team>,
  participantsById: Map<string, Participant>,
): string {
  const teamId = side === "1" ? match.team1_id : match.team2_id;
  if (teamId) {
    const team = teamsById.get(teamId);
    if (team?.name) return team.name;
    return teamId.slice(0, 8) + "…";
  }
  const pId = side === "1" ? match.participant1_id : match.participant2_id;
  if (pId) {
    const p = participantsById.get(pId);
    if (p?.name) return p.name;
    return pId.slice(0, 8) + "…";
  }
  return "BYE";
}

export function isMatchWinner(match: Match, side: "1" | "2"): boolean {
  const teamId = side === "1" ? match.team1_id : match.team2_id;
  const participantId = side === "1" ? match.participant1_id : match.participant2_id;
  return Boolean(
    (match.winner_team_id && match.winner_team_id === teamId) ||
    (match.winner_participant_id && match.winner_participant_id === participantId),
  );
}
