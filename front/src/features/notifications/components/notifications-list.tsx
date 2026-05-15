import * as React from "react";
import {
  Users, CheckCircle, XCircle, Swords, Clock,
  Calendar, Trophy, Bell, Check,
} from "lucide-react";
import type { Notification } from "@/shared/types/api";
import { formatDateTime } from "@/shared/lib/date";
import { Button } from "@/shared/ui/button";
import { cn } from "@/shared/lib/cn";

type IconConfig = { icon: React.ElementType; color: string; bg: string };

const TYPE_ICON: Record<string, IconConfig> = {
  added_to_team:                  { icon: Users,        color: "#ff5500", bg: "#ff5500" },
  team_participation_confirmed:   { icon: CheckCircle,  color: "#22c55e", bg: "#22c55e" },
  team_participation_declined:    { icon: XCircle,      color: "#ef4444", bg: "#ef4444" },
  match_assigned:                 { icon: Swords,       color: "#ff5500", bg: "#ff5500" },
  match_time_changed:             { icon: Clock,        color: "#f59e0b", bg: "#f59e0b" },
  match_rescheduled:              { icon: Calendar,     color: "#f59e0b", bg: "#f59e0b" },
  match_cancelled:                { icon: XCircle,      color: "#ef4444", bg: "#ef4444" },
  result_submitted:               { icon: Trophy,       color: "#ff5500", bg: "#ff5500" },
  result_confirmed:               { icon: CheckCircle,  color: "#22c55e", bg: "#22c55e" },
  tournament_finished:            { icon: Trophy,       color: "#f59e0b", bg: "#f59e0b" },
};

function NotificationIcon({ type }: { type: string }) {
  const cfg = TYPE_ICON[type] ?? { icon: Bell, color: "#666666", bg: "#666666" };
  const Icon = cfg.icon;
  return (
    <div
      className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl"
      style={{ background: `${cfg.bg}1a` }}
    >
      <Icon className="h-4 w-4" style={{ color: cfg.color }} />
    </div>
  );
}

export function NotificationsList({
  items,
  onMarkRead,
  renderActions,
}: {
  items: Notification[];
  onMarkRead: (id: string) => void;
  renderActions?: (notification: Notification) => React.ReactNode;
}) {
  return (
    <div className="rounded-xl border border-[#2d2d2d] bg-[#1a1a1a] divide-y divide-[#2d2d2d] overflow-hidden">
      {items.map((n) => (
        <div
          key={n.id}
          className={cn(
            "relative flex gap-4 px-5 py-4 transition-colors",
            !n.is_read
              ? "bg-[#ff5500]/[0.04] hover:bg-[#ff5500]/[0.07]"
              : "hover:bg-[#2a2a2a]/40",
          )}
        >
          {/* unread accent line */}
          {!n.is_read && (
            <span className="absolute left-0 top-0 bottom-0 w-0.5 bg-[#ff5500] rounded-r" />
          )}

          <NotificationIcon type={n.type} />

          <div className="flex-1 min-w-0 space-y-1.5">
            <div className="flex flex-wrap items-start justify-between gap-2">
              <div className="flex items-center gap-2">
                <p className={cn("text-sm font-semibold", n.is_read ? "text-[#9e9e9e]" : "text-white")}>
                  {n.title}
                </p>
                {!n.is_read && (
                  <span className="h-1.5 w-1.5 rounded-full bg-[#ff5500] shrink-0" />
                )}
              </div>
              <span className="text-xs text-[#444444] shrink-0">{formatDateTime(n.created_at)}</span>
            </div>

            {n.message && (
              <p className="text-xs text-[#666666] leading-relaxed">{n.message}</p>
            )}

            {/* action buttons */}
            <div className="flex flex-wrap items-center gap-2 pt-1">
              {renderActions?.(n)}
              {!n.is_read && (
                <button
                  onClick={() => onMarkRead(n.id)}
                  className="flex items-center gap-1 text-xs text-[#444444] hover:text-[#9e9e9e] transition-colors"
                >
                  <Check className="h-3 w-3" />
                  Прочитано
                </button>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
