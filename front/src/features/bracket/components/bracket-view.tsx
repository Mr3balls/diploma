import { useState } from "react";
import { Trophy } from "lucide-react";
import { toast } from "sonner";
import type { Match, Team } from "@/shared/types/api";
import type { Participant } from "@/features/challonge/types";
import { buildTeamsById, pickSideName, isMatchWinner } from "@/shared/lib/bracket";
import { useAdminSetResult } from "@/features/matches/hooks";
import { getErrorMessage } from "@/shared/lib/http";
import { Button } from "@/shared/ui/button";
import { formatDateTime } from "@/shared/lib/date";

// ── Layout constants (px) ────────────────────────────────────────────────────
// CARD_H must be >= the tallest rendered card. Admin mode adds a ~44px button
// row; scheduled_at adds ~28px. Use 140 to cover all combinations safely.
const CARD_H = 140;
const V_GAP  = 20;
const CARD_W = 220;
const H_GAP  = 48;

// ── Round label ───────────────────────────────────────────────────────────────
function roundLabel(section: string, roundIdx: number, totalRounds: number) {
  if (section === "GF") {
    return roundIdx === 0 ? "Гранд-финал" : "Гранд-финал · реванш";
  }
  const prefix = section === "LB" ? "LB · " : "";
  const fromEnd = totalRounds - 1 - roundIdx;
  if (section === "WB") {
    if (fromEnd === 0 && totalRounds > 1) return "Финал";
    if (fromEnd === 1 && totalRounds > 2) return "Полуфинал";
    if (fromEnd === 2 && totalRounds > 3) return "Четвертьфинал";
  }
  return `${prefix}Раунд ${roundIdx + 1}`;
}

// ── SVG connector paths ───────────────────────────────────────────────────────
function buildConnectorPaths(
  prevCount: number,
  nextCount: number,
  prevSlotH: number,
  nextSlotH: number,
  prevColLeft: number,
  nextColLeft: number,
): string {
  const x0   = prevColLeft + CARD_W;
  const x1   = nextColLeft;
  const xMid = (x0 + x1) / 2;
  const parts: string[] = [];

  const prevMidY = (i: number) => i * prevSlotH + (prevSlotH - CARD_H) / 2 + CARD_H / 2;
  const nextMidY = (i: number) => i * nextSlotH + (nextSlotH - CARD_H) / 2 + CARD_H / 2;

  if (prevCount === nextCount) {
    // 1:1 — LB major rounds (each LB winner faces a fresh WB dropout)
    for (let i = 0; i < nextCount; i++) {
      parts.push(`M ${x0} ${prevMidY(i)} H ${x1}`);
    }
  } else {
    // 2:1 merge — WB rounds and LB minor rounds
    for (let ni = 0; ni < nextCount; ni++) {
      const c1i = ni * 2;
      const c2i = ni * 2 + 1;
      const nmY = nextMidY(ni);
      const hasC1 = c1i < prevCount;
      const hasC2 = c2i < prevCount;
      if (hasC1 && hasC2) {
        const c1Y = prevMidY(c1i);
        const c2Y = prevMidY(c2i);
        parts.push(
          `M ${x0} ${c1Y} H ${xMid}`,
          `M ${xMid} ${c1Y} V ${c2Y}`,
          `M ${x0} ${c2Y} H ${xMid}`,
          `M ${xMid} ${nmY} H ${x1}`,
        );
      } else if (hasC1) {
        parts.push(`M ${x0} ${prevMidY(c1i)} H ${x1}`);
      } else if (hasC2) {
        parts.push(`M ${x0} ${prevMidY(c2i)} H ${x1}`);
      }
    }
  }

  return parts.join(" ");
}

