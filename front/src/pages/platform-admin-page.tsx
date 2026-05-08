import { toast } from "sonner";
import { useAdminTournaments, useAdminUsers, useBlockUser, useUnblockUser } from "@/features/admin/hooks";
import { PlatformUsersTable } from "@/features/admin/components/platform-users-table";
import { TournamentCard } from "@/features/tournaments/components/tournament-card";
import { PageHeader } from "@/shared/ui/page-header";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import { SectionCard } from "@/shared/ui/section";
import { getErrorMessage } from "@/shared/lib/http";

export function PlatformAdminPage() {
  const usersQuery = useAdminUsers();
  const tournamentsQuery = useAdminTournaments();
  const blockMutation = useBlockUser();
  const unblockMutation = useUnblockUser();

  async function handleBlock(id: string) {
    try {
      await blockMutation.mutateAsync(id);
      toast.success("Пользователь заблокирован");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleUnblock(id: string) {
    try {
      await unblockMutation.mutateAsync(id);
      toast.success("Пользователь разблокирован");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  return (
    <div className="grid gap-6">
      <PageHeader title="Platform admin" description="Глобальное управление пользователями и турнирами платформы." />

      <SectionCard title="Пользователи" description="GET /admin/users + block/unblock">
        {usersQuery.isLoading ? (
          <Spinner />
        ) : usersQuery.isError ? (
          <ErrorState />
        ) : usersQuery.data?.items.length ? (
          <PlatformUsersTable users={usersQuery.data.items} onBlock={handleBlock} onUnblock={handleUnblock} />
        ) : (
          <EmptyState title="Пользователей нет" description="Список пользователей пуст." />
        )}
      </SectionCard>

      <SectionCard title="Турниры платформы" description="GET /admin/tournaments">
        {tournamentsQuery.isLoading ? (
          <Spinner />
        ) : tournamentsQuery.isError ? (
          <ErrorState />
        ) : tournamentsQuery.data?.items.length ? (
          <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
            {tournamentsQuery.data.items.map((tournament) => (
              <TournamentCard key={tournament.id} tournament={tournament} />
            ))}
          </div>
        ) : (
          <EmptyState title="Турниров нет" description="Пока нет данных для отображения." />
        )}
      </SectionCard>
    </div>
  );
}