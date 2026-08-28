"use client";

import React, { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";

interface TableInfo {
  name: string;
}

interface ColumnInfo {
  name: string;
  type: string;
}

export default function DataSeederPlugin() {
  const searchParams = useSearchParams();
  const initialDb = searchParams.get("id") || "";
  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const [databases, setDatabases] = useState<string[]>([]);
  const [selectedDb, setSelectedDb] = useState(initialDb);

  const [tables, setTables] = useState<TableInfo[]>([]);
  const [selectedTable, setSelectedTable] = useState("");

  const [columns, setColumns] = useState<ColumnInfo[]>([]);
  const [columnMap, setColumnMap] = useState<Record<string, string>>({});
  
  const [rowCount, setRowCount] = useState(10);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [error, setError] = useState("");

  // Fetch Databases
  useEffect(() => {
    fetch(`${API_URL}/api/v1/databases`, { credentials: 'include' })
      .then((res) => res.json())
      .then((data) => {
        if (data.databases) {
          setDatabases(data.databases);
          if (!selectedDb && data.databases.length > 0) {
            setSelectedDb(data.databases[0]);
          }
        }
      })
      .catch((e) => console.error(e));
  }, []);

  // Fetch Tables
  useEffect(() => {
    if (!selectedDb) return;
    setTables([]);
    setSelectedTable("");
    setColumns([]);
    
    fetch(`${API_URL}/api/v1/databases/${selectedDb}/tables`, { credentials: 'include' })
      .then((res) => res.json())
      .then((data) => {
        if (data.tables) setTables(data.tables);
      })
      .catch((e) => console.error(e));
  }, [selectedDb]);

  // Fetch Columns
  useEffect(() => {
    if (!selectedDb || !selectedTable) return;
    
    fetch(`${API_URL}/api/v1/databases/${selectedDb}/tables/${selectedTable}/columns`, { credentials: 'include' })
      .then(async (res) => {
        if (!res.ok) throw new Error("Failed to fetch schema");
        return res.json();
      })
      .then((data) => {
        if (data.columns) {
          setColumns(data.columns);
          
          // Auto-guess some types based on name
          const newMap: Record<string, string> = {};
          data.columns.forEach((c: ColumnInfo) => {
            const lowerName = c.name.toLowerCase();
            if (lowerName.includes("id")) newMap[c.name] = "uuid";
            else if (lowerName.includes("email")) newMap[c.name] = "email";
            else if (lowerName.includes("name")) newMap[c.name] = "person";
            else if (lowerName.includes("date") || lowerName.includes("time") || lowerName.includes("created") || lowerName.includes("updated") || lowerName.includes("deleted") || c.type.toLowerCase().includes("timestamp") || c.type.toLowerCase().includes("date") || c.type.toLowerCase().includes("time")) newMap[c.name] = "date";
            else if (lowerName.includes("price") || lowerName.includes("amount") || lowerName.includes("cost") || lowerName.includes("total") || lowerName.includes("money")) newMap[c.name] = "money";
            else if (c.type.toLowerCase().includes("int") || c.type.toLowerCase().includes("num") || c.type.toLowerCase().includes("float") || c.type.toLowerCase().includes("double")) newMap[c.name] = "number";
            else if (c.type.toLowerCase().includes("uuid")) newMap[c.name] = "uuid";
            else newMap[c.name] = "word";
          });
          setColumnMap(newMap);
        }
      })
      .catch((e) => console.error(e));
  }, [selectedDb, selectedTable]);

  const handleSeed = async () => {
    if (!selectedDb || !selectedTable) return;
    setLoading(true);
    setError("");
    setResult(null);
    
    const payload = {
      table: selectedTable,
      columns: columnMap,
      count: rowCount,
    };
    
    try {
      // Must include credentials if sessions are used, assuming localhost or using browser cookies automatically if same origin.
      // We will add a small hack: dboke API uses Authorization Bearer token from localStorage? Actually Dboke stores session in cookies.
      // Wait, plugin_handler checks sessionID from contextkeys.SessionIDKey.
      const token = localStorage.getItem("dboke_token"); // if any, or it uses cookies.
      
      const csrf = localStorage.getItem('dboke_csrf_token') || '';
      const res = await fetch(`${API_URL}/api/v1/databases/${selectedDb}/plugins/data-seeder/execute`, {
        method: "POST",
        headers: { 
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf
        },
        body: JSON.stringify(payload),
        credentials: 'include',
      });
      
      let data;
      try {
        data = await res.json();
      } catch (err) {
        throw new Error("Failed to parse API response. The server may have returned an unexpected error.");
      }
      
      if (!res.ok) {
        throw new Error(data?.message || data?.error || "Failed to seed data");
      }
      setResult(data);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fakeDataTypes = [
    { value: "uuid", label: "UUID" },
    { value: "number", label: "Number (1-1000)" },
    { value: "money", label: "Money / Price" },
    { value: "person", label: "Person Name" },
    { value: "email", label: "Email Address" },
    { value: "date", label: "Date & Time" },
    { value: "word", label: "Random Word" },
    { value: "address", label: "Address" },
    { value: "company", label: "Company" },
    { value: "skip", label: "-- Skip Column --" }
  ];

  const updateCol = (colName: string, val: string) => {
    if (val === "skip") {
      const newMap = { ...columnMap };
      delete newMap[colName];
      setColumnMap(newMap);
    } else {
      setColumnMap({ ...columnMap, [colName]: val });
    }
  };

  return (
    <div className="p-8 max-w-4xl mx-auto animate-in fade-in duration-500">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight mb-2">Data Seeder</h1>
        <p className="text-gray-500 dark:text-gray-400">
          Quickly populate your database tables with realistic fake data.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-8">
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 p-5 rounded-xl shadow-sm">
          <label className="block text-sm font-semibold mb-2">Target Database</label>
          <select 
            value={selectedDb} 
            onChange={(e) => setSelectedDb(e.target.value)}
            className="w-full p-2.5 bg-gray-50 dark:bg-black border border-gray-200 dark:border-gray-800 rounded-lg focus:ring-2 focus:ring-purple-500 outline-none"
          >
            <option value="">Select Database</option>
            {databases.map(db => (
              <option key={db} value={db}>{db}</option>
            ))}
          </select>
        </div>

        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 p-5 rounded-xl shadow-sm">
          <label className="block text-sm font-semibold mb-2">Target Table</label>
          <select 
            value={selectedTable} 
            onChange={(e) => setSelectedTable(e.target.value)}
            disabled={!selectedDb || tables.length === 0}
            className="w-full p-2.5 bg-gray-50 dark:bg-black border border-gray-200 dark:border-gray-800 rounded-lg focus:ring-2 focus:ring-purple-500 outline-none disabled:opacity-50"
          >
            <option value="">Select Table</option>
            {tables.map(t => (
              <option key={t.name} value={t.name}>{t.name}</option>
            ))}
          </select>
        </div>
      </div>

      {selectedTable && columns.length > 0 && (
        <div className="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-sm overflow-hidden mb-8 animate-in slide-in-from-bottom-4">
          <div className="p-5 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 flex justify-between items-center">
            <h2 className="font-semibold text-lg">Column Mapping</h2>
            <div className="flex items-center gap-3">
              <label className="text-sm font-medium">Rows to generate:</label>
              <input 
                type="number" 
                min="1" 
                max="1000"
                value={rowCount}
                onChange={(e) => setRowCount(parseInt(e.target.value) || 10)}
                className="w-24 p-1.5 bg-white dark:bg-black border border-gray-200 dark:border-gray-700 rounded-md text-center focus:ring-2 focus:ring-purple-500 outline-none"
              />
            </div>
          </div>
          
          <div className="p-5 space-y-4 max-h-[400px] overflow-y-auto">
            {columns.map(col => (
              <div key={col.name} className="flex items-center justify-between p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-800/50 border border-transparent hover:border-gray-100 dark:hover:border-gray-800 transition-colors">
                <div>
                  <div className="font-medium">{col.name}</div>
                  <div className="text-xs text-gray-500">{col.type}</div>
                </div>
                <select 
                  value={columnMap[col.name] || "skip"}
                  onChange={(e) => updateCol(col.name, e.target.value)}
                  className="w-48 p-2 text-sm bg-gray-50 dark:bg-black border border-gray-200 dark:border-gray-700 rounded-md focus:ring-2 focus:ring-purple-500 outline-none"
                >
                  {fakeDataTypes.map(ft => (
                    <option key={ft.value} value={ft.value}>{ft.label}</option>
                  ))}
                </select>
              </div>
            ))}
          </div>
          
          <div className="p-5 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
            <button 
              onClick={handleSeed}
              disabled={loading || Object.keys(columnMap).length === 0}
              className="w-full py-3 bg-purple-600 hover:bg-purple-700 text-white font-semibold rounded-lg shadow-md shadow-purple-500/20 transition-all disabled:opacity-50 flex justify-center items-center gap-2"
            >
              {loading ? (
                <>
                  <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Seeding Database...
                </>
              ) : (
                'Generate and Insert Data'
              )}
            </button>
          </div>
        </div>
      )}

      {error && (
        <div className="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-xl mb-6">
          <p className="font-semibold text-sm">Error</p>
          <p className="text-sm">{error}</p>
        </div>
      )}

      {result && (
        <div className="p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-400 rounded-xl mb-6 animate-in slide-in-from-bottom-2">
          <p className="font-semibold text-sm flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><path d="m9 11 3 3L22 4"/></svg>
            Success
          </p>
          <p className="text-sm mt-1">Successfully inserted {rowCount} rows into {selectedTable}.</p>
        </div>
      )}
    </div>
  );
}
