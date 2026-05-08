import * as React from "react";
import { cn } from "@/shared/lib/cn";

type BadgeTone = "default" | "success" | "warning" | "danger" | "muted";

const toneClassMap: Record<BadgeTone, string> = {
  default: "bg-[#002366] text-white",
  success: "bg-emerald-900 text-emerald-300",
  warning: "bg-amber-900 text-amber-300",
  danger: "bg-red-900 text-red-300",
  muted: "bg-[#001f52] text-[#90afd4]",
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