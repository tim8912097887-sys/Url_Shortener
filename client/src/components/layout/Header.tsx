import { Link } from "react-router";
import LogoutButton from "../auth/LogoutButton";
import { useAuthStore } from "../../store/useAuthStore";

export default function Header() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);

  return (
    <header className="border-b border-slate-200 bg-white">
      <div className="mx-auto flex h-16 max-w-5xl items-center justify-between px-6">
        <Link
          to="/"
          className="text-lg font-semibold tracking-tight text-slate-900"
        >
          Acme<span className="text-teal-600">.</span>
        </Link>

        {isAuthenticated ? (
          <div className="flex items-center gap-4">
            <span className="hidden text-sm text-slate-500 sm:inline">
              {user?.email ?? user?.username ?? "Signed in"}
            </span>
            <LogoutButton />
          </div>
        ) : (
          <nav className="flex items-center gap-2">
            <Link
              to="/login"
              className="rounded-lg px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-100"
            >
              Log in
            </Link>
            <Link
              to="/signup"
              className="rounded-lg bg-teal-600 px-3 py-2 text-sm font-medium text-white hover:bg-teal-700"
            >
              Sign up
            </Link>
          </nav>
        )}
      </div>
    </header>
  );
}
