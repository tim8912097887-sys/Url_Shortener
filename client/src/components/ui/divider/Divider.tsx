import { cn } from "../utils/cn";
import type { DividerProps } from "./types";

export default function Divider({
  label,
  className = "",
  ...props
}: DividerProps) {
  return (
    <div className={cn("flex items-center gap-3", className)} {...props}>
      <span aria-hidden="true" className="h-px flex-1 bg-slate-200" />
      {label != null && (
        <span className="text-xs uppercase tracking-wide text-slate-400">
          {label}
        </span>
      )}
      <span aria-hidden="true" className="h-px flex-1 bg-slate-200" />
    </div>
  );
}
