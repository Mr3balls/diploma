import { useState } from "react";
import { Trophy } from "lucide-react";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import type { ChallongeMatch, Participant } from "../types";
import { useSubmitResult, useResetMatch } from "../hooks";

type Round = {
  section: string;
  roundNumber: number;
  matches: ChallongeMatch[];
};

function buildRounds(matches: ChallongeMatch[]): Round[] {
  const map = new Map<string, Round>();
  for (const m of matches) {
    const section = m.bracket_section ?? "WB";
    const round = m.round_number ?? 1;
    const key = `${section}-${round}`;
    if (!map.has(key)) map.set(key, { section, roundNumber: round, matches: [] });
    map.get(key)!.matches.push(m);
  }
  const sectionOrder: Record<string, number> = { WB: 0, LB: 1, GF: 2 };
  return [...map.values()].sort((a, b) => {
    const so = (sectionOrder[a.section] ?? 9) - (sectionOrder[b.section] ?? 9);
    return so !== 0 ? so : a.roundNumber - b.roundNumber;
  });
}

function statusLabel(s: string) {
  const map: Record<string, string> = {
    scheduled: "Ожидание",
    in_progress: "Идёт",
    finished: "Завершён",
    cancelled: "Отменён",
  };
  return map[s] ?? s;
}

function participantName(id: string | null | undefined, byId: Map<string, Participant>) {
  if (!id) return "BYE";
  return byId.get(id)?.name ?? "—";
}

function MatchCard({
  match,
  byId,
  isManager,
  slug,
}: {
  match: ChallongeMatch;
  byId: Map<string, Participant>;
  isManager: boolean;
  slug: string;
}) {
  const [picking, setPicking] = useState(false);
  const submitResult = useSubmitResult(slug);
  const resetMatch = useResetMatch(slug);

  const p1name = participantName(match.participant1_id, byId);
  const p2name = participantName(match.participant2_id, byId);
  const winner = match.winner_participant_id;

  const isFinished = match.status === "finished" || Boolean(winner);
  const canSubmit =
    isManager &&
    !match.is_bye &&
    !isFinished &&
    match.participant1_id &&
    match.participant2_id;

  const canReset = isManager && isFinished;

  function submit(winnerID: string) {
    submitResult.mutate(
      { matchID: match.id, winnerID },
      { onSuccess: () => setPicking(false) },
    );
  }

  return (
    <Card className="border-[#0a3575] bg-[#001a4a]">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-xs text-[#90afd4]">
            {match.bracket_section} · #{match.slot_index ?? match.global_number ?? "—"}
          </CardTitle>
          <div className="flex items-center gap-1">
            {match.is_bye && <Badge tone="warning">BYE</Badge>}
            <Badge tone={isFinished ? "success" : "muted"}>{statusLabel(match.status)}</Badge>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-1 pt-0">
        {/* Participant rows */}
        {[
          { id: match.participant1_id, name: p1name },
          { id: match.participant2_id, name: p2name },
        ].map(({ id, name }) => {
          const isWinner = id && winner === id;
          return (
            <div
              key={id ?? name}
              className={`flex items-center justify-between rounded-lg px-2 py-1 text-sm ${
                isWinner ? "bg-[#2255ff]/20 font-semibold text-white" : "text-[#90afd4]"
              }`}
            >
              <span className="truncate">{name}</span>
              {isWinner && <Trophy className="ml-2 h-3 w-3 shrink-0 text-yellow-400" />}
            </div>
          );
        })}

        {/* Score */}
        {match.score_text && (
          <p className="pt-1 text-center text-xs text-[#90afd4]">{match.score_text}</p>
        )}

        {/* Actions */}
        {canSubmit && !picking && (
          <Button size="sm" variant="secondary" className="mt-2 w-full" onClick={() => setPicking(true)}>
            Указать победителя
          </Button>
        )}

        {picking && (
          <div className="mt-2 space-y-1">
            <p className="text-xs text-[#90afd4]">Победитель:</p>
            {[
              { id: match.participant1_id!, name: p1name },
              { id: match.participant2_id!, name: p2name },
            ].map(({ id, name }) => (
              <Button
                key={id}
                size="sm"
                variant="outline"
                className="w-full justify-start"
                disabled={submitResult.isPending}
                onClick={() => submit(id)}
              >
                {name}
              </Button>
            ))}
            <Button
              size="sm"
              variant="ghost"
              className="w-full"
              onClick={() => setPicking(false)}
            >
              Отмена
            </Button>
          </div>
        )}

        {canReset && (
          <Button
            size="sm"
            variant="ghost"
            className="mt-1 w-full text-xs text-[#90afd4]"
            disabled={resetMatch.isPending}
            onClick={() => resetMatch.mutate(match.id)}
          >
            Сбросить результат
          </Button>
        )}
      </CardContent>
    </Card>
  );
}

export function ChallongeBracket({
  matches,
  participants,
  isManager,
  slug,
}: {
  matches: ChallongeMatch[];
  participants: Participant[];
  isManager: boolean;
  slug: string;
}) {
  const byId = new Map(participants.map((p) => [p.id, p]));
  const rounds = buildRounds(matches);

  if (!rounds.length) {
    return (
      <Card className="border-[#0a3575]">
        <CardContent className="py-8 text-sm text-[#90afd4]">
          Сетка ещё не сгенерирована — добавьте участников и нажмите «Начать».
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
            <h3 className="text-xs font-semibold uppercase tracking-wide text-[#90afd4]">
              {round.section !== "WB" ? `${round.section} · ` : ""}Раунд {round.roundNumber}
            </h3>
            {round.matches
              .sort((a, b) => (a.slot_index ?? 0) - (b.slot_index ?? 0))
              .map((m) => (
                <MatchCard key={m.id} match={m} byId={byId} isManager={isManager} slug={slug} />
              ))}
          </div>
        ))}
      </div>
    </div>
  );
}
