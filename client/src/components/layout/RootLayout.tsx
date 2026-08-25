import { Outlet } from "react-router";
import Header from "./Header";

export default function RootLayout() {
  return (
    <div className="min-h-screen bg-slate-50">
      <Header />
      <Outlet />
    </div>
  );
}
