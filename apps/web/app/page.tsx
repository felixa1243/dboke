"use client";

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ThemeToggle } from '../components/ThemeToggle';
import { authApi } from '../lib/api/auth';

export default function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [dbType, setDbType] = useState('mysql');
  const [port, setPort] = useState('3306');
  const router = useRouter();

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    const savedDbType = localStorage.getItem('dboke_last_dbType');
    const savedPort = localStorage.getItem('dboke_last_port');
    const savedUsername = localStorage.getItem('dboke_last_username');

    if (savedDbType) setDbType(savedDbType);
    if (savedPort) setPort(savedPort);
    if (savedUsername) setUsername(savedUsername);
  }, []);

  const handleDbTypeChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newType = e.target.value;
    setDbType(newType);
    switch (newType) {
      case 'mysql': setPort('3306'); break;
      case 'pgsql': setPort('5432'); break;
      case 'mongodb': setPort('27017'); break;
      case 'sqlite': setPort(''); break;
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    
    try {
      await authApi.login({ dbType, port, username, password });
      
      localStorage.setItem('dboke_last_dbType', dbType);
      localStorage.setItem('dboke_last_port', port);
      localStorage.setItem('dboke_last_username', username);
      
      router.push('/databases');
    } catch (err: any) {
      setError(err.message || 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen w-full flex items-center justify-center bg-white dark:bg-black text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black relative">
      <div className="absolute top-6 right-6">
        <ThemeToggle />
      </div>
      <div className="w-full max-w-sm p-8 animate-fade-in">
        <h1 className="text-2xl font-bold tracking-tight mb-1">Dboke</h1>
        <p className="text-gray-500 dark:text-gray-400 text-sm mb-8">Secure database connection</p>

        {error && (
          <div className="mb-6 p-3 text-sm border border-black dark:border-white bg-gray-50 dark:bg-gray-900">
            {error}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-5">
          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-2 space-y-1">
              <label className="text-xs font-semibold uppercase tracking-widest text-gray-500 dark:text-gray-400">Database</label>
              <select 
                value={dbType}
                onChange={handleDbTypeChange}
                className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent appearance-none rounded-none cursor-pointer"
                required
              >
                <option value="mysql" className="dark:bg-black">MySQL</option>
                <option value="pgsql" className="dark:bg-black">PostgreSQL</option>
                <option value="sqlite" className="dark:bg-black">SQLite</option>
                <option value="mongodb" className="dark:bg-black">MongoDB</option>
              </select>
            </div>
            <div className="col-span-1 space-y-1">
              <label className="text-xs font-semibold uppercase tracking-widest text-gray-500 dark:text-gray-400">Port</label>
              <input 
                type="text" 
                value={port}
                onChange={(e) => setPort(e.target.value)}
                disabled={dbType === 'sqlite'}
                className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent disabled:opacity-50 placeholder-gray-300 dark:placeholder-gray-600"
                placeholder={dbType === 'sqlite' ? '-' : 'Port'}
                required={dbType !== 'sqlite'}
              />
            </div>
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold uppercase tracking-widest text-gray-500 dark:text-gray-400">Username</label>
            <input 
              type="text" 
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent transition-colors placeholder-gray-300 dark:placeholder-gray-600"
              placeholder="admin"
              required
            />
          </div>

          <div className="space-y-1">
            <label className="text-xs font-semibold uppercase tracking-widest text-gray-500 dark:text-gray-400">Password</label>
            <input 
              type="password" 
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent transition-colors placeholder-gray-300 dark:placeholder-gray-600"
              placeholder="••••••••"
              required
            />
          </div>
          
          <button 
            type="submit" 
            disabled={loading}
            className="w-full py-3 mt-4 text-sm font-medium bg-black dark:bg-white text-white dark:text-black hover:bg-gray-800 dark:hover:bg-gray-200 disabled:opacity-70 transition-colors"
          >
            {loading ? 'Connecting...' : 'Authenticate'}
          </button>
        </form>
      </div>
    </main>
  );
}
