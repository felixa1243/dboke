"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ThemeToggle } from "../../components/ThemeToggle";
import { LogoutButton } from "../../components/LogoutButton";
import { databasesApi } from "../../lib/api/databases";

export default function DatabasesLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [databases, setDatabases] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);

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
    <div className="min-h-screen w-full flex bg-white dark:bg-black text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      {/* Sidebar */}
      <aside className="w-64 border-r border-gray-200 dark:border-gray-800 flex flex-col bg-gray-50/50 dark:bg-gray-900/30">
        <div className="p-6 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="font-bold tracking-tight text-lg">Dboke</h2>
          <ThemeToggle />
        </div>
        
        <div className="flex-1 overflow-y-auto py-4 space-y-1">
          <div className="text-[10px] font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-3 mt-2 px-6">Databases</div>
          <div className="px-3">
            {loading ? (
              <div className="px-3 py-2 text-sm text-gray-500">Loading...</div>
            ) : databases.length > 0 ? (
              databases.map((dbName) => (
                <Link 
                  key={dbName} 
                  href={`/databases/${dbName}`}
                  className="block px-3 py-2 text-sm text-gray-600 dark:text-gray-400 hover:text-black dark:hover:text-white hover:bg-gray-200/50 dark:hover:bg-gray-800/50 rounded-md transition-colors"
                >
                  {dbName}
                </Link>
              ))
            ) : (
              <div className="px-3 py-2 text-sm text-gray-500">No databases found</div>
            )}
          </div>
        </div>
        
        <div className="p-4 border-t border-gray-200 dark:border-gray-800">
          <button className="w-full py-2 text-xs font-semibold uppercase tracking-widest border border-gray-200 dark:border-gray-700 hover:border-black dark:hover:border-white rounded-md transition-colors">
            New Database
          </button>
          <LogoutButton />
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 h-screen overflow-y-auto">
        {children}
      </div>
    </div>
  );
}
