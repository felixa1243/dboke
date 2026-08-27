"use client";

import { useState, useEffect, Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import { databasesApi, Table } from '../../lib/api/databases';

function DatabaseContent() {
  const searchParams = useSearchParams();
  const dbId = searchParams.get('id');

  const [tables, setTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');

  useEffect(() => {
    if (!dbId) return;
    setLoading(true);
    
    async function loadTables() {
      try {
        const res = await databasesApi.getTables(dbId as string);
        setTables(res.tables || []);
      } catch (err) {
        console.error("Failed to load tables:", err);
      } finally {
        setLoading(false);
      }
    }
    loadTables();
  }, [dbId]);

  if (!dbId) {
    return (
      <div className="h-full min-h-screen flex items-center justify-center p-12 bg-white dark:bg-black">
        <div className="text-center max-w-md animate-fade-in">
          <h1 className="text-2xl font-light tracking-tight text-black dark:text-white mb-3">Select a Database</h1>
          <p className="text-gray-500 dark:text-gray-400 text-sm">Choose a database from the sidebar to view and manage its tables, or create a new one.</p>
        </div>
      </div>
    );
  }

  const filteredTables = (tables || []).filter(t => t.name.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="p-8 md:p-12 animate-fade-in bg-white dark:bg-black min-h-full">
      <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12">
        <div>
          <h1 className="text-3xl font-light tracking-tight text-black dark:text-white mb-2">
            Database_{dbId}
          </h1>
          <p className="text-gray-500 dark:text-gray-400 text-sm">Viewing tables</p>
        </div>
        
        <div className="w-full md:w-72">
          <input 
            type="text" 
            placeholder="Search tables..." 
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-4 py-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none transition-colors bg-transparent placeholder-gray-400 dark:placeholder-gray-500 text-black dark:text-white"
          />
        </div>
      </header>

      <div className="w-full">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-800">
              <th className="py-4 text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest w-1/2">Table</th>
              <th className="py-4 text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest">Rows</th>
              <th className="py-4 text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-widest text-right">Size</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-100 dark:divide-gray-800/50">
            {loading ? (
              Array.from({ length: 4 }).map((_, i) => (
                <tr key={i} className="animate-pulse">
                  <td className="py-5"><div className="h-4 w-32 bg-gray-100 dark:bg-gray-800 rounded-sm"></div></td>
                  <td className="py-5"><div className="h-4 w-16 bg-gray-100 dark:bg-gray-800 rounded-sm"></div></td>
                  <td className="py-5"><div className="h-4 w-16 bg-gray-100 dark:bg-gray-800 rounded-sm ml-auto"></div></td>
                </tr>
              ))
            ) : filteredTables.length > 0 ? (
              filteredTables.map((table) => (
                <tr key={table.name} className="group hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors cursor-pointer">
                  <td className="py-4 text-sm font-medium text-black dark:text-white">{table.name}</td>
                  <td className="py-4 text-sm text-gray-500 dark:text-gray-400">{table.rows}</td>
                  <td className="py-4 text-sm text-gray-500 dark:text-gray-400 text-right">{table.size}</td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={3} className="py-12 text-center text-gray-400 dark:text-gray-500 text-sm font-light">
                  No tables matching "{search}"
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function DatabasesIndex() {
  return (
    <Suspense fallback={<div className="p-12 text-gray-500">Loading...</div>}>
      <DatabaseContent />
    </Suspense>
  );
}
