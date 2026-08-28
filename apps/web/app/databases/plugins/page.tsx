"use client";

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
interface PluginInfo {
  id: string;
  name: string;
  status: string;
  description: string;
  type: string;
}

export default function PluginsPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [pluginName, setPluginName] = useState("");
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [externalPlugins, setExternalPlugins] = useState<PluginInfo[]>([]);

  const fetchPlugins = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/v1/plugins');
      const data = await res.json();
      if (Array.isArray(data)) {
        setExternalPlugins(data);
      }
    } catch (e) {
      console.error("Failed to fetch plugins", e);
    }
  };

  useEffect(() => {
    fetchPlugins();
  }, []);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setSelectedFile(e.target.files[0]);
    }
  };

  const handleUpload = async () => {
    if (!pluginName || !selectedFile) {
      alert("Please provide a name and select a plugin executable.");
      return;
    }

    setUploading(true);
    const formData = new FormData();
    formData.append("name", pluginName);
    formData.append("executable", selectedFile);

    try {
      const res = await fetch("http://localhost:8080/api/v1/plugins/upload", {
        method: "POST",
        body: formData,
      });
      const data = await res.json();
      
      if (res.ok) {
        setIsModalOpen(false);
        setPluginName("");
        setSelectedFile(null);
        fetchPlugins(); // Refresh the list
      } else {
        alert("Upload failed: " + data.error);
      }
    } catch (err: any) {
      alert("Error: " + err.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="p-8 max-w-5xl mx-auto animate-in fade-in duration-500 relative">
      
      {/* Upload Plugin Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div 
            className="absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity"
            onClick={() => setIsModalOpen(false)}
          />
          <div className="relative z-10 w-full max-w-md bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-2xl p-6 animate-in zoom-in-95 duration-200">
            <h2 className="text-xl font-bold tracking-tight mb-1">Upload Custom Plugin</h2>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-6">
              Install a third-party plugin binary using HashiCorp go-plugin architecture.
            </p>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1.5 text-gray-700 dark:text-gray-300">
                  Plugin Name
                </label>
                <input 
                  type="text" 
                  value={pluginName}
                  onChange={(e) => setPluginName(e.target.value)}
                  placeholder="e.g. Visual Query Builder"
                  className="w-full px-3 py-2 bg-gray-50 dark:bg-black border border-gray-200 dark:border-gray-700 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-black dark:focus:ring-white transition-shadow"
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1.5 text-gray-700 dark:text-gray-300">
                  Plugin Bundle (.zip)
                </label>
                <div className={`mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-dashed rounded-lg transition-colors bg-gray-50 dark:bg-black/50 group ${selectedFile ? 'border-green-500 dark:border-green-400' : 'border-gray-300 dark:border-gray-700 hover:border-black dark:hover:border-white'}`}>
                  <div className="space-y-1 text-center">
                    <svg className={`mx-auto h-10 w-10 transition-colors ${selectedFile ? 'text-green-500' : 'text-gray-400 group-hover:text-gray-500 dark:group-hover:text-gray-300'}`} stroke="currentColor" fill="none" viewBox="0 0 48 48" aria-hidden="true">
                      <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                    <div className="flex text-sm text-gray-600 dark:text-gray-400 justify-center">
                      <label htmlFor="file-upload" className="relative cursor-pointer bg-transparent rounded-md font-medium text-blue-600 dark:text-blue-400 hover:text-blue-500 focus-within:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-blue-500">
                        <span>{selectedFile ? 'File selected' : 'Upload a .zip'}</span>
                        <input id="file-upload" name="file-upload" type="file" accept=".zip" className="sr-only" onChange={handleFileChange} />
                      </label>
                      {!selectedFile && <p className="pl-1">or drag and drop</p>}
                    </div>
                    <p className="text-xs text-gray-500 dark:text-gray-500">
                      {selectedFile ? selectedFile.name : 'Plugin bundle (.zip) up to 50MB'}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <div className="mt-8 flex justify-end gap-3">
              <button 
                onClick={() => {
                  setIsModalOpen(false);
                  setPluginName("");
                  setSelectedFile(null);
                }}
                disabled={uploading}
                className="px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50"
              >
                Cancel
              </button>
              <button 
                onClick={handleUpload}
                disabled={uploading}
                className="px-4 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-md disabled:opacity-50"
              >
                {uploading ? 'Installing...' : 'Install Plugin'}
              </button>
            </div>
          </div>
        </div>
      )}

      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight mb-2">Plugins</h1>
          <p className="text-gray-500 dark:text-gray-400">
            Manage installed extensions and adapters for Dboke.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={() => setIsModalOpen(true)}
            className="flex items-center gap-2 px-4 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-lg shadow-black/10 dark:shadow-white/10"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            Add Plugin
          </button>
        </div>
      </div>

      <h2 className="text-lg font-semibold mb-4">Core Adapters</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-8">
        {/* Postgres Plugin Card */}
        <div className="relative group overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-black/50 backdrop-blur-sm p-6 transition-all hover:border-black dark:hover:border-white">
          <div className="absolute inset-0 bg-gradient-to-br from-blue-500/5 to-purple-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
          <div className="flex items-start justify-between relative z-10">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400 font-bold">
                Pg
              </div>
              <div>
                <h3 className="font-semibold text-lg">PostgreSQL</h3>
                <p className="text-sm text-gray-500">Official Database Adapter</p>
              </div>
            </div>
            <span className="px-2.5 py-1 text-xs font-medium bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded-full">
              Installed
            </span>
          </div>
          <p className="mt-4 text-sm text-gray-600 dark:text-gray-300 relative z-10">
            Provides robust connection and schema exploration capabilities for PostgreSQL databases.
          </p>
        </div>

        {/* MySQL Plugin Card */}
        <div className="relative group overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-black/50 backdrop-blur-sm p-6 transition-all hover:border-black dark:hover:border-white">
          <div className="absolute inset-0 bg-gradient-to-br from-orange-500/5 to-red-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
          <div className="flex items-start justify-between relative z-10">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center text-orange-600 dark:text-orange-400 font-bold">
                My
              </div>
              <div>
                <h3 className="font-semibold text-lg">MySQL</h3>
                <p className="text-sm text-gray-500">Community Adapter</p>
              </div>
            </div>
            <span className="px-2.5 py-1 text-xs font-medium bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded-full">
              Installed
            </span>
          </div>
          <p className="mt-4 text-sm text-gray-600 dark:text-gray-300 relative z-10">
            Connect to MySQL and MariaDB instances. Supports standard queries and schema exploration.
          </p>
        </div>

        {/* MongoDB Plugin Card */}
        <div className="relative group overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-black/50 backdrop-blur-sm p-6 transition-all hover:border-black dark:hover:border-white">
          <div className="absolute inset-0 bg-gradient-to-br from-green-500/5 to-teal-500/5 opacity-0 group-hover:opacity-100 transition-opacity" />
          <div className="flex items-start justify-between relative z-10">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400 font-bold">
                Mg
              </div>
              <div>
                <h3 className="font-semibold text-lg">MongoDB</h3>
                <p className="text-sm text-gray-500">NoSQL Adapter</p>
              </div>
            </div>
            <span className="px-2.5 py-1 text-xs font-medium bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400 rounded-full">
              Installed
            </span>
          </div>
          <p className="mt-4 text-sm text-gray-600 dark:text-gray-300 relative z-10">
            Connect to MongoDB clusters and explore NoSQL document collections directly.
          </p>
        </div>
      </div>

      {externalPlugins.length > 0 && (
        <>
          <h2 className="text-lg font-semibold mb-4">External Plugins</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {externalPlugins.map((plugin) => (
              <div key={plugin.id} className="relative group overflow-hidden rounded-xl border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-black/50 backdrop-blur-sm p-6 transition-all hover:border-black dark:hover:border-white">
                <div className="absolute inset-0 bg-gradient-to-br from-purple-500/5 to-pink-500/5 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none" />
                <div className="flex items-start justify-between relative z-10">
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold transition-colors ${plugin.status === 'Active' ? 'bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'}`}>
                      {plugin.name.substring(0, 2).toUpperCase()}
                    </div>
                    <div>
                      <h3 className={`font-semibold text-lg transition-colors ${plugin.status === 'Inactive' ? 'text-gray-400' : ''}`}>{plugin.name}</h3>
                      <p className="text-sm text-gray-500">External Plugin</p>
                    </div>
                  </div>
                  <span className={`px-2.5 py-1 text-xs font-medium rounded-full ${plugin.status === 'Active' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' : 'bg-gray-200 text-gray-600 dark:bg-gray-700 dark:text-gray-400'}`}>
                    {plugin.status}
                  </span>
                </div>
                <p className={`mt-4 text-sm relative z-10 ${plugin.status === 'Inactive' ? 'text-gray-400' : 'text-gray-600 dark:text-gray-300'}`}>
                  {plugin.description}
                </p>
                <div className="mt-6 flex items-center gap-3 relative z-10">
                  {plugin.status === 'Active' && (
                    <Link href={`/databases/plugins/${plugin.id}`} className="px-4 py-2 text-sm font-semibold bg-black text-white dark:bg-white dark:text-black rounded-lg hover:opacity-90 transition-opacity">
                      Open Plugin
                    </Link>
                  )}
                  <button 
                    onClick={async () => {
                      await fetch(`http://localhost:8080/api/v1/plugins/${plugin.id}/toggle`, { method: 'POST' });
                      fetchPlugins();
                    }}
                    className="px-4 py-2 text-sm font-medium border border-gray-200 dark:border-gray-700 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
                  >
                    {plugin.status === 'Active' ? 'Deactivate' : 'Activate'}
                  </button>
                  <button 
                    onClick={async () => {
                      if(confirm('Are you sure you want to completely delete this plugin?')) {
                        await fetch(`http://localhost:8080/api/v1/plugins/${plugin.id}`, { method: 'DELETE' });
                        fetchPlugins();
                      }
                    }}
                    className="px-4 py-2 text-sm font-medium text-red-600 dark:text-red-400 border border-red-200 dark:border-red-900/50 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors ml-auto"
                  >
                    Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  );
}
