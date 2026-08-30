type PaginationButtonProps = {
  direction: "previous" | "next";
  disabled: boolean;
  onClick: () => void;
  ariaLabel: string;
};

export const PaginationButton = ({
  direction,
  disabled,
  onClick,
  ariaLabel,
}: PaginationButtonProps) => {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
      className="
        flex h-10 w-10 items-center justify-center
        rounded-lg border border-slate-200
        bg-white text-slate-700
        shadow-sm
        transition
        hover:bg-slate-50
        focus:outline-none
        focus:ring-2
        focus:ring-slate-400
        focus:ring-offset-2
        disabled:cursor-not-allowed
        disabled:opacity-40
      "
    >
      {direction === "previous" ? (
        <svg
          className="h-5 w-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M15 19l-7-7 7-7"
          />
        </svg>
      ) : (
        <svg
          className="h-5 w-5"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
        >
          <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
        </svg>
      )}
    </button>
  );
};
