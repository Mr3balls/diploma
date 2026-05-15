import * as React from "react";
export function FormField({
  label,
  error,
  hint,
  children,
}: {
  label: string;
  error?: string;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="flex flex-col gap-2 text-sm font-medium text-white">
      <span>{label}</span>
      {children}
      {hint ? <span className="text-xs font-normal text-[#9e9e9e]">{hint}</span> : null}
      {error ? <span className="text-xs font-normal text-red-400">{error}</span> : null}
    </label>
  );
}