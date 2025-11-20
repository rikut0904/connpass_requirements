'use client';

import { useRouter, usePathname } from 'next/navigation';
import Link from 'next/link';
import { logout } from '@/lib/auth';
import { Button } from '@/components/ui/button';

export function Header() {
  const router = useRouter();
  const pathname = usePathname();

  const handleLogout = async () => {
    try {
      await logout();
      router.replace('/');
    } catch (err) {
      console.error(err);
    }
  };

  const isActive = (path: string) => pathname === path;

  return (
    <header className="mb-6 flex items-center justify-between border-b border-slate-200 pb-4">
      <div className="flex items-center gap-6">
        <Link href="/dashboard" className="text-lg font-bold text-slate-800">
          Connpass Notifier
        </Link>
        <nav className="flex gap-4">
          <Link
            href="/dashboard"
            className={`text-sm ${isActive('/dashboard') ? 'font-semibold text-primary' : 'text-slate-600 hover:text-slate-800'}`}
          >
            ダッシュボード
          </Link>
          <Link
            href="/rules"
            className={`text-sm ${isActive('/rules') ? 'font-semibold text-primary' : 'text-slate-600 hover:text-slate-800'}`}
          >
            ルール管理
          </Link>
          <Link
            href="/status"
            className={`text-sm ${isActive('/status') ? 'font-semibold text-primary' : 'text-slate-600 hover:text-slate-800'}`}
          >
            ステータス
          </Link>
        </nav>
      </div>
      <Button variant="secondary" onClick={handleLogout}>
        ログアウト
      </Button>
    </header>
  );
}
