export default function SecurityPage() {
  return (
    <>
      <h1 className="text-4xl font-light tracking-tight mb-4">Security Architecture</h1>
      <p className="text-gray-500 dark:text-gray-400 mb-12 text-lg">
        How Dboke keeps your database credentials secure.
      </p>

      <section className="space-y-8">
        <div>
          <h2 className="text-2xl font-medium tracking-tight mb-3">Zero-Trust Session Model</h2>
          <div className="p-6 rounded-xl border border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-900/30">
            <p className="text-gray-600 dark:text-gray-400">
              Dboke utilizes a Zero-Trust Session model. The moment your backend server restarts, all database 
              configurations are wiped from memory. We leverage strict HTTP-Only, SameSite cookies alongside 
              CSRF tokens to guarantee enterprise-grade security over your local data.
            </p>
          </div>
        </div>
      </section>
    </>
  );
}
