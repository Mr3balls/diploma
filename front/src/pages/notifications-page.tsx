import { toast } from "sonner";
import { Bell, CheckCheck } from "lucide-react";
import {
  useNotificationAction,
  useNotificationStream,
  useNotifications,
  useReadAllNotifications,
  useMarkNotificationRead,
} from "@/features/notifications/hooks";
import { NotificationsList } from "@/features/notifications/components/notifications-list";
import { useAcceptParticipation, useDeclineParticipation } from "@/features/teams/hooks";
import { useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/shared/lib/query-keys";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
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

  useNotificationStream();

  const items = (notificationsQuery.data?.pages ?? []).flatMap((p) => p.items);
  const unreadCount = items.filter((n) => !n.is_read).length;
  const hasMore = notificationsQuery.hasNextPage;

  async function handleMarkRead(id: string) {
    try { await markReadMutation.mutateAsync(id); }
    catch (error) { toast.error(getErrorMessage(error)); }
  }

  async function handleReadAll() {
    try {
      await readAllMutation.mutateAsync();
      toast.success("Все уведомления прочитаны");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  function getMemberId(notification: Notification) {
    const pj = notification.payload_json;
    if (pj && typeof pj.team_member_id === "string") return pj.team_member_id;
    return null;
  }

  async function handleAccept(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) { toast.error("В уведомлении нет team_member_id"); return; }
    try {
      await acceptMutation.mutateAsync(memberId);
      if (!notification.is_read) await markReadMutation.mutateAsync(notification.id);
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
      toast.success("Участие подтверждено");
    } catch (error) {
      toast.error(getErrorMessage(error));
    }
  }

  async function handleDecline(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) { toast.error("В уведомлении нет team_member_id"); return; }
    try {
      await declineMutation.mutateAsync(memberId);
      if (!notification.is_read) await markReadMutation.mutateAsync(notification.id);
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
      toast.success("Действие выполнено");
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
          <Button size="sm" onClick={() => void handleAccept(notification)}
            disabled={acceptMutation.isPending}>
            Принять
          </Button>
          <Button size="sm" variant="destructive" onClick={() => void handleDecline(notification)}
            disabled={declineMutation.isPending}>
            Отклонить
          </Button>
        </>
      );
    }
    if (notification.action_payload_json?.action_available === true) {
      return (
        <Button size="sm" variant="outline" onClick={() => void handleGenericAction(notification)}>
          Выполнить действие
        </Button>
      );
    }
    return null;
  }

  return (
    <div className="grid gap-0">

      {/* ── Banner ───────────────────────────────────────────────── */}
      <div
        style={{
          width: "100vw",
          marginLeft: "calc(50% - 50vw)",
          background: "#111111",
          borderBottom: "1px solid #2d2d2d",
        }}
      >
        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#ff5500]/10">
                <Bell className="h-5 w-5 text-[#ff5500]" />
              </div>
              <div>
                <h1
                  className="font-black uppercase text-white"
                  style={{ fontSize: "clamp(1.5rem, 4vw, 2.5rem)", letterSpacing: "-0.03em" }}
                >
                  Уведомления
                </h1>
                <p className="text-sm text-[#666666]">
                  {unreadCount > 0 ? `${unreadCount} непрочитанных` : "Всё прочитано"}
                </p>
              </div>
            </div>

            {unreadCount > 0 && (
              <Button
                variant="outline"
                className="gap-2"
                disabled={readAllMutation.isPending}
                onClick={() => void handleReadAll()}
              >
                <CheckCheck className="h-4 w-4" />
                Прочитать все
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* ── Content ──────────────────────────────────────────────── */}
      <div className="py-8 grid gap-4">
        {notificationsQuery.isLoading ? <Spinner /> : null}
        {notificationsQuery.isError ? <ErrorState /> : null}
        {!notificationsQuery.isLoading && !notificationsQuery.isError && !items.length ? (
          <EmptyState title="Уведомлений нет" description="Новые уведомления появятся здесь." />
        ) : null}
        {items.length ? (
          <NotificationsList
            items={items}
            onMarkRead={(id) => void handleMarkRead(id)}
            renderActions={renderActions}
          />
        ) : null}
        {hasMore && (
          <div className="flex justify-center">
            <Button
              variant="outline"
              disabled={notificationsQuery.isFetchingNextPage}
              onClick={() => void notificationsQuery.fetchNextPage()}
            >
              {notificationsQuery.isFetchingNextPage ? "Загрузка..." : "Загрузить ещё"}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
