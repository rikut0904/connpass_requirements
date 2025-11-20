'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api';
import { Button } from '@/components/ui/button';

export function TestNotification({ ruleId }: { ruleId: number }) {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  const handleTest = async () => {
    setLoading(true);
    setResult(null);
    try {
      const { data } = await apiClient.post(`/rules/${ruleId}/test`);
      setResult(data?.message ?? 'テスト通知を送信しました');
    } catch (err) {
      console.error(err);
      setResult('通知の送信に失敗しました。権限とチャンネルIDを確認してください。');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-2">
      <Button onClick={handleTest} disabled={loading} variant="secondary">
        {loading ? '送信中...' : 'テスト通知を送る'}
      </Button>
      {result && <p className="text-sm text-slate-600">{result}</p>}
    </div>
  );
}
