import type { HTMLAttributes } from "react";

export const SPINNER_SIZES = {
  sm: "h-4 w-4 border-2",
  md: "h-6 w-6 border-2",
  lg: "h-8 w-8 border-[3px]",
} as const;

export type SpinnerSize = keyof typeof SPINNER_SIZES;

export type SpinnerProps = {
  size?: SpinnerSize;
  label?: string;
} & HTMLAttributes<HTMLSpanElement>;