// ── Match card ────────────────────────────────────────────────────────────────
// The picking overlay uses `absolute inset-0` so it covers the card without
// changing its height — avoiding layout shifts that cause card overlap.
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

  const p1 = pickSideName(match, "1", teamsById, participantsById);
  const p2 = pickSideName(match, "2", teamsById, participantsById);
  // Don't show winner styling on is_bye matches — they may be stale from a
  // backend auto-advance bug; the admin should reset them.
  const w1 = !match.is_bye && isMatchWinner(match, "1");
  const w2 = !match.is_bye && isMatchWinner(match, "2");
  const done = match.status === "finished";

  const isParticipantMatch = Boolean(match.participant1_id || match.participant2_id);
  const canPick =
    adminMode && !done && !match.is_bye && Boolean(tournamentId) &&
    (isParticipantMatch
      ? Boolean(match.participant1_id) && Boolean(match.participant2_id)
      : Boolean(match.team1_id) && Boolean(match.team2_id));

  function submit(id: string) {
    const payload = isParticipantMatch
      ? { winner_participant_id: id }
      : { winner_team_id: id };
    adminSetResult.mutate(
      { matchId: match.id, payload },
      {
        onSuccess: () => setPicking(false),
        onError: (err) => toast.error(getErrorMessage(err)),
      },
    );
  }

  return (
    <div className="relative overflow-hidden rounded-lg border border-[#0a3575] bg-[#001a4a] text-sm">
      {match.scheduled_at && (
        <div className="border-b border-[#0a3575] px-2 py-1 text-[10px] text-[#4a7ab5]">
          {formatDateTime(match.scheduled_at)}
        </div>
      )}

      {([["1", p1, w1] as const, ["2", p2, w2] as const]).map(([side, name, won]) => (
        <div
          key={side}
          className={`flex items-center justify-between gap-1 px-2 py-1.5 ${
            won ? "bg-[#2255ff]/20 font-semibold text-white" : "text-[#90afd4]"
          }`}
        >
          <span className="min-w-0 flex-1 truncate">{name}</span>
          {won && <Trophy className="h-3 w-3 shrink-0 text-yellow-400" />}
        </div>
      ))}

      {match.score_text && (
        <div className="border-t border-[#0a3575] px-2 py-0.5 text-center text-[10px] text-[#90afd4]">
          {match.score_text}
        </div>
      )}

      {canPick && (
        <div className="border-t border-[#0a3575] px-2 py-1.5">
          <Button size="sm" variant="secondary" className="w-full text-xs" onClick={() => setPicking(true)}>
            Указать победителя
          </Button>
        </div>
      )}

      {/* Overlay picker — absolute inset-0 keeps card height fixed */}
      {picking && (
        <div className="absolute inset-0 flex flex-col gap-1 bg-[#001a4a] p-2">
          {isParticipantMatch ? (
            <>
              {match.participant1_id && (
                <Button size="sm" variant="outline" className="w-full flex-1 justify-start text-xs"
                  disabled={adminSetResult.isPending} onClick={() => submit(match.participant1_id!)}>
                  {p1}
                </Button>
              )}
              {match.participant2_id && (
                <Button size="sm" variant="outline" className="w-full flex-1 justify-start text-xs"
                  disabled={adminSetResult.isPending} onClick={() => submit(match.participant2_id!)}>
                  {p2}
                </Button>
              )}
            </>
          ) : (
            <>
              {match.team1_id && (
                <Button size="sm" variant="outline" className="w-full flex-1 justify-start text-xs"
                  disabled={adminSetResult.isPending} onClick={() => submit(match.team1_id!)}>
                  {p1}
                </Button>
              )}
              {match.team2_id && (
                <Button size="sm" variant="outline" className="w-full flex-1 justify-start text-xs"
                  disabled={adminSetResult.isPending} onClick={() => submit(match.team2_id!)}>
                  {p2}
                </Button>
              )}
            </>
          )}
          <Button size="sm" variant="ghost" className="w-full text-xs" onClick={() => setPicking(false)}>
            Отмена
          </Button>
        </div>
      )}
    </div>
  );
}

