import { useEffect, useRef, useState } from "react";
import { LogOut, MonitorOff, ChevronDown } from "lucide-react";
import Button from "../ui/button";
import { useAuthStore } from "../../store/useAuthStore";

type LogoutMenuProps = {
  onLogout: () => Promise<void>;
  onLogoutAll: () => Promise<void>;
};

type LoadingState = "logout" | "logoutAll" | null;

export default function LogoutMenu({ onLogout, onLogoutAll }: LogoutMenuProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [loadingState, setLoadingState] = useState<LoadingState>(null);
  const isLoading = useAuthStore((state) => state.isLoading);

  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, []);

  return (
    <div ref={menuRef} className="relative inline-block">
      {/* Compact menu trigger */}
      <Button
        type="button"
        variant="secondary"
        size="sm"
        onClick={() => setIsOpen((prev) => !prev)}
        aria-haspopup="menu"
        aria-expanded={isOpen}
        className="gap-1.5"
      >
        <LogOut className="size-4" />
        <span className="hidden sm:inline">Logout</span>
        <ChevronDown
          className={`size-4 transition-transform ${
            isOpen ? "rotate-180" : ""
          }`}
        />
      </Button>

      {isOpen && (
        <div
          role="menu"
          className="
            absolute
            right-0
            top-full
            z-50
            mt-2
            w-60
            overflow-hidden
            rounded-xl
            border
            border-gray-200
            bg-white
            p-1
            shadow-lg
            shadow-black/10
          "
        >
          {/* Logout current device */}
          <Button
            type="button"
            role="menuitem"
            disabled={isLoading || loadingState != null}
            onClick={async () => {
              setLoadingState("logout");
              await onLogout();
            }}
            variant="ghost"
            className="
              flex
              w-full
              items-center
              gap-3
              px-3
              py-2.5
              text-left
              text-sm
              font-medium
              text-gray-700
              transition
              hover:bg-red-50
              hover:text-red-600
            "
          >
            <LogOut className="size-4 shrink-0" />

            <span className="flex-1">
              {loadingState === "logout" ? "Logging out..." : "Logout"}
            </span>
          </Button>

          <div className="my-1 border-t border-gray-100" />

          {/* Logout all devices */}
          <Button
            type="button"
            role="menuitem"
            disabled={isLoading || loadingState != null}
            onClick={async () => {
              setLoadingState("logoutAll");
              await onLogoutAll();
            }}
            variant="ghost"
            className="
              flex
              w-full
              items-center
              gap-3
              px-3
              py-2.5
              text-left
              text-sm
              font-medium
              text-red-600
              transition
              hover:bg-red-50
            "
          >
            <MonitorOff className="size-4 shrink-0" />

            <div className="flex flex-1 flex-col">
              <span>
                {loadingState === "logoutAll"
                  ? "Logging out all devices..."
                  : "Logout all devices"}
              </span>

              {loadingState === "logoutAll" && (
                <span className="mt-0.5 text-xs font-normal text-gray-400">
                  End all active sessions
                </span>
              )}
            </div>
          </Button>
        </div>
      )}
    </div>
  );
}
