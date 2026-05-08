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
            value === tab.value ? "border-[#2255ff] bg-[#2255ff] text-white" : "border-border bg-[#001f52] text-[#90afd4] hover:bg-[#002366]",
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  );
}