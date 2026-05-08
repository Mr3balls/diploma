import * as React from "react";
import { cn } from "@/shared/lib/cn";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "flex h-10 w-full rounded-xl border border-input bg-white px-3 py-2 text-sm text-slate-900 outline-none placeholder:text-slate-400 focus-visible:ring-2 focus-visible:ring-ring",
      className,
    )}
    {...props}
  />
));
Input.displayName = "Input";

export { Input };