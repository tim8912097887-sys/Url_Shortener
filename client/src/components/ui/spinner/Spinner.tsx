import { cn } from "../utils/cn";
import { SPINNER_SIZES, type SpinnerProps } from "./types";

export default function Spinner({
  size = "md",
  // Can customize with more descriptive label like "Data fetching..."
  label = "Loading",
  className,
  ...props
}: SpinnerProps) {
  return (
    <span
      role="status"
      aria-label={label}
      className={cn(
        "inline-block animate-spin rounded-full border-current border-t-transparent",
        SPINNER_SIZES[size],
        className,
      )}
      {...props}
    />
  );
}
