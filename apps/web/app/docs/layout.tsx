import Link from "next/link";
import { ThemeToggle } from "../../components/ThemeToggle";

export default function DocsLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen w-full flex bg-white dark:bg-black text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      {/* Sidebar */}
      <aside className="w-64 border-r border-gray-200 dark:border-gray-800 flex flex-col bg-gray-50/50 dark:bg-gray-900/30">
        <div className="p-6 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="font-bold tracking-tight text-lg">Dboke Docs</h2>
          <ThemeToggle />
        </div>
        
        <div className="flex-1 overflow-y-auto py-4 space-y-1">
          <div className="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-3 mt-2 px-6">Navigation</div>
          <div className="px-3 space-y-1">
            <Link 
              href="/docs"
              className="block px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-black dark:hover:text-white hover:bg-gray-200/50 dark:hover:bg-gray-800/50 rounded-md transition-colors"
            >
              Getting Started
            </Link>
            <Link 
              href="/docs/deployment"
              className="block px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-black dark:hover:text-white hover:bg-gray-200/50 dark:hover:bg-gray-800/50 rounded-md transition-colors"
            >
              Deployment (NGINX & Apache)
            </Link>
          </div>
        </div>
        
        <div className="p-4 border-t border-gray-200 dark:border-gray-800">
          <Link href="/">
            <button className="w-full py-2 text-xs font-semibold uppercase tracking-widest border border-gray-200 dark:border-gray-700 hover:border-black dark:hover:border-white rounded-md transition-colors">
              Back to App
            </button>
          </Link>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 h-screen overflow-y-auto">
        <div className="max-w-4xl mx-auto p-8 md:p-12 animate-fade-in">
          {children}
        </div>
      </div>
    </div>
  );
}
