"use client";

import React from "react";

export default function LandingPage() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center relative overflow-hidden bg-white dark:bg-black">
      
      {/* Background Graphic */}
      <div className="absolute inset-0 flex items-center justify-center opacity-5 pointer-events-none select-none">
        <h1 className="text-[12rem] font-bold tracking-tighter sm:text-[16rem] md:text-[20rem]">Dboke</h1>
      </div>

      <div className="z-10 text-center max-w-3xl px-6">
        <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight mb-6">
          The Universal <br /> Database Manager
        </h1>
        <p className="text-lg md:text-xl text-gray-500 dark:text-gray-400 mb-10 max-w-2xl mx-auto font-medium">
          A secure, high-performance database management tool built with Go and Next.js. 
          Experience a minimalist, elegant aesthetic with native support for PostgreSQL, MySQL, and MongoDB.
        </p>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
          <a
            href="https://github.com/felixa1243/dboke/releases/latest"
            className="flex items-center gap-2 px-8 py-4 bg-black text-white dark:bg-white dark:text-black font-semibold rounded-full hover:opacity-90 transition-all shadow-lg hover:shadow-xl hover:-translate-y-0.5"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
              <polyline points="7 10 12 15 17 10" />
              <line x1="12" y1="15" x2="12" y2="3" />
            </svg>
            Download Latest Release
          </a>
          
          <a
            href="https://github.com/felixa1243/dboke"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 px-8 py-4 bg-gray-100 text-black dark:bg-gray-900 dark:text-white font-semibold rounded-full hover:bg-gray-200 dark:hover:bg-gray-800 transition-all"
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 0 0-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0 0 20 4.77 5.07 5.07 0 0 0 19.91 1S18.73.65 16 2.48a13.38 13.38 0 0 0-7 0C6.27.65 5.09 1 5.09 1A5.07 5.07 0 0 0 5 4.77a5.44 5.44 0 0 0-1.5 3.78c0 5.42 3.3 6.61 6.44 7A3.37 3.37 0 0 0 9 18.13V22" />
            </svg>
            View Source Code
          </a>
        </div>
      </div>
      
      {/* Features Showcase */}
      <div className="z-10 mt-24 grid grid-cols-1 md:grid-cols-3 gap-8 max-w-5xl px-6 w-full text-left">
        <div className="p-6 rounded-2xl bg-gray-50 dark:bg-gray-900/50 border border-gray-100 dark:border-gray-800">
          <div className="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400 mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
          </div>
          <h3 className="text-lg font-bold mb-2">Enterprise Security</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">AES-256 encryption, protected backend API endpoints, and CSRF token validation built-in.</p>
        </div>

        <div className="p-6 rounded-2xl bg-gray-50 dark:bg-gray-900/50 border border-gray-100 dark:border-gray-800">
          <div className="w-10 h-10 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400 mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path><polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline><line x1="12" y1="22.08" x2="12" y2="12"></line></svg>
          </div>
          <h3 className="text-lg font-bold mb-2">Plugin Ecosystem</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">Extend functionality seamlessly with compiled Go plugins injected directly into the UI workflow.</p>
        </div>

        <div className="p-6 rounded-2xl bg-gray-50 dark:bg-gray-900/50 border border-gray-100 dark:border-gray-800">
          <div className="w-10 h-10 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center text-green-600 dark:text-green-400 mb-4">
            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>
          </div>
          <h3 className="text-lg font-bold mb-2">High Performance</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 leading-relaxed">Blazing fast Go backend with clean architecture, combined with a highly optimized Next.js frontend.</p>
        </div>
      </div>
    </main>
  );
}
