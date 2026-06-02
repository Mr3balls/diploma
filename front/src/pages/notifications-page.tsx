import { toast } from "sonner";
import { Bell, CheckCheck, Trash2, Settings } from "lucide-react";
import { useState } from "react";
import {
  useNotificationAction,
  useNotificationStream,
  useNotifications,
  useReadAllNotifications,
  useMarkNotificationRead,
  useDeleteNotification,
  useDeleteAllNotifications,
  useNotificationPreferences,
  useSetNotificationPreferences,
  useRegisterPush,
} from "@/features/notifications/hooks";
import { NotificationsList } from "@/features/notifications/components/notifications-list";
import { useAcceptParticipation, useDeclineParticipation } from "@/features/teams/hooks";
import { useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/shared/lib/query-keys";
import { useLang } from "@/app/providers/lang-provider";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { ErrorState } from "@/shared/ui/error-state";
import { Spinner } from "@/shared/ui/spinner";
import type { Notification, NotificationType } from "@/shared/types/api";
import { getErrorMessage } from "@/shared/lib/http";

const ALL_TYPES: NotificationType[] = [
  "added_to_team",
  "team_participation_confirmed",
  "team_participation_declined",
  "match_assigned",
  "match_time_changed",
  "match_rescheduled",
  "match_cancelled",
  "result_submitted",
  "result_confirmed",
  "tournament_finished",
];

export function NotificationsPage() {
  const { t } = useLang();
  const queryClient = useQueryClient();
  const [showPrefs, setShowPrefs] = useState(false);

  const notificationsQuery = useNotifications();
  const markReadMutation = useMarkNotificationRead();
  const readAllMutation = useReadAllNotifications();
  const deleteMutation = useDeleteNotification();
  const deleteAllMutation = useDeleteAllNotifications();
  const actionMutation = useNotificationAction();
  const acceptMutation = useAcceptParticipation();
  const declineMutation = useDeclineParticipation();
  const prefsQuery = useNotificationPreferences();
  const setPrefsMutation = useSetNotificationPreferences();
  const registerPush = useRegisterPush();

  useNotificationStream();

  const items = (notificationsQuery.data?.pages ?? []).flatMap((p) => p.items);
  const unreadCount = items.filter((n) => !n.is_read).length;
  const hasMore = notificationsQuery.hasNextPage;
  const disabled: NotificationType[] = (prefsQuery.data?.disabled ?? []) as NotificationType[];

  function toggleType(type: NotificationType) {
    const next = disabled.includes(type)
      ? disabled.filter((d) => d !== type)
      : [...disabled, type];
    setPrefsMutation.mutate(next, {
      onError: (err) => toast.error(getErrorMessage(err)),
    });
  }

  async function handleMarkRead(id: string) {
    try { await markReadMutation.mutateAsync(id); }
    catch (err) { toast.error(getErrorMessage(err)); }
  }

  async function handleDelete(id: string) {
    try { await deleteMutation.mutateAsync(id); }
    catch (err) { toast.error(getErrorMessage(err)); }
  }

  async function handleDeleteAll() {
    try {
      await deleteAllMutation.mutateAsync();
      toast.success(t("notifications.deletedAll"));
    } catch (err) { toast.error(getErrorMessage(err)); }
  }

  async function handleReadAll() {
    try {
      await readAllMutation.mutateAsync();
      toast.success(t("notifications.allReadSuccess"));
    } catch (err) { toast.error(getErrorMessage(err)); }
  }

  function getMemberId(notification: Notification) {
    const pj = notification.payload_json;
    if (pj && typeof pj.team_member_id === "string") return pj.team_member_id;
    return null;
  }

  async function handleAccept(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) { toast.error(t("notifications.noMemberId")); return; }
    try {
      await acceptMutation.mutateAsync(memberId);
      if (!notification.is_read) await markReadMutation.mutateAsync(notification.id);
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
      toast.success(t("notifications.accepted"));
    } catch (err) { toast.error(getErrorMessage(err)); }
  }

  async function handleDecline(notification: Notification) {
    const memberId = getMemberId(notification);
    if (!memberId) { toast.error(t("notifications.noMemberId")); return; }
    try {
      await declineMutation.mutateAsync(memberId);
      if (!notification.is_read) await markReadMutation.mutateAsync(notification.id);
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
      toast.success(t("notifications.declined"));
    } catch (err) { toast.error(getErrorMessage(err)); }
  }

  async function handleGenericAction(notification: Notification) {
    try {
      await actionMutation.mutateAsync({ id: notification.id });
      toast.success(t("notifications.done"));
    } catch (err) { toast.error(getErrorMessage(err)); }
  }

  function renderActions(notification: Notification) {
    if (notification.type === "added_to_team") {
      const memberId = getMemberId(notification);
      if (!memberId) return null;
      return (
        <>
          <Button size="sm" onClick={() => void handleAccept(notification)} disabled={acceptMutation.isPending}>
            {t("notifications.accept")}
          </Button>
          <Button size="sm" variant="destructive" onClick={() => void handleDecline(notification)} disabled={declineMutation.isPending}>
            {t("notifications.decline")}
          </Button>
        </>
      );
    }
    if (notification.action_payload_json?.action_available === true) {
      return (
        <Button size="sm" variant="outline" onClick={() => void handleGenericAction(notification)}>
          {t("notifications.action")}
        </Button>
      );
    }
    return null;
  }

  return (
    <div className="grid gap-0">
      {/* Banner */}
      <div style={{ width: "100vw", marginLeft: "calc(50% - 50vw)", background: "#111111", borderBottom: "1px solid #2d2d2d" }}>
        <div className="mx-auto w-full max-w-7xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-[#ff5500]/10">
                <Bell className="h-5 w-5 text-[#ff5500]" />
              </div>
              <div>
                <h1 className="font-black uppercase text-white" style={{ fontSize: "clamp(1.5rem, 4vw, 2.5rem)", letterSpacing: "-0.03em" }}>
                  {t("notifications.title")}
                </h1>
                <p className="text-sm text-[#666666]">
                  {unreadCount > 0 ? t("notifications.unread", { n: unreadCount }) : t("notifications.allRead")}
                </p>
              </div>
            </div>

            <div className="flex flex-wrap gap-2">
              <Button variant="ghost" size="sm" className="gap-1.5 text-[#9e9e9e]" onClick={() => setShowPrefs((v) => !v)}>
                <Settings className="h-4 w-4" />
                {t("notifications.settings")}
              </Button>
              <Button variant="ghost" size="sm" className="gap-1.5 text-[#9e9e9e]"
                onClick={() => void registerPush().catch(() => null)}>
                {t("notifications.enablePush")}
              </Button>
              {unreadCount > 0 && (
                <Button variant="outline" size="sm" className="gap-1.5" disabled={readAllMutation.isPending} onClick={() => void handleReadAll()}>
                  <CheckCheck className="h-4 w-4" />
                  {t("notifications.markAllRead")}
                </Button>
              )}
              {items.length > 0 && (
                <Button variant="destructive" size="sm" className="gap-1.5" disabled={deleteAllMutation.isPending} onClick={() => void handleDeleteAll()}>
                  <Trash2 className="h-4 w-4" />
                  {t("notifications.clearAll")}
                </Button>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Preferences panel */}
      {showPrefs && (
        <div className="border-b border-[#2d2d2d] bg-[#111]">
          <div className="mx-auto w-full max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
            <p className="mb-3 text-sm font-semibold text-[#9e9e9e] uppercase tracking-wide">{t("notifications.prefsTitle")}</p>
            <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {ALL_TYPES.map((type) => {
                const isDisabled = disabled.includes(type);
                return (
                  <label key={type} className="flex cursor-pointer items-center gap-2 rounded-lg border border-[#2d2d2d] px-3 py-2 text-sm text-white transition-colors hover:border-[#ff5500]/40">
                    <input
                      type="checkbox"
                      checked={!isDisabled}
                      onChange={() => toggleType(type)}
                      className="accent-[#ff5500]"
                    />
                    {t(`matchStatus.${type}` as Parameters<typeof t>[0]) !== `matchStatus.${type}`
                      ? t(`matchStatus.${type}` as Parameters<typeof t>[0])
                      : type}
                  </label>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Content */}
      <div className="py-8 grid gap-4">
        {notificationsQuery.isLoading ? <Spinner /> : null}
        {notificationsQuery.isError ? <ErrorState /> : null}
        {!notificationsQuery.isLoading && !notificationsQuery.isError && !items.length ? (
          <EmptyState title={t("notifications.empty")} description={t("notifications.emptyDesc")} />
        ) : null}
        {items.length ? (
          <NotificationsList
            items={items}
            onMarkRead={(id) => void handleMarkRead(id)}
            onDelete={(id) => void handleDelete(id)}
            renderActions={renderActions}
            markReadLabel={t("notifications.markRead")}
          />
        ) : null}
        {hasMore && (
          <div className="flex justify-center">
            <Button variant="outline" disabled={notificationsQuery.isFetchingNextPage} onClick={() => void notificationsQuery.fetchNextPage()}>
              {notificationsQuery.isFetchingNextPage ? t("notifications.loading") : t("notifications.loadMore")}
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
