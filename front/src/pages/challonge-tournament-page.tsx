import { useRef, useState } from "react";
import { useParams } from "react-router-dom";
import {
  Plus,
  Shuffle,
  Play,
  RotateCcw,
  Trash2,
  ClipboardList,
  Trophy,
} from "lucide-react";
import { Badge } from "@/shared/ui/badge";
import { Button } from "@/shared/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Spinner } from "@/shared/ui/spinner";
import { ChallongeBracket } from "@/features/challonge/components/challonge-bracket";
import {
  useChallongeBracket,
  useChallongeStandings,
  useAddParticipant,
  useBulkAdd,
  useRemoveParticipant,
  useShuffle,
  useStartTournament,
  useResetTournament,
} from "@/features/challonge/hooks";

const STATUS_LABEL: Record<string, string> = {
  draft: "Черновик",
  ready: "Готов к старту",
  in_progress: "Идёт",
  completed: "Завершён",
};

const STATUS_TONE: Record<string, "muted" | "warning" | "success" | "default"> = {
  draft: "muted",
  ready: "default",
  in_progress: "warning",
  completed: "success",
};

export function ChallongeTournamentPage() {
  const { slug } = useParams<{ slug: string }>();
  const { data, isLoading, error } = useChallongeBracket(slug);

  const [singleName, setSingleName] = useState("");
  const [bulkText, setBulkText] = useState("");
  const [showBulk, setShowBulk] = useState(false);
  const singleInputRef = useRef<HTMLInputElement>(null);

  const addOne = useAddParticipant(slug!);
  const bulkAdd = useBulkAdd(slug!);
  const remove = useRemoveParticipant(slug!);
  const shuffle = useShuffle(slug!);
  const start = useStartTournament(slug!);
  const reset = useResetTournament(slug!);

  if (isLoading) {
    return (
      <div className="page-shell flex items-center justify-center py-24">
        <Spinner />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="page-shell py-16 text-center text-[#90afd4]">
        Турнир не найден или у вас нет доступа.
      </div>
    );
  }

  const { tournament, participants, matches, current_user_role } = data;
  const isManager =
    current_user_role === "organizer" || current_user_role === "co_organizer";
  const isDraftOrReady = tournament.status === "draft" || tournament.status === "ready";
  const isActive =
    tournament.status === "in_progress" || tournament.status === "completed";

  function handleAddOne(e: React.FormEvent) {
    e.preventDefault();
    const name = singleName.trim();
    if (!name) return;
    addOne.mutate(name, {
      onSuccess: () => {
        setSingleName("");
        singleInputRef.current?.focus();
      },
    });
  }

  function handleBulkAdd() {
    const names = bulkText
      .split("\n")
      .map((n) => n.trim())
      .filter(Boolean);
    if (!names.length) return;
    bulkAdd.mutate(names, {
      onSuccess: () => {
        setBulkText("");
        setShowBulk(false);
      },
    });
  }

  return (
    <div className="page-shell space-y-8 py-8">
      {/* Header */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">{tournament.title}</h1>
            <Badge tone={STATUS_TONE[tournament.status] ?? "muted"}>
              {STATUS_LABEL[tournament.status] ?? tournament.status}
            </Badge>
          </div>
          <p className="mt-1 text-sm text-[#90afd4]">
            {tournament.format === "double_elimination"
              ? "Double Elimination"
              : "Single Elimination"}{" "}
            · Слаг: <code className="text-white">{tournament.slug}</code>
          </p>
        </div>

        {isManager && isDraftOrReady && (
          <div className="flex gap-2">
            <Button
              variant="secondary"
              size="sm"
              disabled={shuffle.isPending || participants.length < 2}
              onClick={() => shuffle.mutate()}
            >
              <Shuffle className="mr-1 h-4 w-4" />
              Перемешать
            </Button>
            <Button
              size="sm"
              disabled={start.isPending || participants.length < 2}
              onClick={() => start.mutate()}
            >
              <Play className="mr-1 h-4 w-4" />
              Начать
            </Button>
          </div>
        )}

        {isManager && isActive && (
          <Button
            variant="outline"
            size="sm"
            disabled={reset.isPending}
            onClick={() => {
              if (confirm("Сбросить сетку и вернуться к черновику?")) reset.mutate();
            }}
          >
            <RotateCcw className="mr-1 h-4 w-4" />
            Сбросить турнир
          </Button>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
        {/* Left: Participant panel */}
        <div className="space-y-4">
          <Card className="border-[#0a3575] bg-[#001a4a]">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2 text-base">
                  <ClipboardList className="h-4 w-4" />
                  Участники ({participants.length}
                  {tournament.max_participants
                    ? ` / ${tournament.max_participants}`
                    : ""}
                  )
                </CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-3">
              {/* Add inputs - only in draft/ready */}
              {isManager && isDraftOrReady && (
                <>
                  <form onSubmit={handleAddOne} className="flex gap-2">
                    <Input
                      ref={singleInputRef}
                      placeholder="Имя участника"
                      value={singleName}
                      onChange={(e) => setSingleName(e.target.value)}
                      className="flex-1"
                    />
                    <Button type="submit" size="sm" disabled={addOne.isPending || !singleName.trim()}>
                      <Plus className="h-4 w-4" />
                    </Button>
                  </form>

                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full text-xs"
                    onClick={() => setShowBulk((v) => !v)}
                  >
                    {showBulk ? "Скрыть" : "Добавить списком (несколько сразу)"}
                  </Button>

                  {showBulk && (
                    <div className="space-y-2">
                      <textarea
                        className="w-full rounded-xl border border-[#0a3575] bg-[#001538] px-3 py-2 text-sm text-white placeholder-[#4a7ab5] focus:outline-none focus:ring-1 focus:ring-[#2255ff]"
                        rows={6}
                        placeholder={"Один участник на строку:\nАлексей\nМихаил\nСергей"}
                        value={bulkText}
                        onChange={(e) => setBulkText(e.target.value)}
                      />
                      <Button
                        size="sm"
                        className="w-full"
                        disabled={bulkAdd.isPending || !bulkText.trim()}
                        onClick={handleBulkAdd}
                      >
                        {bulkAdd.isPending ? "Добавление..." : "Добавить всех"}
                      </Button>
                    </div>
                  )}
                </>
              )}

              {/* Participant list */}
              {participants.length === 0 ? (
                <p className="py-4 text-center text-sm text-[#90afd4]">
                  Пока нет участников
                </p>
              ) : (
                <ul className="divide-y divide-[#0a3575]">
                  {[...participants]
                    .sort((a, b) => a.seed - b.seed)
                    .map((p) => (
                      <li key={p.id} className="flex items-center justify-between gap-2 py-2">
                        <div className="flex items-center gap-2 min-w-0">
                          <span className="w-6 shrink-0 text-right text-xs text-[#4a7ab5]">
                            {p.seed}
                          </span>
                          <span className="truncate text-sm text-white">{p.name}</span>
                          {p.status === "champion" && (
                            <Trophy className="h-3.5 w-3.5 shrink-0 text-yellow-400" />
                          )}
                          {p.status === "eliminated" && (
                            <span className="text-xs text-[#4a7ab5]">выбыл</span>
                          )}
                        </div>
                        {isManager && isDraftOrReady && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-6 w-6 shrink-0 p-0 text-[#4a7ab5] hover:text-red-400"
                            disabled={remove.isPending}
                            onClick={() => remove.mutate(p.id)}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </li>
                    ))}
                </ul>
              )}

              {isDraftOrReady && participants.length < 2 && (
                <p className="rounded-xl bg-yellow-900/20 px-3 py-2 text-xs text-yellow-400">
                  Добавьте минимум 2 участника для запуска
                </p>
              )}
            </CardContent>
          </Card>

          {/* Standings (completed only) */}
          {tournament.status === "completed" && (
            <StandingsCard slug={slug!} />
          )}
        </div>

        {/* Right: Bracket */}
        <div>
          {isActive ? (
            <ChallongeBracket
              matches={matches}
              participants={participants}
              isManager={isManager}
              slug={slug!}
            />
          ) : (
            <Card className="border-[#0a3575] bg-[#001a4a]">
              <CardContent className="py-12 text-center text-sm text-[#90afd4]">
                {participants.length < 2
                  ? "Добавьте участников и нажмите «Начать»"
                  : "Нажмите «Начать» для генерации сетки"}
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}

function StandingsCard({ slug }: { slug: string }) {
  const { data } = useChallongeStandings(slug);
  if (!data?.standings?.length) return null;

  return (
    <Card className="border-[#0a3575] bg-[#001a4a]">
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-base">
          <Trophy className="h-4 w-4 text-yellow-400" />
          Итоговое место
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ul className="divide-y divide-[#0a3575]">
          {data.standings.map((s) => (
            <li key={s.seed} className="flex items-center justify-between py-2 text-sm">
              <span className="text-white">
                {s.rank}. Сид #{s.seed}
                {s.tied ? " (tied)" : ""}
              </span>
              <span className="text-[#90afd4]">
                {s.wins}W / {s.losses}L
              </span>
            </li>
          ))}
        </ul>
      </CardContent>
    </Card>
  );
}
