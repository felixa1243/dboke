"use client";

import { useRouter } from 'next/navigation';
import { authApi } from '../lib/api/auth';

export function LogoutButton() {
  const router = useRouter();

  const handleLogout = async () => {
    try {
      await authApi.logout();
    } catch (e) {
      console.error(e);
    } finally {
      router.push('/');
    }
  };

  return (
    <button 
      onClick={handleLogout}
      className="w-full py-2 mt-2 text-xs font-semibold uppercase tracking-widest border border-red-200 dark:border-red-900/50 text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 hover:border-red-500 rounded-md transition-colors"
    >
      Logout
    </button>
  );
}
