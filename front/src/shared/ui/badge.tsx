import * as React from "react";
import { cn } from "@/shared/lib/cn";

type BadgeTone = "default" | "success" | "warning" | "danger" | "muted";

const toneClassMap: Record<BadgeTone, string> = {
  default: "bg-[#ff5500]/20 text-[#ff7733]",
  success: "bg-emerald-900/60 text-emerald-400",
  warning: "bg-amber-900/60 text-amber-400",
  danger: "bg-red-900/60 text-red-400",
  muted: "bg-[#2a2a2a] text-[#9e9e9e]",
};

export function Badge({
  children,
  tone = "default",
  className,
}: {
  children: React.ReactNode;
  tone?: BadgeTone;
  className?: string;
}) {
  return (
    <span className={cn("inline-flex items-center rounded-full px-2.5 py-1 text-xs font-medium", toneClassMap[tone], className)}>
      {children}
    </span>
  );
}