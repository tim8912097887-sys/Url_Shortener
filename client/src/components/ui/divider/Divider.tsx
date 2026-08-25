type DividerProps = {
  label?: string;
  className?: string;
};

export default function Divider({ label, className = "" }: DividerProps) {
  return (
    <div className={`flex items-center gap-3 ${className}`}>
      <span className="h-px flex-1 bg-slate-200" />
      {label && (
        <span className="text-xs uppercase tracking-wide text-slate-400">
          {label}
        </span>
      )}
      <span className="h-px flex-1 bg-slate-200" />
    </div>
  );
}
