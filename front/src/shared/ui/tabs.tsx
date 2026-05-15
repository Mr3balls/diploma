import { cn } from "@/shared/lib/cn";

export function Tabs({
  value,
  onValueChange,
  tabs,
  variant = "pills",
}: {
  value: string;
  onValueChange: (value: string) => void;
  tabs: { value: string; label: string }[];
  variant?: "pills" | "underline";
}) {
  if (variant === "underline") {
    return (
      <div className="flex gap-0 border-b border-[#2d2d2d]">
        {tabs.map((tab) => (
          <button
            key={tab.value}
            type="button"
            onClick={() => onValueChange(tab.value)}
            className={cn(
              "relative px-5 py-3 text-sm font-semibold transition-colors",
              value === tab.value
                ? "text-white after:absolute after:bottom-0 after:left-0 after:right-0 after:h-0.5 after:bg-[#ff5500]"
                : "text-[#666666] hover:text-[#9e9e9e]",
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>
    );
  }

  return (
    <div className="flex flex-wrap gap-2">
      {tabs.map((tab) => (
        <button
          key={tab.value}
          type="button"
          onClick={() => onValueChange(tab.value)}
          className={cn(
            "rounded-xl border px-3 py-2 text-sm font-medium transition-colors",
            value === tab.value
              ? "border-[#ff5500] bg-[#ff5500] text-white"
              : "border-border bg-[#1a1a1a] text-[#9e9e9e] hover:bg-[#2a2a2a]",
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}
