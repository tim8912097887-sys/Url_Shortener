import type { HTMLAttributes, ReactNode } from "react";

export type DividerProps = {
  label?: ReactNode;
} & HTMLAttributes<HTMLDivElement>;