// ── One section (WB / LB / GF) ───────────────────────────────────────────────
function BracketSection({
  section,
  matches,
  teamsById,
  participantsById,
  adminMode,
  tournamentId,
}: {
  section: string;
  matches: Match[];
  teamsById: Map<string, Team>;
  participantsById: Map<string, Participant>;
  adminMode: boolean;
  tournamentId?: string;
}) {
  const byRound = new Map<number, Match[]>();
  for (const m of matches) {
    const r = m.round_number ?? 1;
    if (!byRound.has(r)) byRound.set(r, []);
    byRound.get(r)!.push(m);
  }
  const rounds = Array.from(byRound.entries())
    .sort(([a], [b]) => a - b)
    .map(([, ms]) => [...ms].sort((a, b) => (a.slot_index ?? 0) - (b.slot_index ?? 0)));

  if (!rounds.length) return null;

  // Data-driven layout: totalH set by densest round; each round's slotH =
  // totalH / matchCount so all rounds share the same vertical extent.
  const matchCounts = rounds.map((r) => r.length);
  const maxCount    = Math.max(...matchCounts);
  const totalH      = maxCount * (CARD_H + V_GAP);
  const totalW      = rounds.length * (CARD_W + H_GAP) - H_GAP;

  const slotHs = matchCounts.map((c) => totalH / c);

  const cardTop = (ri: number, mi: number) => {
    const sh = slotHs[ri];
    return mi * sh + (sh - CARD_H) / 2;
  };

  const colLeft = (ri: number) => ri * (CARD_W + H_GAP);

  const connectorPaths: string[] = [];
  for (let ri = 0; ri < rounds.length - 1; ri++) {
    const p = buildConnectorPaths(
      matchCounts[ri],
      matchCounts[ri + 1],
      slotHs[ri],
      slotHs[ri + 1],
      colLeft(ri),
      colLeft(ri + 1),
    );
    if (p) connectorPaths.push(p);
  }

  const sectionLabel =
    section === "LB" ? "Lower Bracket" : "Grand Final";

  return (
    <div className="space-y-3">
      {section !== "WB" && (
        <h3 className="text-xs font-semibold uppercase tracking-widest text-[#4a7ab5]">
          {sectionLabel}
        </h3>
      )}

      <div className="relative" style={{ width: totalW, height: totalH }}>
        <svg
          className="pointer-events-none absolute inset-0"
          width={totalW}
          height={totalH}
          overflow="visible"
        >
          {connectorPaths.map((d, i) => (
            <path key={i} d={d} fill="none" stroke="#1a4a7a" strokeWidth={1.5} />
          ))}
        </svg>

        {rounds.map((roundMatches, ri) =>
          roundMatches.map((match, mi) => (
            <div
              key={match.id}
              className="absolute"
              style={{ left: colLeft(ri), top: cardTop(ri, mi), width: CARD_W }}
            >
              <MatchCard
                match={match}
                teamsById={teamsById}
                participantsById={participantsById}
                adminMode={adminMode}
                tournamentId={tournamentId}
              />
            </div>
          )),
        )}
      </div>

      <div className="relative" style={{ width: totalW }}>
        {rounds.map((_, ri) => (
          <div
            key={ri}
            className="absolute text-center text-[10px] font-semibold uppercase tracking-wide text-[#4a7ab5]"
            style={{ left: colLeft(ri), width: CARD_W }}
          >
            {roundLabel(section, ri, rounds.length)}
          </div>
        ))}
      </div>
    </div>
  );
}

// ── Public export ─────────────────────────────────────────────────────────────
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
  // Only WB seeding byes (empty bracket slots) should be hidden.
  // LB and GF matches are never genuine byes — always show them so the full
  // bracket structure is visible even before opponents drop in.
  const visible = matches.filter(
    (m) => !m.is_bye || (m.bracket_section !== "WB" && m.bracket_section != null),
  );

  const teamsById = buildTeamsById(teams);
  const participantsById = new Map(participants.map((p) => [p.id, p]));

  if (!visible.length) {
    return (
      <div className="rounded-xl border border-[#0a3575] px-6 py-8 text-sm text-[#90afd4]">
        Сетка пока не создана.
      </div>
    );
  }

  const sections = ["WB", "LB", "GF"] as const;
  const bySectionMatches = new Map<string, Match[]>();
  for (const m of visible) {
    const s = m.bracket_section ?? "WB";
    if (!bySectionMatches.has(s)) bySectionMatches.set(s, []);
    bySectionMatches.get(s)!.push(m);
  }

  return (
    <div className="overflow-x-auto pb-4">
      <div className="inline-flex min-w-max flex-col gap-10">
        {sections.map((s) => {
          const ms = bySectionMatches.get(s);
          if (!ms?.length) return null;
          return (
            <BracketSection
              key={s}
              section={s}
              matches={ms}
              teamsById={teamsById}
              participantsById={participantsById}
              adminMode={adminMode}
              tournamentId={tournamentId}
            />
          );
        })}
      </div>
    </div>
  );
}
