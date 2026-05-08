import { useMemo, useState } from "react";
import { toast } from "sonner";
import { useAuth } from "@/app/providers/auth-provider";
import { CreateTournamentForm } from "@/features/tournaments/components/create-tournament-form";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { useCreateTournament, useTournaments } from "@/features/tournaments/hooks";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Input } from "@/shared/ui/input";
import { PageHeader } from "@/shared/ui/page-header";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import type { TournamentFormValues } from "@/features/tournaments/schemas";
import { getErrorMessage } from "@/shared/lib/http";

export function TournamentsListPage() {
  const { isAuthenticated } = useAuth();
  const [showCreate, setShowCreate] = useState(false);
  const [query, setQuery] = useState("");

  const tournamentsQuery = useTournaments();
  const createMutation = useCreateTournament();

  const filtered = useMemo(() => {
    const items = tournamentsQuery.data?.items ?? [];
    const normalized = query.trim().toLowerCase();
    if (!normalized) return items;
    return items.filter((item) => item.title.toLowerCase().includes(normalized));
  }, [query, tournamentsQuery.data?.items]);

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
        description="Публичный список турниров и создание нового турнира зарегистрированным пользователем."
        actions={
          isAuthenticated ? (
            <Button onClick={() => setShowCreate((value) => !value)}>
              {showCreate ? "Скрыть форму" : "Создать турнир"}
            </Button>
          ) : null
        }
      />

      <Card>
        <CardContent className="pt-5">
          <Input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Поиск по названию" />
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
      {!tournamentsQuery.isLoading && !tournamentsQuery.isError && !filtered.length ? (
        <EmptyState
          title="Ничего не найдено"
          description="Измените поисковый запрос или создайте новый турнир, если вы авторизованы."
        />
      ) : null}
      {filtered.length ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {filtered.map((tournament) => (
            <TournamentCard key={tournament.id} tournament={tournament} />
          ))}
        </div>
      ) : null}
    </div>
  );
}