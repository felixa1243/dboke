export default function PluginsDocsPage() {
  return (
    <>
      <h1 className="text-4xl font-light tracking-tight mb-4">Plugin Development</h1>
      <p className="text-gray-500 dark:text-gray-400 mb-12 text-lg">
        Extend Dboke with custom plugin bundles.
      </p>

      <section className="space-y-8">
          <div className="p-6 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30">
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Dboke features a powerful plugin architecture. You can seamlessly extend Dboke with custom features (like Visual Query Builders, Data Generators, etc.) by uploading a <code>.zip</code> bundle that contains both Go backend logic and React frontend UI.
            </p>

            <h3 className="text-lg font-semibold mb-2">1. Backend Development (Go)</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Your backend plugin runs as an isolated process using HashiCorp's <code>go-plugin</code> framework over gRPC. It must implement the shared Dboke interface. Compile this into a binary named <code>backend.exe</code> (or <code>backend</code> on Linux/Mac).
            </p>
            <pre className="p-4 bg-black text-gray-300 rounded-lg text-sm overflow-x-auto mb-6">
{`package main
import (
  "github.com/hashicorp/go-plugin"
  "dboke-plugins/shared"
)

type MyPlugin struct{}
func (m *MyPlugin) GetFrontendComponent() (string, error) { return "/databases/plugins/my-plugin", nil }
func (m *MyPlugin) BuildQuery(payload string) (string, error) { return "SELECT 1", nil }

func main() {
  plugin.Serve(&plugin.ServeConfig{
    HandshakeConfig: plugin.HandshakeConfig{ProtocolVersion: 1, MagicCookieKey: "DBOKE_PLUGIN", MagicCookieValue: "secure_plugin_handshake"},
    Plugins: map[string]plugin.Plugin{"feature": &shared.FeaturePlugin{Impl: &MyPlugin{}}},
  })
}`}
            </pre>

            <h3 className="text-lg font-semibold mb-2">2. Frontend Development (React / Next.js)</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Your frontend is dynamically injected directly into Dboke's Next.js App Router! Create a file at <code>frontend/page.tsx</code> containing a default exported React component. Because it runs natively inside Dboke, you can use Tailwind CSS.
            </p>
            <pre className="p-4 bg-black text-gray-300 rounded-lg text-sm overflow-x-auto mb-6">
{`"use client";
import React from 'react';

export default function CustomPluginPage() {
  return (
    <div className="p-8 max-w-5xl mx-auto animate-in fade-in duration-500">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold tracking-tight mb-2">My Custom Plugin</h1>
          <p className="text-gray-500 dark:text-gray-400">
            A native extension seamlessly integrated into Dboke.
          </p>
        </div>
        <button 
          onClick={() => alert('Executing Custom Action!')}
          className="px-4 py-2 bg-black text-white dark:bg-white dark:text-black text-sm font-semibold rounded-lg hover:opacity-90 transition-opacity shadow-lg shadow-black/10 dark:shadow-white/10"
        >
          Execute Action
        </button>
      </div>

      <div className="rounded-xl border border-gray-200 dark:border-gray-800 bg-white/50 dark:bg-black/50 backdrop-blur-sm p-8 shadow-sm">
        <h2 className="text-lg font-semibold mb-4">Plugin Dashboard</h2>
        <p className="text-sm text-gray-600 dark:text-gray-300">
          This UI is rendered natively inside the Next.js App Router, meaning you have full access to Tailwind CSS, animations, and the user's current Light/Dark mode state.
        </p>
      </div>
    </div>
  );
}`}
            </pre>

            <h3 className="text-lg font-semibold mb-2">3. Packaging & Installation</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Once you have compiled <code>backend.exe</code> and created <code>frontend/page.tsx</code>, compress them into a single <code>.zip</code> file structure:
            </p>
            <pre className="p-4 bg-black text-gray-300 rounded-lg text-sm overflow-x-auto mb-6">
{`my-plugin.zip/
├── backend.exe
└── frontend/
    └── page.tsx`}
            </pre>
            <p className="text-gray-600 dark:text-gray-400">
              Navigate to the <b>Plugins</b> tab in Dboke, click <b>(+) Add Plugin</b>, and upload your <code>.zip</code> file. Dboke will securely extract the backend binary into the background engine, and dynamically hot-reload your <code>page.tsx</code> directly into the web application. Your plugin will be instantly live!
            </p>
            <h3 className="text-lg font-semibold mb-2 mt-8">4. Managing Plugins</h3>
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Dboke provides full lifecycle management for your external plugins directly from the UI:
            </p>
            <ul className="list-disc list-inside text-gray-600 dark:text-gray-400 space-y-2 mb-4 ml-2">
              <li><b>Launch:</b> Your active plugins automatically appear in the left sidebar under <i>External Tools</i>.</li>
              <li><b>Deactivate:</b> Clicking Deactivate instantly removes the plugin from the sidebar and securely hides its Next.js route (throwing a 404), ensuring complete disabling without data loss.</li>
              <li><b>Delete:</b> Permanently removes both the backend binary and the injected frontend UI folder from the Dboke server infrastructure.</li>
            </ul>
          </div>
      </section>
    </>
  );
}
