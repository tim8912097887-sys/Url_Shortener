type AlertProps = {
  variant?: "error" | "success" | "info";
  children?: React.ReactNode;
  className?: string;
};

const VARIANTS = {
  error: "bg-red-50 text-red-700 border-red-200",
  success: "bg-teal-50 text-teal-700 border-teal-200",
  info: "bg-slate-50 text-slate-700 border-slate-200",
};

export default function Alert({
  variant = "info",
  children,
  className = "",
}: AlertProps) {
  if (!children) return null;
  return (
    <div
      role="alert"
      className={`rounded-lg border px-4 py-3 text-sm ${VARIANTS[variant]} ${className}`}
    >
      {children}
    </div>
  );
}
