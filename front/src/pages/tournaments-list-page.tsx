import { useEffect, useState } from "react";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { CreateTournamentForm } from "@/features/tournaments/components/create-tournament-form";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { useCreateTournament, useTournaments } from "@/features/tournaments/hooks";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Input } from "@/shared/ui/input";
import { Select } from "@/shared/ui/select";
import { PageHeader } from "@/shared/ui/page-header";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { tournamentFormatLabel, tournamentStatusLabel } from "@/shared/lib/enums";
import type { TournamentFormat, TournamentStatus } from "@/shared/types/api";
import type { TournamentFormValues } from "@/features/tournaments/schemas";
import { getErrorMessage } from "@/shared/lib/http";

const STATUS_OPTIONS: { value: string; label: string }[] = [
  { value: "", label: "Все статусы" },
  ...(
    [
      "draft",
      "registration_open",
      "registration_closed",
      "bracket_generated",
      "in_progress",
      "finished",
      "cancelled",
    ] as TournamentStatus[]
  ).map((s) => ({ value: s, label: tournamentStatusLabel[s] })),
];

const FORMAT_OPTIONS: { value: string; label: string }[] = [
  { value: "", label: "Все форматы" },
  ...(["single_elimination", "double_elimination"] as TournamentFormat[]).map((f) => ({
    value: f,
    label: tournamentFormatLabel[f],
  })),
];

export function TournamentsListPage() {
  const { isAuthenticated } = useAuth();
  const [showCreate, setShowCreate] = useState(false);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [status, setStatus] = useState("");
  const [format, setFormat] = useState("");

  useEffect(() => {
    const id = setTimeout(() => setDebouncedQuery(query.trim()), 400);
    return () => clearTimeout(id);
  }, [query]);

  const tournamentsQuery = useTournaments({
    status: status || undefined,
    format: format || undefined,
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
    <div className="grid gap-6">
      <PageHeader
        title="Турниры"
        description="Список турниров платформы."
        actions={
          isAuthenticated ? (
            <Button onClick={() => setShowCreate((v) => !v)}>
              {showCreate ? "Скрыть" : "Создать турнир"}
            </Button>
          ) : null
        }
      />

      <Card>
        <CardContent className="pt-5">
          <div className="flex flex-col gap-3 sm:flex-row">
            <Input
              className="flex-1"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Поиск по названию"
            />
            <Select
              className="sm:w-52"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
            >
              {STATUS_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </Select>
            <Select
              className="sm:w-52"
              value={format}
              onChange={(e) => setFormat(e.target.value)}
            >
              {FORMAT_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </Select>
          </div>
        </CardContent>
      </Card>

      {showCreate ? (
        <Card>
          <CardContent className="pt-5">
            <CreateTournamentForm
              submitLabel="Создать"
              onSubmit={handleCreate}
              isSubmitting={createMutation.isPending}
            />
          </CardContent>
        </Card>
      ) : null}

      {tournamentsQuery.isLoading ? <Spinner /> : null}
      {tournamentsQuery.isError ? <ErrorState /> : null}
      {!tournamentsQuery.isLoading && !tournamentsQuery.isError && !items.length ? (
        <EmptyState
          title="Турниров нет"
          description="По заданным фильтрам турниры не найдены."
        />
      ) : null}
      {items.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {items.map((tournament) => (
            <TournamentCard key={tournament.id} tournament={tournament} />
          ))}
        </div>
      ) : null}
      {tournamentsQuery.data && tournamentsQuery.data.total > 0 ? (
        <p className="text-center text-sm text-muted-foreground">
          Показано {items.length} из {tournamentsQuery.data.total} турниров
        </p>
      ) : null}
    </div>
  );
}
