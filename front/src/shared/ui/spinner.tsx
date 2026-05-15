export function Spinner({ label = "Загрузка..." }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 text-sm text-[#9e9e9e]">
      <div className="h-4 w-4 animate-spin rounded-full border-2 border-[#2d2d2d] border-t-[#ff5500]" />
      <span>{label}</span>
    </div>
  );
}