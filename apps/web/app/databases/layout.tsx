"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ThemeToggle } from "../../components/ThemeToggle";
import { LogoutButton } from "../../components/LogoutButton";
import { databasesApi } from "../../lib/api/databases";
import { SidebarTree } from "../../components/SidebarTree";

export default function DatabasesLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [databases, setDatabases] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  useEffect(() => {
    async function loadDatabases() {
      try {
        const res = await databasesApi.getDatabases();
        setDatabases(res.databases);
      } catch (err: any) {
        if (err.status === 401 || err.status === 403) {
          router.push('/');
        }
        console.error("Failed to load databases:", err);
      } finally {
        setLoading(false);
      }
    }
    loadDatabases();
  }, []);

  return (
    <div className="h-screen w-full flex flex-col md:flex-row overflow-hidden bg-white dark:bg-black text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      
      {/* Mobile backdrop */}
      {isSidebarOpen && (
        <div 
          className="fixed inset-0 bg-black/60 backdrop-blur-sm z-40 md:hidden transition-opacity" 
          onClick={() => setIsSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside className={`fixed inset-y-0 left-0 z-50 w-72 md:w-64 transform ${isSidebarOpen ? 'translate-x-0' : '-translate-x-full'} md:relative md:translate-x-0 transition-transform duration-300 cubic-bezier(0.4, 0, 0.2, 1) border-r border-gray-200 dark:border-gray-800 flex flex-col bg-gray-50/95 dark:bg-gray-900/95 backdrop-blur-md md:bg-gray-50/50 md:dark:bg-gray-900/30`}>
        <div className="p-6 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="font-bold tracking-tight text-lg">Dboke</h2>
          <div className="flex items-center gap-3">
            <ThemeToggle />
            <button 
              className="md:hidden text-gray-500 hover:text-black dark:hover:text-white transition-colors"
              onClick={() => setIsSidebarOpen(false)}
            >
              <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
            </button>
          </div>
        </div>
        
        <div className="flex-1 overflow-x-hidden overflow-y-auto pb-4">
          {loading ? (
            <div className="px-6 py-4 text-sm text-gray-500">Loading databases...</div>
          ) : databases.length > 0 ? (
            <SidebarTree databases={databases} onNavigate={() => setIsSidebarOpen(false)} />
          ) : (
            <div className="px-6 py-4 text-sm text-gray-500">No databases found</div>
          )}
        </div>
        
        <div className="p-4 border-t border-gray-200 dark:border-gray-800">
          <button className="w-full py-2 text-xs font-semibold uppercase tracking-widest border border-gray-200 dark:border-gray-700 hover:border-black dark:hover:border-white rounded-md transition-colors">
            New Database
          </button>
          <LogoutButton />
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        
        {/* Mobile Header */}
        <div className="md:hidden flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-800 bg-white/80 dark:bg-black/80 backdrop-blur-md z-10">
          <button 
            onClick={() => setIsSidebarOpen(true)}
            className="text-gray-600 dark:text-gray-300 hover:text-black dark:hover:text-white transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="4" x2="20" y1="12" y2="12"/><line x1="4" x2="20" y1="6" y2="6"/><line x1="4" x2="20" y1="18" y2="18"/></svg>
          </button>
          <span className="font-semibold tracking-tight text-lg">Dboke</span>
          <div className="w-6" /> {/* Visual Balance */}
        </div>

        <div className="flex-1 overflow-y-auto">
          {children}
        </div>
      </div>
    </div>
  );
}
