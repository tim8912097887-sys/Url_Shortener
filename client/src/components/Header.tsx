const Header = () => {
  return (
    <header className="bg-white border-b border-slate-200 py-5 px-6 w-full sticky top-0 z-10">
      <div className="max-w-5xl mx-auto flex items-center justify-between">
        <h1 className="font-bold text-2xl tracking-tight text-slate-900">
          Fast<span className="text-emerald-600">URL</span>
        </h1>
        <span className="text-xs font-medium text-slate-500 bg-slate-100 px-2.5 py-1 rounded-full">
          v1.0
        </span>
      </div>
    </header>
  );
};

export default Header;
