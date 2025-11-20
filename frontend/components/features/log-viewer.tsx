'use client';

import { useEffect, useState, memo } from 'react';
import { ImportantLog } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';

export function LogViewer() {
  const [logs, setLogs] = useState<ImportantLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const run = async () => {
      try {
        const { data } = await apiClient.get('/logs', { params: { limit: 20 } });
        setLogs(data);
      } catch (err) {
        console.error(err);
        setError('ログの取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };
    void run();
  }, []);

  if (loading) {
    return (
      <div className="space-y-2">
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
        <Skeleton className="h-24 w-full" />
      </div>
    );
  }
  if (error) {
    return <p className="text-sm text-red-600">{error}</p>;
  }

  return (
    <div className="space-y-2 text-sm">
      {logs.length === 0 && <p className="text-slate-500">表示するログはありません。</p>}
      {logs.map((log) => (
        <LogItem key={log.id} log={log} />
      ))}
    </div>
  );
}

const LogItem = memo(({ log }: { log: ImportantLog }) => (
  <div className="rounded border border-slate-200 p-3">
    <div className="flex items-center justify-between text-xs uppercase tracking-wide text-slate-500">
      <span>{log.level}</span>
      <span>{new Date(log.createdAt).toLocaleString('ja-JP')}</span>
    </div>
    <p className="mt-1 font-medium text-slate-700">{log.message}</p>
    <p className="text-xs text-slate-500">イベント: {log.eventType}</p>
    {log.metadata && <pre className="mt-2 overflow-x-auto rounded bg-slate-100 p-2 text-xs">{log.metadata}</pre>}
  </div>
));
