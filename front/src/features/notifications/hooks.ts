import { useEffect, useRef } from "react";
import { useMutation, useQuery, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { notificationsApi } from "@/features/notifications/api";
import { tokenStorage } from "@/shared/api/token-storage";
import { queryKeys } from "@/shared/lib/query-keys";

const PAGE_SIZE = 20;

export function useNotifications(enabled = true) {
  return useInfiniteQuery({
    queryKey: queryKeys.notifications,
    queryFn: ({ pageParam = 0 }) => notificationsApi.list(PAGE_SIZE, pageParam as number),
    getNextPageParam: (lastPage, allPages) => {
      const loaded = allPages.flatMap((p) => p.items).length;
      return lastPage.items.length === PAGE_SIZE ? loaded : undefined;
    },
    initialPageParam: 0,
    enabled,
  });
}

export function useUnreadNotifications(enabled = true) {
  return useQuery({
    queryKey: queryKeys.unreadNotifications,
    queryFn: notificationsApi.unreadCount,
    enabled,
    refetchInterval: 30_000,
  });
}

/** Opens an SSE connection to receive real-time notification events.
 *  Invalidates the notifications and unread-count queries on each event. */
export function useNotificationStream(enabled = true) {
  const queryClient = useQueryClient();
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!enabled) return;
    const token = tokenStorage.getAccessToken();
    if (!token) return;

    const base = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "http://localhost:8080";
    const url = `${base}/notifications/stream?token=${encodeURIComponent(token)}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onmessage = () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      void queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    };

    es.onerror = () => {
      es.close();
      esRef.current = null;
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [enabled, queryClient]);
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
