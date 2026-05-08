import { toast } from "sonner";
import { useNotificationAction, useNotifications, useReadAllNotifications, useMarkNotificationRead } from "@/features/notifications/hooks";
import { NotificationsList } from "@/features/notifications/components/notifications-list";
import { useAcceptParticipation, useDeclineParticipation } from "@/features/teams/hooks";
import { useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/shared/lib/query-keys";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { PageHeader } from "@/shared/ui/page-header";
import { Spinner } from "@/shared/ui/spinner";
import type { Notification } from "@/shared/types/api";
import { getErrorMessage } from "@/shared/lib/http";

export function NotificationsPage() {
  const queryClient = useQueryClient();
  const notificationsQuery = useNotifications();
  const markReadMutation = useMarkNotificationRead();
  const readAllMutation = useReadAllNotifications();
  const actionMutation = useNotificationAction();
  const acceptMutation = useAcceptParticipation();
  const declineMutation = useDeclineParticipation();

  async function handleMarkRead(id: string) {
    try {
      await markReadMutation.mutateAsync(id);
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleReadAll() {
    try {
      await readAllMutation.mutateAsync();
      toast.success("Все уведомления отмечены как прочитанные");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  function getMemberId(notification: Notification) {
    return (
      notification.team_member_id ||
      (typeof notification.payload?.team_member_id === "string" ? notification.payload.team_member_id : null)
    );
  }

  async function handleAccept(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) {
      toast.error("В уведомлении нет team_member_id");
      return;
    }

    try {
      await acceptMutation.mutateAsync(memberId);
      if (!notification.is_read) {
        await markReadMutation.mutateAsync(notification.id);
      }
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
      toast.success("Участие подтверждено");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleDecline(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) {
      toast.error("В уведомлении нет team_member_id");
      return;
    }

    try {
      await declineMutation.mutateAsync(memberId);
      if (!notification.is_read) {
        await markReadMutation.mutateAsync(notification.id);
      }
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
      toast.success("Участие отклонено");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleGenericAction(notification: Notification) {
    try {
      await actionMutation.mutateAsync({ id: notification.id });
      toast.success("Действие отправлено");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  function renderActions(notification: Notification) {
    if (notification.type === "added_to_team") {
      const memberId = getMemberId(notification);
      if (!memberId) return null;
      return (
        <>
          <Button size="sm" onClick={() => void handleAccept(notification)}>
            Принять участие
          </Button>
          <Button size="sm" variant="destructive" onClick={() => void handleDecline(notification)}>
            Отклонить
          </Button>
        </>
      );
    }

    if (notification.payload?.action_available === true) {
      return (
        <Button size="sm" variant="outline" onClick={() => void handleGenericAction(notification)}>
          Выполнить действие
        </Button>
      );
    }

    return null;
  }

  return (
    <div className="grid gap-6">
      <PageHeader
        title="Уведомления"
        description="Список уведомлений и inline-действия только там, где backend реально поддерживает сценарий."
        actions={
          <Button variant="outline" onClick={() => void handleReadAll()} disabled={readAllMutation.isPending}>
            Прочитать все
          </Button>
        }
      />

      {notificationsQuery.isLoading ? <Spinner /> : null}
      {notificationsQuery.isError ? <ErrorState /> : null}
      {!notificationsQuery.isLoading && !notificationsQuery.isError && !notificationsQuery.data?.items.length ? (
        <EmptyState title="Уведомлений нет" description="Новые уведомления появятся здесь." />
      ) : null}
      {notificationsQuery.data?.items.length ? (
        <NotificationsList
          items={notificationsQuery.data.items}
          onMarkRead={(id) => void handleMarkRead(id)}
          renderActions={renderActions}
        />
      ) : null}
    </div>
  );
}