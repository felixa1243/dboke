"use client";

import { useState, useEffect, Suspense, Fragment } from 'react';
import { useSearchParams } from 'next/navigation';
import { databasesApi, Table } from '../../lib/api/databases';

const ChevronRightIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <polyline points="9 18 15 12 9 6" />
  </svg>
);

const FolderIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z" />
  </svg>
);

function DatabaseContent() {
  const searchParams = useSearchParams();
  const dbId = searchParams.get('id');

  const [tables, setTables] = useState<Table[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [expandedTable, setExpandedTable] = useState<string | null>(null);

  const toggleTable = (name: string) => {
    console.log("Clicked table:", name);
    setExpandedTable(prev => prev === name ? null : name);
  };

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
                <Fragment key={table.name}>
                  <tr 
                    onClick={() => toggleTable(table.name)}
                    className="group hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors cursor-pointer"
                  >
                    <td className="py-4 text-sm font-medium text-black dark:text-white flex items-center gap-2">
                      <ChevronRightIcon className={`text-gray-400 transition-transform ${expandedTable === table.name ? 'rotate-90' : ''}`} />
                      {table.name}
                    </td>
                    <td className="py-4 text-sm text-gray-500 dark:text-gray-400">{table.rows}</td>
                    <td className="py-4 text-sm text-gray-500 dark:text-gray-400 text-right">{table.size}</td>
                  </tr>
                  {expandedTable === table.name && (
                    <tr className="bg-gray-50/50 dark:bg-gray-900/20">
                      <td colSpan={3} className="px-8 py-4 border-l-2 border-blue-500">
                        <div className="space-y-3 text-sm text-gray-600 dark:text-gray-400 pl-4 py-2">
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Columns</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Constraints</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Foreign Keys</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Indexes</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Dependencies</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> References</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Partitions</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Triggers</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Rules</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"><FolderIcon className="text-yellow-500 w-4 h-4" /> Policies</div>
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
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
