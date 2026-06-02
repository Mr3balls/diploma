import { useCallback, useEffect, useRef } from "react";
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
 *  Reconnects automatically on disconnect with a 3-second delay. */
export function useNotificationStream(enabled = true) {
  const queryClient = useQueryClient();
  const esRef = useRef<EventSource | null>(null);
  const retryRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (!enabled) return;

    function connect() {
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
        retryRef.current = setTimeout(connect, 3_000);
      };
    }

    connect();

    return () => {
      esRef.current?.close();
      esRef.current = null;
      if (retryRef.current !== null) {
        clearTimeout(retryRef.current);
        retryRef.current = null;
      }
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

export function useDeleteNotification() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => notificationsApi.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    },
  });
}

export function useDeleteAllNotifications() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: notificationsApi.deleteAll,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.notifications });
      await queryClient.invalidateQueries({ queryKey: queryKeys.unreadNotifications });
    },
  });
}

export function useNotificationPreferences() {
  return useQuery({
    queryKey: ["notification-preferences"],
    queryFn: notificationsApi.getPreferences,
  });
}

export function useSetNotificationPreferences() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (disabled: string[]) => notificationsApi.setPreferences(disabled),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["notification-preferences"] }),
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

function urlBase64ToUint8Array(base64String: string): ArrayBuffer {
  const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
  const rawData = atob(base64);
  const arr = new Uint8Array(rawData.length);
  for (let i = 0; i < rawData.length; i++) arr[i] = rawData.charCodeAt(i);
  return arr.buffer;
}

/** Requests browser push permission and registers a push subscription. */
export function useRegisterPush() {
  return useCallback(async () => {
    if (!("serviceWorker" in navigator) || !("PushManager" in window)) return;

    const permission = await Notification.requestPermission();
    if (permission !== "granted") return;

    const vapidKey = await notificationsApi.getVAPIDPublicKey();
    if (!vapidKey) return;

    const reg = await navigator.serviceWorker.register("/sw.js");
    await navigator.serviceWorker.ready;

    const sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(vapidKey),
    });

    const json = sub.toJSON() as { endpoint: string; keys: { p256dh: string; auth: string } };
    await notificationsApi.registerPush(json.endpoint, json.keys.p256dh, json.keys.auth);
  }, []);
}
