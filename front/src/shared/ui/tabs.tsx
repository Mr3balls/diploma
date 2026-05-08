import { cn } from "@/shared/lib/cn";

export function Tabs({
  value,
  onValueChange,
  tabs,
}: {
  value: string;
  onValueChange: (value: string) => void;
  tabs: { value: string; label: string }[];
}) {
  return (
    <div className="flex flex-wrap gap-2">
      {tabs.map((tab) => (
        <button
          key={tab.value}
          type="button"
          onClick={() => onValueChange(tab.value)}
          className={cn(
            "rounded-xl border px-3 py-2 text-sm font-medium transition-colors",
            value === tab.value ? "border-slate-900 bg-slate-900 text-white" : "border-border bg-white text-slate-700 hover:bg-slate-50",
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}