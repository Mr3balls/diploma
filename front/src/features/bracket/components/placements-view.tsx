import { Trophy } from "lucide-react";
import { cn } from "@/shared/lib/cn";
import type { TeamPlacement } from "@/shared/types/api";
import { useLang } from "@/app/providers/lang-provider";

function placeLabel(from: number, to: number): string {
  if (from === to) return `${from}`;
  return `${from}–${to}`;
}

function medalColor(from: number): string {
  if (from === 1) return "text-yellow-400";
  if (from === 2) return "text-slate-300";
  if (from === 3) return "text-amber-600";
  return "text-[#666666]";
}

function PlacementCard({ p, inGameLabel }: { p: TeamPlacement; inGameLabel: string }) {
  const label = placeLabel(p.place_from, p.place_to);
  const isTop3 = p.place_from <= 3;
  const color = medalColor(p.place_from);

  return (
    <div
      className={cn(
        "flex flex-col items-center gap-2 rounded-xl border px-4 py-5 text-center transition-colors",
        p.is_active
          ? "border-[#ff5500]/30 bg-[#ff5500]/[0.05]"
          : "border-[#2d2d2d] bg-[#1a1a1a]",
        isTop3 && !p.is_active && "border-[#2d2d2d]",
      )}
    >
      {isTop3 && !p.is_active && (
        <Trophy className={cn("h-4 w-4 shrink-0", color)} />
      )}
      <span className={cn("text-2xl font-black leading-none", color)}>
        {label}
      </span>
      <span
        className={cn(
          "text-sm font-semibold leading-tight",
          p.is_active ? "text-[#ff5500]" : "text-white",
        )}
      >
        {p.team_name}
      </span>
      {p.is_active && (
        <span className="text-[10px] font-bold uppercase tracking-widest text-[#ff5500]">
          {inGameLabel}
        </span>
      )}
    </div>
  );
}

export function PlacementsView({ placements }: { placements: TeamPlacement[] }) {
  const { t } = useLang();
  const inGameLabel = t("placement.inGame");

  if (!placements.length) return null;

  // Group by place_from so shared places appear side by side
  const grouped: TeamPlacement[][] = [];
  let i = 0;
  while (i < placements.length) {
    const from = placements[i].place_from;
    let j = i;
    while (j < placements.length && placements[j].place_from === from) j++;
    grouped.push(placements.slice(i, j));
    i = j;
  }

  return (
    <div className="space-y-3">
      {grouped.map((group) => {
        const size = group.length;
        const cols =
          size === 1 ? "grid-cols-1 max-w-xs mx-auto" :
          size === 2 ? "grid-cols-2 max-w-md mx-auto" :
          size <= 4 ? "grid-cols-2 sm:grid-cols-4" :
          "grid-cols-2 sm:grid-cols-3 md:grid-cols-4";

        return (
          <div key={group[0].place_from} className={cn("grid gap-3", cols)}>
            {group.map((p) => (
              <PlacementCard key={p.team_id} p={p} inGameLabel={inGameLabel} />
            ))}
          </div>
        );
      })}
    </div>
  );
}
