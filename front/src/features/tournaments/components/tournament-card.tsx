import { Link } from "react-router-dom";
import { Calendar, Users, Swords } from "lucide-react";
import type { Tournament } from "@/shared/types/api";
import { tournamentFormatLabel, tournamentStatusLabel } from "@/shared/lib/enums";
import { formatDate } from "@/shared/lib/date";

function statusAccent(status: Tournament["status"]): string {
  switch (status) {
    case "registration_open": return "#ff5500";
    case "in_progress":       return "#f59e0b";
    case "finished":
    case "completed":         return "#22c55e";
    case "cancelled":         return "#ef4444";
    default:                  return "#2d2d2d";
  }
}

function statusDot(status: Tournament["status"]) {
  const color = statusAccent(status);
  const label = tournamentStatusLabel[status] ?? status;
  return (
    <span className="flex items-center gap-1.5 text-xs font-medium" style={{ color }}>
      <span className="h-1.5 w-1.5 rounded-full shrink-0" style={{ background: color }} />
      {label}
    </span>
  );
}

export function TournamentCard({ tournament }: { tournament: Tournament }) {
  const accent = statusAccent(tournament.status);
  const date = tournament.start_at ?? tournament.registration_deadline ?? tournament.created_at;

  return (
    <Link to={`/tournaments/${tournament.id}`} className="group block">
      <div
        className="relative flex flex-col h-full rounded-xl bg-[#1a1a1a] border border-[#2d2d2d] overflow-hidden transition-all duration-200 group-hover:border-[#ff5500]/50 group-hover:bg-[#1f1f1f]"
        style={{ borderTopWidth: 2, borderTopColor: accent }}
      >
        <div className="flex flex-col gap-3 p-5 flex-1">
          {/* top row */}
          <div className="flex items-center justify-between gap-2">
            {statusDot(tournament.status)}
            {tournament.visibility === "private" && (
              <span className="text-[10px] uppercase tracking-wider text-[#666666] border border-[#2d2d2d] rounded px-1.5 py-0.5">
                Приватный
              </span>
            )}
          </div>

          {/* title */}
          <h3 className="text-base font-bold text-white leading-snug line-clamp-2 group-hover:text-[#ff7733] transition-colors">
            {tournament.title}
          </h3>

          {/* tags */}
          <div className="flex flex-wrap gap-1.5">
            {tournament.discipline && (
              <span className="flex items-center gap-1 rounded-md bg-[#2a2a2a] px-2 py-0.5 text-[11px] text-[#9e9e9e]">
                <Swords className="h-3 w-3" />
                {tournament.discipline}
              </span>
            )}
            <span className="rounded-md bg-[#2a2a2a] px-2 py-0.5 text-[11px] text-[#9e9e9e]">
              {tournamentFormatLabel[tournament.format] ?? tournament.format}
            </span>
            {tournament.registration_mode === "individual" && (
              <span className="rounded-md bg-[#2a2a2a] px-2 py-0.5 text-[11px] text-[#9e9e9e]">
                Соло
              </span>
            )}
          </div>
        </div>

        {/* bottom row */}
        <div className="flex items-center justify-between border-t border-[#2d2d2d] px-5 py-3 text-xs text-[#666666]">
          {tournament.max_teams ? (
            <span className="flex items-center gap-1.5">
              <Users className="h-3.5 w-3.5" />
              до {tournament.max_teams} участников
            </span>
          ) : (
            <span />
          )}
          {date && (
            <span className="flex items-center gap-1.5">
              <Calendar className="h-3.5 w-3.5" />
              {formatDate(date)}
            </span>
          )}
        </div>
      </div>
    </Link>
  );
}
