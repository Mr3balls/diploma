import { useEffect, useState } from "react";
import { Search, Plus, X } from "lucide-react";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { CreateTournamentForm } from "@/features/tournaments/components/create-tournament-form";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { useCreateTournament, useTournaments } from "@/features/tournaments/hooks";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { Card, CardContent } from "@/shared/ui/card";
import { tournamentStatusLabel } from "@/shared/lib/enums";
import type { TournamentStatus } from "@/shared/types/api";
import type { TournamentFormValues } from "@/features/tournaments/schemas";
import { getErrorMessage } from "@/shared/lib/http";
import { cn } from "@/shared/lib/cn";

const STATUS_PILLS: { value: TournamentStatus | ""; label: string }[] = [
  { value: "",                   label: "Все" },
  { value: "registration_open",  label: tournamentStatusLabel["registration_open"] },
  { value: "in_progress",        label: tournamentStatusLabel["in_progress"] },
  { value: "finished",           label: tournamentStatusLabel["finished"] },
  { value: "draft",              label: tournamentStatusLabel["draft"] },
  { value: "cancelled",          label: tournamentStatusLabel["cancelled"] },
];

export function TournamentsListPage() {
  const { isAuthenticated } = useAuth();
  const [showCreate, setShowCreate] = useState(false);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [status, setStatus] = useState<TournamentStatus | "">("");

  useEffect(() => {
    const id = setTimeout(() => setDebouncedQuery(query.trim()), 400);
    return () => clearTimeout(id);
  }, [query]);

  const tournamentsQuery = useTournaments({
    status: status || undefined,
    q: debouncedQuery || undefined,
  });
  const createMutation = useCreateTournament();
  const items = tournamentsQuery.data?.items ?? [];

  async function handleCreate(values: TournamentFormValues) {
    try {
      await createMutation.mutateAsync(values);
      toast.success("Турнир создан");
      setShowCreate(false);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-0">

      {/* ── Page header ─────────────────────────────────────────── */}
      <div
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          borderBottom: "1px solid #2d2d2d",
          background: "#111111",
        }}
      >
        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <p className="mb-1 text-xs font-bold uppercase tracking-[0.35em] text-[#ff5500]">
                Платформа
              </p>
              <h1
                className="font-black uppercase text-white"
                style={{ fontSize: "clamp(2rem, 5vw, 3.5rem)", letterSpacing: "-0.03em" }}
              >
                Турниры
              </h1>
            </div>
            {isAuthenticated && (
              <Button
                className="gap-2"
                onClick={() => setShowCreate((v) => !v)}
              >
                {showCreate ? (
                  <><X className="h-4 w-4" /> Закрыть</>
                ) : (
                  <><Plus className="h-4 w-4" /> Создать турнир</>
                )}
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="py-8 grid gap-6">

        {/* ── Create form ───────────────────────────────────────── */}
        {showCreate && (
          <Card>
            <CardContent className="pt-5">
              <CreateTournamentForm
                submitLabel="Создать"
                onSubmit={handleCreate}
                isSubmitting={createMutation.isPending}
              />
            </CardContent>
          </Card>
        )}

        {/* ── Search + filters ──────────────────────────────────── */}
        <div className="flex flex-col gap-3">
          {/* search */}
          <div className="relative">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[#666666]" />
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Поиск турнира..."
              className="h-11 w-full rounded-xl border border-[#2d2d2d] bg-[#1a1a1a] pl-9 pr-4 text-sm text-white outline-none placeholder:text-[#666666] focus:border-[#ff5500] transition-colors"
            />
            {query && (
              <button
                onClick={() => setQuery("")}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-[#666666] hover:text-white"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>

          {/* status pills */}
          <div className="flex flex-wrap gap-2">
            {STATUS_PILLS.map((pill) => (
              <button
                key={pill.value}
                onClick={() => setStatus(pill.value)}
                className={cn(
                  "rounded-full border px-4 py-1.5 text-xs font-semibold uppercase tracking-wide transition-all",
                  status === pill.value
                    ? "border-[#ff5500] bg-[#ff5500] text-white"
                    : "border-[#2d2d2d] bg-transparent text-[#9e9e9e] hover:border-[#666666] hover:text-white",
                )}
              >
                {pill.label}
              </button>
            ))}
          </div>
        </div>

        {/* ── Results count ─────────────────────────────────────── */}
        {!tournamentsQuery.isLoading && tournamentsQuery.data && (
          <p className="text-xs text-[#666666]">
            {tournamentsQuery.data.total > 0
              ? `Найдено: ${tournamentsQuery.data.total} турниров`
              : "Ничего не найдено"}
          </p>
        )}

        {/* ── Content ───────────────────────────────────────────── */}
        {tournamentsQuery.isLoading ? <Spinner /> : null}
        {tournamentsQuery.isError ? <ErrorState /> : null}

        {!tournamentsQuery.isLoading && !tournamentsQuery.isError && !items.length ? (
          <EmptyState
            title="Турниров нет"
            description="По заданным фильтрам ничего не найдено."
          />
        ) : null}

        {items.length ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {items.map((t) => (
              <TournamentCard key={t.id} tournament={t} />
            ))}
          </div>
        ) : null}
      </div>
    </div>
  );
}
