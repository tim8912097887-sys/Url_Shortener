export const ALERT_VARIANTS = {
  error: "border-red-200 bg-red-50 text-red-700",
  success: "border-teal-200 bg-teal-50 text-teal-700",
  info: "border-slate-200 bg-slate-50 text-slate-700",
} as const;

export type AlertVariant = keyof typeof ALERT_VARIANTS;

export type AlertProps = {
  variant?: AlertVariant;
  children?: React.ReactNode;
  className?: string;
};
