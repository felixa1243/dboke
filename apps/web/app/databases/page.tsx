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
  const [expandedTreeTable, setExpandedTreeTable] = useState<string | null>(null);
  
  // Table Schema State
  const [tableSchemas, setTableSchemas] = useState<Record<string, any[]>>({});
  const [expandedColumns, setExpandedColumns] = useState<string | null>(null);

  const toggleColumns = async (e: React.MouseEvent, tableName: string) => {
    e.stopPropagation();
    if (expandedColumns === tableName) {
      setExpandedColumns(null);
      return;
    }
    setExpandedColumns(tableName);
    if (!tableSchemas[tableName]) {
      try {
        const res = await databasesApi.getTableSchema(dbId as string, tableName);
        setTableSchemas(prev => ({ ...prev, [tableName]: res.columns || [] }));
      } catch (err) {
        console.error("Failed to load schema", err);
      }
    }
  };

  const [activeTab, setActiveTab] = useState<'tables' | 'data' | 'query'>('tables');
  
  // SQL Query State
  const [query, setQuery] = useState('');
  const [queryResult, setQueryResult] = useState<any>(null);
  const [queryLoading, setQueryLoading] = useState(false);
  const [queryError, setQueryError] = useState('');
  const [queryLimit, setQueryLimit] = useState(200);
  const [queryOffset, setQueryOffset] = useState(0);

  const [savedQueries, setSavedQueries] = useState<{name: string, query: string}[]>([]);
  const [sidebarTab, setSidebarTab] = useState<'tables' | 'saved'>('tables');

  // Table Data State
  const [dataResult, setDataResult] = useState<any>(null);
  const [dataLoading, setDataLoading] = useState(false);
  const [dataError, setDataError] = useState('');
  const [dataLimit, setDataLimit] = useState(200);
  const [dataOffset, setDataOffset] = useState(0);

  const fetchTableData = async (overrideOffset?: number) => {
    if (!expandedTable) return;
    setDataLoading(true);
    setDataError('');
    const offsetToUse = overrideOffset !== undefined ? overrideOffset : dataOffset;
    setDataOffset(offsetToUse);

    try {
      const q = `SELECT * FROM "${expandedTable}"`;
      const res = await databasesApi.executeQuery(dbId as string, q, [], dataLimit, offsetToUse);
      if (offsetToUse > 0 && dataResult?.rows) {
        setDataResult({
          ...res,
          rows: [...dataResult.rows, ...res.rows],
        });
      } else {
        setDataResult(res);
      }
    } catch (err: any) {
      setDataError(err.message || 'Failed to fetch data');
    } finally {
      setDataLoading(false);
    }
  };

  // Automatically fetch table data when switching to Data tab or when selected table changes
  useEffect(() => {
    if (activeTab === 'data' && expandedTable) {
      fetchTableData(0);
    } else if (activeTab === 'data' && !expandedTable) {
      setDataResult(null);
    }
  }, [activeTab, expandedTable, dataLimit]);

  useEffect(() => {
    if (dbId) {
      const q = localStorage.getItem(`dboke_queries_${dbId}`);
      if (q) {
        try {
          setSavedQueries(JSON.parse(q));
        } catch (e) {}
      }
    }
  }, [dbId]);

  const saveQuery = () => {
    if (!query.trim()) return;
    const name = prompt('Enter a name for this query:');
    if (name) {
      const newQueries = [...savedQueries, { name, query }];
      setSavedQueries(newQueries);
      localStorage.setItem(`dboke_queries_${dbId}`, JSON.stringify(newQueries));
      setSidebarTab('saved');
    }
  };

  const deleteSavedQuery = (e: React.MouseEvent, index: number) => {
    e.stopPropagation();
    const newQueries = savedQueries.filter((_, i) => i !== index);
    setSavedQueries(newQueries);
    localStorage.setItem(`dboke_queries_${dbId}`, JSON.stringify(newQueries));
  };

  const executeQuery = async (overrideOffset?: number) => {
    if (!query.trim()) return;
    setQueryLoading(true);
    setQueryError('');
    if (overrideOffset === 0) {
      setQueryResult(null);
    }
    
    const offsetToUse = overrideOffset !== undefined ? overrideOffset : queryOffset;
    setQueryOffset(offsetToUse);

    try {
      const res = await databasesApi.executeQuery(dbId as string, query, [], queryLimit, offsetToUse);
      if (offsetToUse > 0 && queryResult?.rows) {
        setQueryResult({
          ...res,
          rows: [...queryResult.rows, ...res.rows],
        });
      } else {
        setQueryResult(res);
      }
    } catch (err: any) {
      setQueryError(err.message || 'Query failed');
    } finally {
      setQueryLoading(false);
    }
  };

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
      <header className="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-8">
        <div>
          <h1 className="text-3xl font-light tracking-tight text-black dark:text-white mb-4">
            Database_{dbId}
          </h1>
          <div className="flex gap-4 border-b border-gray-200 dark:border-gray-800">
            <button 
              onClick={() => setActiveTab('tables')}
              className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'tables' ? 'border-b-2 border-black dark:border-white text-black dark:text-white' : 'text-gray-500 hover:text-black dark:hover:text-white'}`}
            >
              Tables
            </button>
            <button 
              onClick={() => setActiveTab('data')}
              className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'data' ? 'border-b-2 border-black dark:border-white text-black dark:text-white' : 'text-gray-500 hover:text-black dark:hover:text-white'}`}
            >
              Data
            </button>
            <button 
              onClick={() => setActiveTab('query')}
              className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'query' ? 'border-b-2 border-black dark:border-white text-black dark:text-white' : 'text-gray-500 hover:text-black dark:hover:text-white'}`}
            >
              SQL Console (DML)
            </button>
          </div>
        </div>
        
        {activeTab === 'tables' && (
          <div className="w-full md:w-72">
            <input 
              type="text" 
              placeholder="Search tables..." 
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full px-4 py-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none transition-colors bg-transparent placeholder-gray-400 dark:placeholder-gray-500 text-black dark:text-white"
            />
          </div>
        )}
      </header>

      {activeTab === 'tables' ? (
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
                          <div>
                            <div 
                              onClick={(e) => toggleColumns(e, table.name)}
                              className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"
                            >
                              <ChevronRightIcon className={`w-3.5 h-3.5 transition-transform ${expandedColumns === table.name ? 'rotate-90' : ''}`} />
                              <FolderIcon className="text-yellow-500 w-4 h-4" /> Columns
                            </div>
                            {expandedColumns === table.name && (
                              <div className="pl-6 pt-3 pb-1 space-y-2 text-gray-500 dark:text-gray-400 text-xs font-mono animate-in slide-in-from-top-1">
                                {tableSchemas[table.name] ? (
                                  tableSchemas[table.name].map((col: any) => (
                                    <div key={col.name} className="flex items-center gap-4">
                                      <div className="flex items-center gap-2 w-32">
                                        {col.isPrimaryKey ? <span title="Primary Key">🔑</span> : <span className="w-4 inline-block"></span>}
                                        <span className={col.isPrimaryKey ? 'text-black dark:text-white font-semibold' : ''}>{col.name}</span>
                                      </div>
                                      <span className="opacity-70">{col.type}</span>
                                      {!col.isNullable && <span className="opacity-50 text-[10px] uppercase">NOT NULL</span>}
                                    </div>
                                  ))
                                ) : (
                                  <div className="animate-pulse">Loading...</div>
                                )}
                              </div>
                            )}
                          </div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5"><FolderIcon className="text-yellow-500 w-4 h-4" /> Constraints</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5"><FolderIcon className="text-yellow-500 w-4 h-4" /> Foreign Keys</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5"><FolderIcon className="text-yellow-500 w-4 h-4" /> Indexes</div>
                          <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5"><FolderIcon className="text-yellow-500 w-4 h-4" /> Dependencies</div>
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
      ) : activeTab === 'data' ? (
        <div className="w-full flex-1 flex flex-col min-h-0">
          {!expandedTable ? (
            <div className="flex-1 flex items-center justify-center border border-dashed border-gray-200 dark:border-gray-800 rounded-lg p-12 text-gray-400">
              Please select a table from the Tables tab to view its data.
            </div>
          ) : (
            <div className="flex-1 flex flex-col gap-6 min-w-0 w-full">
              <div className="flex items-center gap-4 bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg p-4 shadow-sm">
                <div className="flex-1 flex items-center gap-2">
                  <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">Table:</span>
                  <span className="text-sm font-mono bg-white dark:bg-black px-2 py-1 rounded border border-gray-200 dark:border-gray-700">{expandedTable}</span>
                </div>
                <div className="flex items-center gap-2 bg-white dark:bg-black border border-gray-200 dark:border-gray-700 rounded-md px-2 py-1 shadow-sm">
                  <span className="text-xs font-semibold text-gray-500 uppercase tracking-widest">Limit</span>
                  <input 
                    type="number" 
                    min="1"
                    max="10000"
                    value={dataLimit}
                    onChange={e => setDataLimit(parseInt(e.target.value) || 200)}
                    className="w-16 text-sm text-center bg-transparent border-none outline-none text-black dark:text-white"
                  />
                </div>
                <button 
                  onClick={() => fetchTableData(0)}
                  disabled={dataLoading}
                  className="px-4 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-sm disabled:opacity-50"
                >
                  {dataLoading && dataOffset === 0 ? 'Fetching...' : 'Refresh'}
                </button>
              </div>

              {dataError && (
                <div className="p-4 bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400 border border-red-200 dark:border-red-900 rounded-lg text-sm font-mono whitespace-pre-wrap shadow-sm">
                  {dataError}
                </div>
              )}

              {dataResult && dataResult.columns && (
                <div className="border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden shadow-sm flex flex-col max-h-[700px]">
                  <div className="bg-gray-50 dark:bg-gray-900 p-3 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center">
                    <span className="text-xs font-semibold text-gray-500 uppercase tracking-widest">
                      Data Grid
                    </span>
                    {dataResult.rows && (
                      <div className="flex items-center gap-3">
                        <span className="text-xs text-gray-500 font-mono">
                          {dataResult.rows.length} rows fetched
                        </span>
                        {dataResult.rows.length >= dataLimit + dataOffset && (
                          <button 
                            onClick={() => fetchTableData(dataOffset + dataLimit)}
                            disabled={dataLoading}
                            className="px-3 py-1 bg-white dark:bg-black border border-gray-200 dark:border-gray-700 text-xs font-semibold rounded hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50"
                          >
                            {dataLoading ? 'Fetching...' : 'Fetch Next'}
                          </button>
                        )}
                      </div>
                    )}
                  </div>
                  <div className="overflow-auto flex-1 relative">
                    <table className="w-full text-left border-collapse whitespace-nowrap">
                      <thead>
                        <tr className="border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/50">
                          {dataResult.columns.map((col: string, i: number) => (
                            <th key={i} className="px-6 py-3 text-xs font-medium text-black dark:text-white uppercase">{col}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                        {dataResult.rows && dataResult.rows.map((row: any, i: number) => (
                          <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-900/50">
                            {dataResult.columns.map((col: string, j: number) => (
                              <td key={j} className="px-6 py-3 text-sm text-gray-600 dark:text-gray-300 max-w-xs truncate">
                                {row[col] === null ? <span className="text-gray-400 italic">NULL</span> : String(row[col])}
                              </td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                    {!dataResult.rows?.length && (
                      <div className="p-12 text-center text-sm text-gray-500">
                        No data available in this table
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      ) : (
        <div className="w-full flex flex-col md:flex-row gap-6 animate-in fade-in duration-300 items-start">
          
          {/* Sidebar for SQL Console */}
          <div className="w-full md:w-64 shrink-0 flex flex-col border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden bg-white dark:bg-black shadow-sm">
            <div className="flex border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900">
              <button 
                onClick={() => setSidebarTab('tables')}
                className={`flex-1 py-2.5 text-xs font-semibold text-center transition-colors ${sidebarTab === 'tables' ? 'bg-white dark:bg-black border-r border-gray-200 dark:border-gray-800 text-black dark:text-white' : 'text-gray-500 hover:text-black dark:hover:text-white'}`}
              >
                Tables
              </button>
              <button 
                onClick={() => setSidebarTab('saved')}
                className={`flex-1 py-2.5 text-xs font-semibold text-center transition-colors ${sidebarTab === 'saved' ? 'bg-white dark:bg-black border-l border-gray-200 dark:border-gray-800 text-black dark:text-white' : 'text-gray-500 hover:text-black dark:hover:text-white'}`}
              >
                Saved Queries
              </button>
            </div>

            {sidebarTab === 'tables' ? (
              <>
                <div className="p-3 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900">
                  <input 
                    type="text" 
                    placeholder="Filter tables..." 
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="w-full px-3 py-1.5 text-xs border border-gray-200 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none rounded bg-white dark:bg-black placeholder-gray-400 dark:placeholder-gray-500 text-black dark:text-white transition-colors"
                  />
                </div>
                <div className="overflow-y-auto max-h-[calc(100vh-250px)] min-h-[300px]">
              {filteredTables.length > 0 ? (
                filteredTables.map(t => (
                  <div key={t.name} className="border-b border-gray-100 dark:border-gray-800/50">
                    <div className="flex items-center group">
                      <button 
                        onClick={(e) => {
                          e.stopPropagation();
                          setExpandedTreeTable(prev => prev === t.name ? null : t.name);
                        }}
                        className="p-2 ml-1 text-gray-400 hover:text-black dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 rounded transition-colors"
                      >
                        <ChevronRightIcon className={`w-3.5 h-3.5 transition-transform ${expandedTreeTable === t.name ? 'rotate-90' : ''}`} />
                      </button>
                      <button 
                        onClick={() => {
                          setQuery(prev => prev + (prev.endsWith(' ') || prev.length === 0 ? '' : ' ') + t.name);
                        }}
                        title="Click to insert table name"
                        className="flex-1 text-left px-2 py-2.5 text-sm hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors flex items-center gap-2"
                      >
                        <span className="truncate flex-1 text-gray-700 dark:text-gray-300 group-hover:text-black dark:group-hover:text-white font-medium">{t.name}</span>
                      </button>
                    </div>
                    {expandedTreeTable === t.name && (
                      <div className="pl-10 pr-3 pb-3 space-y-2.5 text-xs text-gray-500 dark:text-gray-400 animate-in fade-in duration-200">
                        <div>
                          <div 
                            onClick={(e) => toggleColumns(e, t.name)}
                            className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors"
                          >
                            <ChevronRightIcon className={`w-3.5 h-3.5 transition-transform ${expandedColumns === t.name ? 'rotate-90' : ''}`} />
                            <FolderIcon className="w-3.5 h-3.5 text-yellow-500" /> Columns
                          </div>
                          {expandedColumns === t.name && (
                            <div className="pl-6 pt-2 pb-1 space-y-2 text-gray-400 dark:text-gray-500 text-[11px] font-mono animate-in slide-in-from-top-1">
                              {tableSchemas[t.name] ? (
                                tableSchemas[t.name].map((col: any) => (
                                  <div key={col.name} className="flex items-center justify-between">
                                    <div className="flex items-center gap-1.5 truncate pr-2">
                                      {col.isPrimaryKey && <span title="Primary Key">🔑</span>}
                                      <span className={`truncate ${col.isPrimaryKey ? 'text-black dark:text-white font-bold' : ''}`}>{col.name}</span>
                                    </div>
                                    <span className="opacity-60 shrink-0">{col.type}</span>
                                  </div>
                                ))
                              ) : (
                                <div className="animate-pulse">Loading...</div>
                              )}
                            </div>
                          )}
                        </div>
                        <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5">
                          <FolderIcon className="w-3.5 h-3.5 text-yellow-500" /> Constraints
                        </div>
                        <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5">
                          <FolderIcon className="w-3.5 h-3.5 text-yellow-500" /> Indexes
                        </div>
                        <div className="flex items-center gap-2 cursor-pointer hover:text-black dark:hover:text-white transition-colors ml-5">
                          <FolderIcon className="w-3.5 h-3.5 text-yellow-500" /> Foreign Keys
                        </div>
                      </div>
                    )}
                  </div>
                ))
              ) : (
                <div className="p-6 text-xs text-center text-gray-400 italic">No tables found</div>
              )}
            </div>
              </>
            ) : (
              <div className="overflow-y-auto max-h-[calc(100vh-250px)] min-h-[300px]">
                {savedQueries.length > 0 ? (
                  savedQueries.map((q, i) => (
                    <div key={i} className="group border-b border-gray-100 dark:border-gray-800/50 hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors p-3 flex justify-between items-start cursor-pointer" onClick={() => setQuery(q.query)}>
                      <div className="flex-1 overflow-hidden pr-3">
                        <div className="font-semibold text-sm text-black dark:text-white mb-1 truncate">{q.name}</div>
                        <div className="text-xs text-gray-500 font-mono truncate">{q.query}</div>
                      </div>
                      <button 
                        onClick={(e) => deleteSavedQuery(e, i)}
                        className="opacity-0 group-hover:opacity-100 p-1.5 text-gray-400 hover:text-red-500 transition-all rounded-md hover:bg-red-50 dark:hover:bg-red-900/30"
                      >
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
                      </button>
                    </div>
                  ))
                ) : (
                  <div className="p-6 text-xs text-center text-gray-400 italic">No saved queries</div>
                )}
              </div>
            )}
          </div>

          {/* Editor & Results */}
          <div className="flex-1 flex flex-col gap-6 min-w-0 w-full">
            <div className="relative shadow-sm rounded-lg">
              <textarea 
                value={query}
                onChange={e => setQuery(e.target.value)}
                placeholder="Enter SQL query (e.g. INSERT INTO users (name) VALUES ('Admin'), SELECT * FROM users...)"
                className="w-full h-48 p-4 font-mono text-sm border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 focus:border-black dark:focus:border-white outline-none rounded-lg resize-y transition-colors"
              />
              <div className="absolute bottom-4 left-4 flex items-center gap-4">
                <div className="flex items-center gap-2 bg-white dark:bg-black border border-gray-200 dark:border-gray-700 rounded-md px-2 py-1 shadow-sm">
                  <span className="text-xs font-semibold text-gray-500 uppercase tracking-widest">Limit</span>
                  <input 
                    type="number" 
                    min="1"
                    max="10000"
                    value={queryLimit}
                    onChange={e => setQueryLimit(parseInt(e.target.value) || 200)}
                    className="w-16 text-sm text-center bg-transparent border-none outline-none text-black dark:text-white"
                  />
                </div>
              </div>
              <div className="absolute bottom-4 right-4 flex gap-2">
                <button 
                  onClick={saveQuery}
                  className="px-4 py-2 bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300 text-sm font-semibold rounded-lg hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors shadow-sm"
                >
                  Save
                </button>
                <button 
                  onClick={() => executeQuery(0)}
                  disabled={queryLoading}
                  className="px-6 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-md disabled:opacity-50"
                >
                  {queryLoading && queryOffset === 0 ? 'Executing...' : 'Run Query'}
                </button>
              </div>
            </div>

            {queryError && (
              <div className="p-4 bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400 border border-red-200 dark:border-red-900 rounded-lg text-sm font-mono whitespace-pre-wrap shadow-sm">
                {queryError}
              </div>
            )}

            {queryResult && queryResult.columns && (
              <div className="border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden shadow-sm flex flex-col max-h-[600px]">
                <div className="bg-gray-50 dark:bg-gray-900 p-3 border-b border-gray-200 dark:border-gray-800 flex justify-between items-center">
                  <span className="text-xs font-semibold text-gray-500 uppercase tracking-widest">
                    Result Output
                  </span>
                  {queryResult.rows && (
                    <div className="flex items-center gap-3">
                      <span className="text-xs text-gray-500 font-mono">
                        {queryResult.rows.length} rows fetched
                      </span>
                      {queryResult.rows.length >= queryLimit + queryOffset && (
                        <button 
                          onClick={() => executeQuery(queryOffset + queryLimit)}
                          disabled={queryLoading}
                          className="px-3 py-1 bg-white dark:bg-black border border-gray-200 dark:border-gray-700 text-xs font-semibold rounded hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors disabled:opacity-50"
                        >
                          {queryLoading ? 'Fetching...' : 'Fetch Next'}
                        </button>
                      )}
                    </div>
                  )}
                </div>
                <div className="overflow-auto flex-1 relative">
                  <table className="w-full text-left border-collapse whitespace-nowrap">
                    <thead>
                      <tr className="border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/50">
                        {queryResult.columns.map((col: string, i: number) => (
                          <th key={i} className="px-6 py-3 text-xs font-medium text-black dark:text-white uppercase">{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-100 dark:divide-gray-800/50">
                      {queryResult.rows && queryResult.rows.length > 0 ? (
                        queryResult.rows.map((row: any, i: number) => (
                          <tr key={i} className="hover:bg-gray-50 dark:hover:bg-gray-900/50 transition-colors">
                            {queryResult.columns.map((col: string, j: number) => (
                              <td key={j} className="px-6 py-3 text-sm text-gray-600 dark:text-gray-300 font-mono">
                                {row[col] === null ? <span className="italic text-gray-400">null</span> : String(row[col])}
                              </td>
                            ))}
                          </tr>
                        ))
                      ) : (
                        <tr>
                          <td colSpan={queryResult.columns.length} className="py-8 text-center text-gray-400 text-sm italic">
                            Query executed successfully. (0 rows returned)
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
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
