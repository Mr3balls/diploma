export function Spinner({ label = "Загрузка..." }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 text-sm text-[#90afd4]">
      <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#0a3575] border-t-[#2255ff]" />
      <span>{label}</span>
    </div>
  );
}