export default function DocsPage() {
  return (
    <>
      <h1 className="text-4xl font-light tracking-tight mb-4">Getting Started</h1>
      <p className="text-gray-500 dark:text-gray-400 mb-12 text-lg">
        Welcome to the official documentation for Dboke, the universal database manager.
      </p>

      <section className="space-y-8">
        <div>
          <h2 className="text-2xl font-medium tracking-tight mb-3">Introduction</h2>
          <div className="p-6 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30">
            <p className="text-gray-600 dark:text-gray-400 mb-4">
              Dboke allows you to connect to various databases securely without storing your plaintext passwords. 
              All connections are symmetrically encrypted in-memory using AES-256-GCM.
            </p>
            <ul className="list-disc list-inside space-y-2 text-gray-600 dark:text-gray-400">
              <li>Select your database type (e.g., PostgreSQL).</li>
              <li>Enter your database host port (default is usually 5432).</li>
              <li>Provide your username and password.</li>
              <li>Dboke will instantly connect and allow you to explore your schemas!</li>
            </ul>
          </div>
        </div>
      </section>
    </>
  );
}
