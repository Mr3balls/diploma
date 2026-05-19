import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { Send } from "lucide-react";
import { useTournamentChat, useSendMessage, useChatStream } from "@/features/chat/hooks";
import { Button } from "@/shared/ui/button";
import { Spinner } from "@/shared/ui/spinner";
import { getErrorMessage } from "@/shared/lib/http";
import type { TournamentMessage } from "@/shared/types/api";

function formatTime(iso: string) {
  const d = new Date(iso);
  return d.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
}

function formatDate(iso: string) {
  const d = new Date(iso);
  return d.toLocaleDateString("ru-RU", { day: "numeric", month: "long" });
}

function avatarLetter(nickname: string) {
  return (nickname[0] ?? "?").toUpperCase();
}

function MessageItem({ msg, isOwn }: { msg: TournamentMessage; isOwn: boolean }) {
  return (
    <div className={`flex gap-2 ${isOwn ? "flex-row-reverse" : ""}`}>
      {/* Avatar */}
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-[#2a2a2a] text-xs font-bold text-[#ff5500]">
        {avatarLetter(msg.user_nickname)}
      </div>
      <div className={`max-w-[70%] ${isOwn ? "items-end" : "items-start"} flex flex-col gap-0.5`}>
        <div className={`flex items-baseline gap-2 ${isOwn ? "flex-row-reverse" : ""}`}>
          <span className="text-xs font-semibold text-white">{msg.user_nickname}</span>
          <span className="text-[10px] text-[#555]">{formatTime(msg.created_at)}</span>
        </div>
        <div
          className={`rounded-2xl px-3 py-2 text-sm leading-relaxed ${
            isOwn
              ? "bg-[#ff5500] text-white rounded-tr-sm"
              : "bg-[#1e1e1e] text-[#e0e0e0] border border-[#2d2d2d] rounded-tl-sm"
          }`}
        >
          {msg.content}
        </div>
      </div>
    </div>
  );
}

function DateSeparator({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-3 py-1">
      <div className="h-px flex-1 bg-[#2d2d2d]" />
      <span className="text-[10px] uppercase tracking-widest text-[#555]">{label}</span>
      <div className="h-px flex-1 bg-[#2d2d2d]" />
    </div>
  );
}

export function ChatPanel({ tournamentId, currentUserId }: { tournamentId: string; currentUserId?: string }) {
  const chatQuery = useTournamentChat(tournamentId, Boolean(currentUserId));
  const sendMutation = useSendMessage(tournamentId);
  useChatStream(tournamentId, Boolean(currentUserId));

  const [text, setText] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const listRef = useRef<HTMLDivElement>(null);
  const isAtBottomRef = useRef(true);

  const allMessages = (chatQuery.data?.pages ?? []).flatMap((p) => p.items ?? []);

  // Scroll to bottom on new messages only if already near bottom
  useEffect(() => {
    if (isAtBottomRef.current) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [allMessages.length]);

  function handleScroll() {
    const el = listRef.current;
    if (!el) return;
    isAtBottomRef.current = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
  }

  async function handleSend() {
    const content = text.trim();
    if (!content) return;
    setText("");
    try {
      await sendMutation.mutateAsync(content);
      isAtBottomRef.current = true;
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    } catch (error) {
      toast.error(getErrorMessage(error));
      setText(content);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      void handleSend();
    }
  }

  // Group messages by date for separators
  const groups: Array<{ date: string; msgs: TournamentMessage[] }> = [];
  for (const msg of allMessages) {
    const date = formatDate(msg.created_at);
    const last = groups[groups.length - 1];
    if (last && last.date === date) {
      last.msgs.push(msg);
    } else {
      groups.push({ date, msgs: [msg] });
    }
  }

  return (
    <div className="flex flex-col rounded-xl border border-[#2d2d2d] bg-[#111111] overflow-hidden" style={{ height: 540 }}>

      {/* Header */}
      <div className="flex items-center gap-2 border-b border-[#2d2d2d] px-4 py-3">
        <span className="h-2 w-2 rounded-full bg-[#ff5500]" />
        <span className="text-sm font-semibold text-white">Чат турнира</span>
        {allMessages.length > 0 && (
          <span className="ml-auto text-xs text-[#555]">{allMessages.length} сообщ.</span>
        )}
      </div>

      {/* Messages */}
      <div
        ref={listRef}
        onScroll={handleScroll}
        className="flex-1 overflow-y-auto px-4 py-3 grid gap-3 content-start"
      >
        {/* Load more history */}
        {chatQuery.hasNextPage && (
          <div className="flex justify-center">
            <button
              className="text-xs text-[#666] hover:text-white transition-colors"
              disabled={chatQuery.isFetchingNextPage}
              onClick={() => void chatQuery.fetchNextPage()}
            >
              {chatQuery.isFetchingNextPage ? "Загрузка..." : "Загрузить старые"}
            </button>
          </div>
        )}

        {chatQuery.isLoading && <div className="flex justify-center py-4"><Spinner /></div>}

        {!chatQuery.isLoading && allMessages.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 gap-2 text-center">
            <p className="text-sm text-[#555]">Сообщений пока нет</p>
            <p className="text-xs text-[#444]">Начните общение первым</p>
          </div>
        )}

        {groups.map((g) => (
          <div key={g.date} className="grid gap-3">
            <DateSeparator label={g.date} />
            {g.msgs.map((msg) => (
              <MessageItem key={msg.id} msg={msg} isOwn={msg.user_id === currentUserId} />
            ))}
          </div>
        ))}

        <div ref={bottomRef} />
      </div>

      {/* Input */}
      {currentUserId ? (
        <div className="border-t border-[#2d2d2d] px-3 py-3 flex gap-2 items-end">
          <textarea
            rows={1}
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Написать сообщение... (Enter — отправить)"
            maxLength={1000}
            className="flex-1 resize-none rounded-lg border border-[#2d2d2d] bg-[#1a1a1a] px-3 py-2 text-sm text-white placeholder-[#555] focus:outline-none focus:border-[#ff5500] transition-colors"
            style={{ maxHeight: 120 }}
          />
          <Button
            size="sm"
            disabled={!text.trim() || sendMutation.isPending}
            onClick={() => void handleSend()}
            className="shrink-0"
          >
            <Send className="h-4 w-4" />
          </Button>
        </div>
      ) : (
        <div className="border-t border-[#2d2d2d] px-4 py-3 text-center text-xs text-[#555]">
          Войдите чтобы написать в чат
        </div>
      )}
    </div>
  );
}
