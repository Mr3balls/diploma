import * as React from "react";
import { cn } from "@/shared/lib/cn";

type BadgeTone = "default" | "success" | "warning" | "danger" | "muted";

const toneClassMap: Record<BadgeTone, string> = {
  default: "bg-slate-100 text-slate-800",
  success: "bg-emerald-100 text-emerald-800",
  warning: "bg-amber-100 text-amber-800",
  danger: "bg-red-100 text-red-800",
  muted: "bg-slate-50 text-slate-600",
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