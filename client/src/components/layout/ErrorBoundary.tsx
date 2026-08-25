import { Component, type ErrorInfo, type ReactNode } from "react";

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
}

export default class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  state: ErrorBoundaryState = { hasError: false };

  static getDerivedStateFromError(_: Error): ErrorBoundaryState {
    return { hasError: true };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    // Swap for a real error-reporting call (Sentry, etc.) in production.
    console.error("Unhandled UI error:", error, info);
  }

  render(): ReactNode {
    if (this.state.hasError) {
      return (
        <div className="flex h-screen flex-col items-center justify-center gap-3 text-center px-4">
          <h1 className="text-xl font-semibold text-slate-900">
            Something went wrong.
          </h1>
          <p className="text-sm text-slate-500">
            Please refresh the page and try again.
          </p>
        </div>
      );
    }
    return this.props.children;
  }
}
