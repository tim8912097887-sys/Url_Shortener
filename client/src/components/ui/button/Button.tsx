import { forwardRef } from "react";
import Spinner from "../spinner/Spinner";
import { SIZES, VARIANTS, type ButtonProps } from "./types";

const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  {
    as: Component = "button",
    variant = "primary",
    size = "md",
    isLoading = false,
    fullWidth = false,
    className = "",
    children,
    disabled,
    type = "button",
    ...props
  },
  ref,
) {
  return (
    <Component
      ref={ref}
      type={Component === "button" ? type : undefined}
      disabled={disabled || isLoading}
      aria-busy={isLoading || undefined}
      className={[
        "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition cursor-pointer",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
        "disabled:cursor-not-allowed",
        VARIANTS[variant],
        SIZES[size],
        fullWidth ? "w-full" : "",
        className,
      ].join(" ")}
      {...props}
    >
      {isLoading && <Spinner size="sm" />}
      {!isLoading && children}
    </Component>
  );
});

export default Button;
