import { forwardRef } from "react";

import Spinner from "../spinner/Spinner";
import { cn } from "../utils/cn";

import { BUTTON_SIZES, BUTTON_VARIANTS, type ButtonProps } from "./types";

const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button(
  {
    as: Component = "button",
    variant = "primary",
    size = "md",
    isLoading = false,
    fullWidth = false,
    className,
    children,
    disabled,
    type = "button",
    ...props
  },
  ref,
) {
  const isDisabled = disabled || isLoading;
  const isNativeButton = Component === "button";

  return (
    <Component
      ref={ref}
      type={isNativeButton ? type : undefined}
      disabled={isNativeButton ? isDisabled : undefined}
      aria-disabled={!isNativeButton && isDisabled ? true : undefined}
      aria-busy={isLoading || undefined}
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition cursor-pointer",
        "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2",
        "disabled:cursor-not-allowed disabled:opacity-60",
        BUTTON_VARIANTS[variant],
        BUTTON_SIZES[size],
        fullWidth && "w-full",
        isDisabled && !isNativeButton && "pointer-events-none opacity-60",
        className,
      )}
      {...props}
    >
      {isLoading && (
        <>
          <Spinner size="sm" aria-hidden="true" />
          <span className="sr-only">Loading</span>
        </>
      )}

      {!isLoading && children}
    </Component>
  );
});

Button.displayName = "Button";

export default Button;
