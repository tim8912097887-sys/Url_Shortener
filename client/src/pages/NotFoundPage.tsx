import { Link } from "react-router";

export default function NotFoundPage() {
  return (
    <div className="flex h-[60vh] flex-col items-center justify-center gap-3 text-center">
      <h1 className="text-3xl font-semibold text-slate-900">404</h1>
      <p className="text-slate-500">This page doesn&apos;t exist.</p>
      <Link
        to="/"
        className="text-sm font-medium text-teal-600 hover:text-teal-700"
      >
        Go home
      </Link>
    </div>
  );
}
