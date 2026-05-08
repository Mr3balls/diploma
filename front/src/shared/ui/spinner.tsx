export function Spinner({ label = "Загрузка..." }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 text-sm text-slate-500">
      <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-800" />
      <span>{label}</span>
    </div>
  );
}