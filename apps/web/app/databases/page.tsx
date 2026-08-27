export default function DatabasesIndex() {
  return (
    <div className="h-full min-h-screen flex items-center justify-center p-12 bg-white dark:bg-black">
      <div className="text-center max-w-md animate-fade-in">
        <h1 className="text-2xl font-light tracking-tight text-black dark:text-white mb-3">Select a Database</h1>
        <p className="text-gray-500 dark:text-gray-400 text-sm">Choose a database from the sidebar to view and manage its tables, or create a new one.</p>
      </div>
    </div>
  );
}
