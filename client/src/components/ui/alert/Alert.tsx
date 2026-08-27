import { cn } from "../utils/cn";
import { ALERT_VARIANTS, type AlertProps } from "./types";

export default function Alert({
  variant = "info",
  children,
  className = "",
}: AlertProps) {
  if (children == null) return null;
  return (
    <div
      role="alert"
      className={cn(
        "rounded-lg border px-4 py-3 text-sm",
        ALERT_VARIANTS[variant],
        className,
      )}
    >
      {children}
    </div>
  );
}
