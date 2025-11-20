'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { SchedulerStatus } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { checkAuth } from '@/lib/auth';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { LogViewer } from '@/components/features/log-viewer';
import { Skeleton } from '@/components/ui/skeleton';
import { Header } from '@/components/features/header';

export default function DashboardPage() {
  const router = useRouter();
  const [status, setStatus] = useState<SchedulerStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [runSuccess, setRunSuccess] = useState<string | null>(null);
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

  const handleRunScheduler = async () => {
    setIsRunning(true);
    setError(null);
    setRunSuccess(null);
    try {
      await apiClient.post('/scheduler/run');
      setRunSuccess('スケジューラーを実行しました。数秒後にログとステータスが更新されます。');
      // ステータスを再取得
      setTimeout(async () => {
        try {
          const { data } = await apiClient.get('/status');
          setStatus(data);
        } catch (err) {
          console.error(err);
        }
      }, 2000);
    } catch (err: any) {
      console.error(err);
      setError(err.response?.data?.message || 'スケジューラーの実行に失敗しました');
    } finally {
      setIsRunning(false);
    }
  };

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
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">ダッシュボード</h1>
          <p className="text-sm text-slate-600">通知ルールの状況と最新ログを確認できます。</p>
        </div>
        <div className="flex gap-2">
          <Button variant="secondary" onClick={handleRunScheduler} disabled={isRunning}>
            {isRunning ? '実行中...' : '今すぐ通知チェック'}
          </Button>
          <Button onClick={() => router.push('/rules/new')}>ルールを追加</Button>
        </div>
      </div>

      {runSuccess && (
        <div className="rounded bg-green-100 p-3 text-sm text-green-700">
          {runSuccess}
        </div>
      )}

      {error && (
        <div className="rounded bg-red-100 p-3 text-sm text-red-600">
          {error}
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>スケジューラ実行</CardTitle>
          </CardHeader>
          <CardContent>
            {status ? (
              <div className="space-y-2 text-sm">
                <p>最終実行: {status.lastRunAt ? new Date(status.lastRunAt).toLocaleString('ja-JP') : '未実行'}</p>
                <p>最終更新: {status.updatedAt ? new Date(status.updatedAt).toLocaleString('ja-JP') : '-'}</p>
                <p className={status.lastError ? 'text-red-600' : 'text-emerald-600'}>
                  状態: {status.lastError ? 'エラーあり' : '正常'}
                </p>
                {status.lastError && <p className="text-xs text-red-600">{status.lastError}</p>}
              </div>
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
            <CardTitle>クイック操作</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-sm">
              <Button variant="secondary" className="w-full" onClick={() => router.push('/rules')}>
                ルール一覧へ
              </Button>
              <Button variant="secondary" className="w-full" onClick={() => router.push('/status')}>
                ステータス詳細
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>ヘルプ</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="list-disc space-y-1 pl-5 text-sm">
              <li>Botがサーバーに参加しているか確認してください。</li>
              <li>通知チャンネルに投稿権限があることを確認してください。</li>
              <li>問題が続く場合はログを確認し、Railwayのログも参照してください。</li>
            </ul>
          </CardContent>
        </Card>
      </div>

        <Card>
          <CardHeader>
            <CardTitle>最新ログ</CardTitle>
          </CardHeader>
          <CardContent>
            <LogViewer />
          </CardContent>
        </Card>
      </main>
    </>
  );
}
