import * as React from "react";
import { cn } from "@/shared/lib/cn";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "flex h-10 w-full rounded-xl border border-[#0a3575] bg-[#002366] px-3 py-2 text-sm text-white outline-none placeholder:text-[#90afd4] focus-visible:ring-2 focus-visible:ring-[#2255ff]",
      className,
    )}
    {...props}
  />
));
Input.displayName = "Input";

export { Input };