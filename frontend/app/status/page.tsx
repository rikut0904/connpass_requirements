'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { SchedulerStatus } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { checkAuth } from '@/lib/auth';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { LogViewer } from '@/components/features/log-viewer';
import { Skeleton } from '@/components/ui/skeleton';
import { Header } from '@/components/features/header';

export default function StatusPage() {
  const router = useRouter();
  const [status, setStatus] = useState<SchedulerStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [authChecked, setAuthChecked] = useState(false);

  useEffect(() => {
    const check = async () => {
      const isLoggedIn = await checkAuth();
      if (!isLoggedIn) {
        router.replace('/');
        return;
      }
      setAuthChecked(true);
    };
    void check();
  }, [router]);

  useEffect(() => {
    if (!authChecked) return;

    const run = async () => {
      try {
        const { data } = await apiClient.get('/status');
        setStatus(data);
      } catch (err) {
        console.error(err);
        setError('ステータス情報の取得に失敗しました');
      }
    };
    void run();
  }, [authChecked]);

  if (!authChecked) {
    return (
      <main className="flex items-center justify-center p-10">
        <p className="text-slate-600">読み込み中...</p>
      </main>
    );
  }

  return (
    <>
      <Header />
      <main className="space-y-6">
        <h1 className="text-2xl font-bold">システムステータス</h1>
      {error && <p className="rounded bg-red-100 p-2 text-sm text-red-600">{error}</p>}

      <Card>
        <CardHeader>
          <CardTitle>スケジューラの状態</CardTitle>
        </CardHeader>
        <CardContent>
          {status ? (
            <ul className="space-y-2 text-sm text-slate-600">
              <li>最終実行: {status.lastRunAt ? new Date(status.lastRunAt).toLocaleString('ja-JP') : '未実行'}</li>
              <li>更新日時: {status.updatedAt ? new Date(status.updatedAt).toLocaleString('ja-JP') : '-'}</li>
              <li className={status.lastError ? 'text-red-600' : 'text-emerald-600'}>
                最終結果: {status.lastError ? status.lastError : '正常終了'}
              </li>
            </ul>
          ) : (
            <div className="space-y-2">
              <Skeleton className="h-4 w-full" />
              <Skeleton className="h-4 w-3/4" />
              <Skeleton className="h-4 w-1/2" />
            </div>
          )}
        </CardContent>
      </Card>

        <Card>
          <CardHeader>
            <CardTitle>重要ログ</CardTitle>
          </CardHeader>
          <CardContent>
            <LogViewer />
          </CardContent>
        </Card>
      </main>
    </>
  );
}
