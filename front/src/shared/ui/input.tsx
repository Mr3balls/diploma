import * as React from "react";
import { cn } from "@/shared/lib/cn";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {}

const Input = React.forwardRef<HTMLInputElement, InputProps>(({ className, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "flex h-10 w-full rounded-xl border border-[#2d2d2d] bg-[#2a2a2a] px-3 py-2 text-sm text-white outline-none placeholder:text-[#9e9e9e] focus-visible:ring-2 focus-visible:ring-[#ff5500]",
      className,
    )}
    {...props}
  />
));
Input.displayName = "Input";

export { Input };