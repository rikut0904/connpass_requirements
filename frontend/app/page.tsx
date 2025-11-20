'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { checkAuth } from '@/lib/auth';

export default function LandingPage() {
  const router = useRouter();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
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

  if (checking) {
    return (
      <main className="flex items-center justify-center p-10">
        <p className="text-slate-600">読み込み中...</p>
      </main>
    );
  }

  return (
    <main className="space-y-10">
      <section className="rounded-xl bg-white p-10 shadow-sm">
        <h1 className="text-3xl font-bold">Connpass Discord 通知システム</h1>
        <p className="mt-4 text-slate-600">
          connpassイベントを監視し、条件に合致したイベントをDiscordへ自動通知します。
        </p>
        <div className="mt-6 flex gap-4">
          <Link
            className="rounded bg-primary px-4 py-2 font-semibold text-white hover:bg-primary/90"
            href="/login"
          >
            Discordでログイン
          </Link>
          <Link className="rounded border px-4 py-2" href="/status">
            ステータスを見る
          </Link>
        </div>
      </section>
      <section className="grid gap-6 md:grid-cols-2">
        <FeatureCard
          title="通知ルールを柔軟に設定"
          description="キーワード・地域・通知タイミングを組み合わせて、欲しい情報だけを受け取れます。"
        />
        <FeatureCard
          title="30分ごとの自動クロール"
          description="スケジューラが定期的にconnpass APIを監視し、重複通知も防止します。"
        />
        <FeatureCard
          title="Discord OAuth2対応"
          description="所属サーバーと権限をチェックし、安全にルールを管理できます。"
        />
        <FeatureCard
          title="重要ログを可視化"
          description="障害時は重要ログとステータス画面で迅速に状況を把握できます。"
        />
      </section>
    </main>
  );
}

function FeatureCard({
  title,
  description
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="rounded-xl bg-white p-6 shadow-sm">
      <h2 className="text-xl font-semibold">{title}</h2>
      <p className="mt-2 text-sm text-slate-600">{description}</p>
    </div>
  );
}
