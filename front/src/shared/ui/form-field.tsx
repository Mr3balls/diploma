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
    <label className="flex flex-col gap-2 text-sm font-medium text-slate-700">
      <span>{label}</span>
      {children}
      {hint ? <span className="text-xs font-normal text-slate-500">{hint}</span> : null}
      {error ? <span className="text-xs font-normal text-red-600">{error}</span> : null}
    </label>
  );
}