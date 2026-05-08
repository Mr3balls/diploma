import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { notificationsApi } from "@/features/notifications/api";
import { queryKeys } from "@/shared/lib/query-keys";

export function useNotifications(enabled = true) {
  return useQuery({
    queryKey: queryKeys.notifications,
    queryFn: notificationsApi.list,
    enabled,
  });
}

export function useUnreadNotifications(enabled = true) {
  return useQuery({
    queryKey: queryKeys.unreadNotifications,
    queryFn: notificationsApi.unreadCount,
    enabled,
  });
}

export function useMarkNotificationRead() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => notificationsApi.markRead(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    },
  });
}

export function useReadAllNotifications() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: notificationsApi.readAll,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    },
  });
}

export function useNotificationAction() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, payload }: { id: string; payload?: Record<string, unknown> }) =>
      notificationsApi.action(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    },
  });
}