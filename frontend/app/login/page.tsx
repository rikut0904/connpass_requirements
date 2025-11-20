'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { exchangeCode, initiateLogin, checkAuth } from '@/lib/auth';

export default function LoginPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [checking, setChecking] = useState(true);
  const hasHandledCode = useRef(false);

  // ログイン済みチェック
  useEffect(() => {
    const search = typeof window !== 'undefined' ? window.location.search : '';
    const code = search ? new URLSearchParams(search).get('code') : null;

    // codeがある場合はOAuthコールバック処理を優先
    if (code) {
      setChecking(false);
      return;
    }

    const check = async () => {
      const isLoggedIn = await checkAuth();
      if (isLoggedIn) {
        router.replace('/dashboard');
      } else {
        setChecking(false);
      }
    };
    void check();
  }, [router]);

  useEffect(() => {
    if (hasHandledCode.current || checking) {
      return;
    }

    const search = typeof window !== 'undefined' ? window.location.search : '';
    const code = search ? new URLSearchParams(search).get('code') : null;
    if (!code) {
      return;
    }

    hasHandledCode.current = true;
    let cancelled = false;

    const run = async () => {
      setLoading(true);
      try {
        await exchangeCode(code);
        if (!cancelled) {
          router.replace('/dashboard');
        }
      } catch (err) {
        console.error(err);
        if (!cancelled) {
          setError('認証処理に失敗しました。権限を確認してもう一度お試しください。');
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };
    void run();

    return () => {
      cancelled = true;
    };
  }, [router, checking]);

  const handleLogin = async () => {
    setLoading(true);
    setError(null);
    try {
      await initiateLogin();
    } catch (err) {
      console.error(err);
      setError('認証の開始に失敗しました。時間をおいて再度お試しください。');
      setLoading(false);
    }
  };

  const handleBack = () => {
    router.push('/');
  };

  if (checking) {
    return (
      <main className="mx-auto max-w-md space-y-6 rounded-xl bg-white p-8 shadow-sm">
        <p className="text-center text-slate-600">読み込み中...</p>
      </main>
    );
  }

  return (
    <main className="mx-auto max-w-md space-y-6 rounded-xl bg-white p-8 shadow-sm">
      <h1 className="text-2xl font-bold">Discordでログイン</h1>
      <p className="text-sm text-slate-600">
        Discordアカウントでログインすると所属サーバーの通知ルールを管理できます。
      </p>
      {error && <p className="rounded bg-red-100 p-2 text-sm text-red-700">{error}</p>}
      <div className="flex gap-3">
        <button
          className="flex-1 rounded bg-primary px-4 py-2 font-semibold text-white disabled:opacity-70"
          onClick={handleLogin}
          disabled={loading}
        >
          {loading ? 'リダイレクト中...' : 'Discordで認証'}
        </button>
        <button className="rounded border px-4 py-2" onClick={handleBack}>
          戻る
        </button>
      </div>
    </main>
  );
}
