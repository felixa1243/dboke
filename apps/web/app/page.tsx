"use client";

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { ThemeToggle } from '../components/ThemeToggle';
import { authApi } from '../lib/api/auth';

interface ConnectionProfile {
  id: string;
  name: string;
  dbType: string;
  port: string;
  username: string;
}

export default function WorkspacePage() {
  const router = useRouter();
  const [connections, setConnections] = useState<ConnectionProfile[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedDriver, setSelectedDriver] = useState<string | null>(null);
  
  // Form State
  const [connName, setConnName] = useState('');
  const [dbType, setDbType] = useState('pgsql');
  const [port, setPort] = useState('5432');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Auth Modal State
  const [authModalOpen, setAuthModalOpen] = useState(false);
  const [activeProfile, setActiveProfile] = useState<ConnectionProfile | null>(null);
  const [authPassword, setAuthPassword] = useState('');

  useEffect(() => {
    const saved = localStorage.getItem('dboke_connections');
    if (saved) {
      try {
        setConnections(JSON.parse(saved));
      } catch (e) {}
    }
  }, []);

  const saveConnection = (profile: ConnectionProfile) => {
    const newConns = [...connections, profile];
    setConnections(newConns);
    localStorage.setItem('dboke_connections', JSON.stringify(newConns));
  };

  const deleteConnection = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    const newConns = connections.filter(c => c.id !== id);
    setConnections(newConns);
    localStorage.setItem('dboke_connections', JSON.stringify(newConns));
  };

  const handleDriverSelect = (driver: string) => {
    setDbType(driver);
    setSelectedDriver(driver);
    switch (driver) {
      case 'mysql': setPort('3306'); break;
      case 'pgsql': setPort('5432'); break;
      case 'mongodb': setPort('27017'); break;
      case 'sqlite': setPort(''); break;
    }
    setConnName(`Local ${driver.toUpperCase()}`);
  };

  const handleConnect = async (profile: ConnectionProfile, overridePassword?: string) => {
    setLoading(true);
    setError('');
    try {
      await authApi.login({ 
        dbType: profile.dbType, 
        port: profile.port, 
        username: profile.username, 
        password: overridePassword || password 
      });
      router.push('/databases');
    } catch (err: any) {
      setError(err.message || 'Connection failed');
    } finally {
      setLoading(false);
    }
  };

  const handleFormSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const newProfile: ConnectionProfile = {
      id: Date.now().toString(),
      name: connName || `${dbType} connection`,
      dbType,
      port,
      username,
    };
    saveConnection(newProfile);
    await handleConnect(newProfile, password);
  };

  const drivers = [
    { id: 'pgsql', name: 'PostgreSQL', icon: 'Pg' },
    { id: 'mysql', name: 'MySQL', icon: 'My' },
    { id: 'mongodb', name: 'MongoDB', icon: 'Mg' },
    { id: 'sqlite', name: 'SQLite', icon: 'Sq' }
  ];

  return (
    <div className="h-screen w-full flex bg-gray-50 dark:bg-gray-950 text-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black overflow-hidden">
      
      {/* Sidebar - Connection Navigator */}
      <aside className="w-72 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-black flex flex-col z-10">
        <div className="p-5 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
          <h2 className="font-bold tracking-tight text-lg">Dboke Navigator</h2>
          <ThemeToggle />
        </div>
        
        <div className="p-4">
          <button 
            onClick={() => { setIsModalOpen(true); setSelectedDriver(null); setError(''); }}
            className="w-full flex items-center justify-center gap-2 py-2.5 px-4 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-lg shadow-black/10 dark:shadow-white/10"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            New Connection
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-3 pb-4">
          <div className="px-2 pt-2 pb-2">
            <span className="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest">Saved Connections</span>
          </div>
          
          {connections.length === 0 ? (
            <div className="px-2 py-4 text-xs text-gray-500 text-center">No saved connections</div>
          ) : (
            <div className="space-y-1">
              {connections.map(c => (
                <div 
                  key={c.id}
                  onClick={() => {
                    setActiveProfile(c);
                    setAuthPassword('');
                    setAuthModalOpen(true);
                  }}
                  className="group flex items-center justify-between px-3 py-2.5 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-black dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-900 rounded-lg transition-colors cursor-pointer"
                >
                  <div className="flex items-center gap-3 truncate">
                    <div className="w-2 h-2 rounded-full bg-green-500"></div>
                    <span className="truncate">{c.name}</span>
                  </div>
                  <button 
                    onClick={(e) => deleteConnection(e, c.id)}
                    className="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-red-500 transition-all"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M3 6h18"></path><path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path><path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      </aside>

      {/* Main Workspace Area */}
      <main className="flex-1 flex flex-col relative bg-gray-50 dark:bg-gray-950">
        <div className="absolute inset-0 flex items-center justify-center opacity-5 pointer-events-none">
           <h1 className="text-[12rem] font-bold tracking-tighter select-none">Dboke</h1>
        </div>
        
        {loading && (
          <div className="absolute top-4 right-4 bg-black dark:bg-white text-white dark:text-black px-4 py-2 rounded shadow-xl text-sm font-medium animate-pulse z-20">
            Connecting to database...
          </div>
        )}
        
        <div className="flex-1 flex items-center justify-center relative z-10">
          <div className="text-center">
            <div className="w-16 h-16 mx-auto mb-4 text-gray-300 dark:text-gray-800">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" strokeLinejoin="round"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path></svg>
            </div>
            <h3 className="text-xl font-medium text-gray-400 dark:text-gray-600">No active connection</h3>
            <p className="text-sm text-gray-400 dark:text-gray-600 mt-2">Select a profile from the navigator or add a new one.</p>
          </div>
        </div>
      </main>

      {/* New Connection Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity animate-in fade-in" onClick={() => setIsModalOpen(false)} />
          <div className="relative z-10 w-full max-w-3xl bg-white dark:bg-black border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl flex flex-col h-[600px] max-h-[90vh] animate-in zoom-in-95 duration-200 overflow-hidden">
            
            <div className="p-6 border-b border-gray-200 dark:border-gray-800">
              <h2 className="text-xl font-bold tracking-tight">Connect to a database</h2>
              <p className="text-sm text-gray-500 mt-1">Select your database driver to continue.</p>
            </div>

            <div className="flex-1 flex overflow-hidden">
              {!selectedDriver ? (
                /* Driver Selection Grid */
                <div className="flex-1 overflow-y-auto p-6 bg-gray-50 dark:bg-gray-950">
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                    {drivers.map(d => (
                      <button 
                        key={d.id} 
                        onClick={() => handleDriverSelect(d.id)}
                        className="flex flex-col items-center justify-center p-6 border border-gray-200 dark:border-gray-800 rounded-xl bg-white dark:bg-black hover:border-black dark:hover:border-white transition-all group"
                      >
                        <div className="w-12 h-12 rounded-full bg-gray-100 dark:bg-gray-900 flex items-center justify-center text-xl font-bold text-gray-600 dark:text-gray-400 group-hover:text-black dark:group-hover:text-white transition-colors mb-4">
                          {d.icon}
                        </div>
                        <span className="text-sm font-medium">{d.name}</span>
                      </button>
                    ))}
                  </div>
                </div>
              ) : (
                /* Connection Form */
                <div className="flex-1 overflow-y-auto p-8">
                  <button onClick={() => setSelectedDriver(null)} className="text-sm text-blue-500 hover:underline mb-6 flex items-center gap-1">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M19 12H5M12 19l-7-7 7-7"/></svg> Back to drivers
                  </button>
                  
                  <h3 className="text-lg font-semibold mb-6 flex items-center gap-2">
                    Configure {drivers.find(d => d.id === selectedDriver)?.name}
                  </h3>

                  {error && <div className="mb-6 p-3 text-sm bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400 rounded-lg">{error}</div>}

                  <form id="conn-form" onSubmit={handleFormSubmit} className="space-y-5 max-w-md">
                    <div className="space-y-1">
                      <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Connection Name</label>
                      <input type="text" value={connName} onChange={e => setConnName(e.target.value)} className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" placeholder="My Database" required />
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                      <div className="col-span-2 space-y-1">
                        <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Host</label>
                        <input type="text" defaultValue="localhost" className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" required />
                      </div>
                      <div className="col-span-1 space-y-1">
                        <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Port</label>
                        <input type="text" value={port} onChange={e => setPort(e.target.value)} className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" required />
                      </div>
                    </div>

                    <div className="space-y-1">
                      <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Username</label>
                      <input type="text" value={username} onChange={e => setUsername(e.target.value)} className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" required />
                    </div>

                    <div className="space-y-1">
                      <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Password</label>
                      <input type="password" value={password} onChange={e => setPassword(e.target.value)} className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" required />
                    </div>
                  </form>
                </div>
              )}
            </div>

            <div className="p-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 flex justify-end gap-3">
              <button onClick={() => setIsModalOpen(false)} className="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-800 rounded-lg transition-colors">
                Cancel
              </button>
              {selectedDriver && (
                <button type="submit" form="conn-form" disabled={loading} className="px-6 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-md disabled:opacity-50">
                  {loading ? 'Connecting...' : 'Connect & Save'}
                </button>
              )}
            </div>
            
          </div>
        </div>
      )}

      {/* Auth Password Modal */}
      {authModalOpen && activeProfile && (
        <div className="fixed inset-0 z-[60] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity animate-in fade-in" onClick={() => setAuthModalOpen(false)} />
          <div className="relative z-10 w-full max-w-sm bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl p-6 animate-in zoom-in-95 duration-200">
            <h2 className="text-lg font-bold tracking-tight mb-1">Connect to {activeProfile.name}</h2>
            <p className="text-sm text-gray-500 mb-6">Enter password for <strong>{activeProfile.username}</strong></p>
            
            {error && <div className="mb-4 p-3 text-sm bg-red-50 text-red-600 dark:bg-red-900/20 dark:text-red-400 rounded-lg">{error}</div>}

            <form onSubmit={(e) => {
              e.preventDefault();
              handleConnect(activeProfile, authPassword);
            }}>
              <div className="space-y-1 mb-6">
                <label className="text-xs font-semibold uppercase tracking-widest text-gray-500">Password</label>
                <input 
                  type="password" 
                  autoFocus
                  value={authPassword} 
                  onChange={e => setAuthPassword(e.target.value)} 
                  className="w-full p-2 text-sm border-b border-gray-300 dark:border-gray-700 focus:border-black dark:focus:border-white outline-none bg-transparent" 
                  required 
                />
              </div>
              <div className="flex justify-end gap-3">
                <button type="button" onClick={() => setAuthModalOpen(false)} className="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-800 rounded-lg transition-colors">
                  Cancel
                </button>
                <button type="submit" disabled={loading} className="px-6 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-md disabled:opacity-50">
                  {loading ? 'Connecting...' : 'Connect'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
