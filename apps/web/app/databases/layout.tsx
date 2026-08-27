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
    <div className="h-screen w-full flex overflow-hidden bg-white dark:bg-black text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
      {/* Sidebar */}
      <aside className="w-64 border-r border-gray-200 dark:border-gray-800 flex flex-col bg-gray-50/50 dark:bg-gray-900/30">
        <div className="p-6 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="font-bold tracking-tight text-lg">Dboke</h2>
          <ThemeToggle />
        </div>
        
        <div className="flex-1 overflow-x-hidden overflow-y-auto pb-4">
          {loading ? (
            <div className="px-6 py-4 text-sm text-gray-500">Loading databases...</div>
          ) : databases.length > 0 ? (
            <SidebarTree databases={databases} />
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
      <div className="flex-1 h-full overflow-y-auto">
        {children}
      </div>
    </div>
  );
}
