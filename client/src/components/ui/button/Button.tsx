import { forwardRef } from "react";
import Spinner from "../spinner/Spinner";

type ButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  as?: React.ElementType;
  variant?: keyof typeof VARIANTS;
  size?: keyof typeof SIZES;
  isLoading?: boolean;
  fullWidth?: boolean;
};

const VARIANTS = {
  primary:
    "bg-teal-600 text-white hover:bg-teal-700 focus-visible:ring-teal-300 disabled:bg-teal-300",
  secondary:
    "bg-white text-slate-700 border border-slate-300 hover:bg-slate-50 focus-visible:ring-slate-200 disabled:text-slate-400",
  ghost:
    "bg-transparent text-slate-600 hover:bg-slate-100 focus-visible:ring-slate-200",
  danger:
    "bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-300 disabled:bg-red-300",
};

const SIZES = {
  sm: "px-3 py-1.5 text-sm",
  md: "px-4 py-2.5 text-sm",
  lg: "px-5 py-3 text-base",
};

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
        "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition",
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
