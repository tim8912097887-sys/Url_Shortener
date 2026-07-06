import Display from "./components/Display";
import Header from "./components/Header";

function App() {
  return (
    <div className="min-h-screen bg-slate-50 flex flex-col font-sans antialiased text-slate-800">
      <Header />
      <main className="flex-1 flex items-center justify-center p-4">
        <Display />
      </main>
    </div>
  );
}

export default App;
