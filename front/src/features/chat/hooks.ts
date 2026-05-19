import { useEffect, useRef } from "react";
import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { chatApi } from "@/features/chat/api";
import { tokenStorage } from "@/shared/api/token-storage";
import { queryKeys } from "@/shared/lib/query-keys";
import type { TournamentMessage } from "@/shared/types/api";

const PAGE_SIZE = 50;

export function useTournamentChat(tournamentId: string, enabled = true) {
  return useInfiniteQuery({
    queryKey: queryKeys.tournamentChat(tournamentId),
    queryFn: ({ pageParam }: { pageParam?: string }) =>
      chatApi.list(tournamentId, PAGE_SIZE, pageParam),
    getNextPageParam: (firstPage) => {
      const items = firstPage.items ?? [];
      if (items.length === PAGE_SIZE) {
        return items[0]?.created_at;
      }
      return undefined;
    },
    initialPageParam: undefined as string | undefined,
    enabled,
    // Keep data fresh; don't refetch on window focus for chat
    refetchOnWindowFocus: false,
  });
}

export function useSendMessage(tournamentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (content: string) => chatApi.send(tournamentId, content),
    onSuccess: (newMsg) => {
      // Optimistically append the new message to the last page
      queryClient.setQueryData(
        queryKeys.tournamentChat(tournamentId),
        (old: { pages: { items: TournamentMessage[] }[]; pageParams: unknown[] } | undefined) => {
          if (!old) return old;
          const pages = [...old.pages];
          const lastPage = pages[pages.length - 1];
          pages[pages.length - 1] = {
            ...lastPage,
            items: [...(lastPage.items ?? []), newMsg],
          };
          return { ...old, pages };
        },
      );
    },
  });
}

/** Opens an SSE connection for tournament chat. Appends received messages directly to the cache. */
export function useChatStream(tournamentId: string, enabled = true) {
  const queryClient = useQueryClient();
  const esRef = useRef<EventSource | null>(null);

  useEffect(() => {
    if (!enabled) return;
    const token = tokenStorage.getAccessToken();
    if (!token) return;

    const base = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "http://localhost:8080";
    const url = `${base}/tournaments/${tournamentId}/chat/stream?token=${encodeURIComponent(token)}`;
    const es = new EventSource(url);
    esRef.current = es;

    es.onmessage = (e: MessageEvent) => {
      try {
        const msg = JSON.parse(e.data as string) as TournamentMessage;
        queryClient.setQueryData(
          queryKeys.tournamentChat(tournamentId),
          (old: { pages: { items: TournamentMessage[] }[]; pageParams: unknown[] } | undefined) => {
            if (!old) return old;
            // Avoid duplicate if optimistic update already added it
            const allItems = old.pages.flatMap((p) => p.items ?? []);
            if (allItems.some((m) => m.id === msg.id)) return old;
            const pages = [...old.pages];
            const lastPage = pages[pages.length - 1];
            pages[pages.length - 1] = {
              ...lastPage,
              items: [...lastPage.items, msg],
            };
            return { ...old, pages };
          },
        );
      } catch {
        // ignore parse errors
      }
    };

    es.onerror = () => {
      es.close();
      esRef.current = null;
    };

    return () => {
      es.close();
      esRef.current = null;
    };
  }, [enabled, tournamentId, queryClient]);
}
