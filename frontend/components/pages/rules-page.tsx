'use client';

import React, { useEffect, useState, useMemo, useCallback } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { Rule, GuildPermission } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { checkAuth } from '@/lib/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select } from '@/components/ui/select';
import { RuleSkeleton, Skeleton } from '@/components/ui/skeleton';
import { Header } from '@/components/features/header';

const NOTIFY_LABELS: Record<string, string> = {
  open: '新規公開',
  start: '申込開始',
  almost_full: '残席わずか',
  before_deadline: '締切前'
};

export default function RulesPage() {
  const router = useRouter();
  const [guilds, setGuilds] = useState<GuildPermission[]>([]);
  const [selectedGuild, setSelectedGuild] = useState<string>('');
  const [rules, setRules] = useState<Rule[]>([]);
  const [guildsLoading, setGuildsLoading] = useState(true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [authChecked, setAuthChecked] = useState(false);
  const [deleting, setDeleting] = useState<number | null>(null);

  const handleDelete = useCallback(async (ruleId: number, ruleName: string) => {
    if (!window.confirm(`「${ruleName}」を削除しますか？この操作は取り消せません。`)) {
      return;
    }

    try {
      setDeleting(ruleId);
      await apiClient.delete(`/rules/${ruleId}`);
      setRules((prev) => prev.filter((r) => r.id !== ruleId));
    } catch (err) {
      console.error(err);
      setError('ルールの削除に失敗しました');
    } finally {
      setDeleting(null);
    }
  }, []);

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

    const init = async () => {
      try {
        setGuildsLoading(true);
        const { data } = await apiClient.get('/me/guilds');
        setGuilds(data);
        if (data.length > 0) {
          setSelectedGuild(data[0].guildId);
        }
      } catch (err) {
        console.error(err);
        setError('管理可能なサーバーの取得に失敗しました');
      } finally {
        setGuildsLoading(false);
      }
    };
    void init();
  }, [authChecked]);

  useEffect(() => {
    if (!selectedGuild) {
      setRules([]);
      return;
    }
    const fetchRules = async () => {
      try {
        setLoading(true);
        const { data } = await apiClient.get('/rules', { params: { guild_id: selectedGuild } });
        setRules(data);
        setError(null);
      } catch (err) {
        console.error(err);
        setError('ルールの取得に失敗しました');
      } finally {
        setLoading(false);
      }
    };
    void fetchRules();
  }, [selectedGuild]);

  const notifyTypeLabel = useCallback((type: string) => {
    const label = NOTIFY_LABELS[type as keyof typeof NOTIFY_LABELS];
    return label || type;
  }, []);

  const canCreateRule = useMemo(() => guilds.length > 0 && !guildsLoading, [guilds.length, guildsLoading]);

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
            <h1 className="text-2xl font-bold">通知ルール</h1>
            <p className="text-sm text-slate-600">サーバーごとの通知ルールを管理します。</p>
          </div>
          <Button onClick={() => router.push('/rules/new')} disabled={!canCreateRule}>
            新規作成
          </Button>
        </div>

      <section className="rounded-lg border border-slate-200 bg-white p-4 shadow-sm">
        {guildsLoading ? (
          <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="space-y-2">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-3 w-64" />
            </div>
            <div className="w-full md:w-72">
              <Skeleton className="h-10 w-full" />
            </div>
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
              <div className="space-y-1">
                <p className="text-sm font-medium text-slate-600">サーバー選択</p>
                <p className="text-xs text-slate-500">
                  管理したいサーバーを選ぶと、該当サーバーの通知ルールが表示されます。
                </p>
              </div>
              <div className="relative w-full md:w-72">
                <Select
                  value={selectedGuild}
                  onChange={(e) => setSelectedGuild(e.target.value)}
                  disabled={guilds.length === 0}
                  className={loading ? 'opacity-50' : ''}
                >
                  <option value="">サーバーを選択してください</option>
                  {guilds.map((guild) => (
                    <option key={guild.guildId} value={guild.guildId}>
                      {guild.guildName}
                    </option>
                  ))}
                </Select>
                {loading && (
                  <div className="pointer-events-none absolute right-10 top-1/2 -translate-y-1/2">
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600"></div>
                  </div>
                )}
              </div>
            </div>
            {guilds.length === 0 && (
              <p className="mt-3 rounded bg-slate-100 px-3 py-2 text-sm text-slate-600">
                管理可能なサーバーが存在しません。DiscordでBotにサーバー管理権限を付与してください。
              </p>
            )}
          </>
        )}
      </section>

      {error && <p className="rounded bg-red-100 p-2 text-sm text-red-600">{error}</p>}

      {loading ? (
        <div className="grid gap-4">
          <RuleSkeleton />
          <RuleSkeleton />
          <RuleSkeleton />
        </div>
      ) : (
        <div className="grid gap-4">
          {!selectedGuild && !guildsLoading && (
            <p className="text-sm text-slate-500">ルールを表示するには、サーバーを選択してください。</p>
          )}
          {selectedGuild && rules.length === 0 && (
            <p className="text-sm text-slate-500">このサーバーのルールはまだありません。</p>
          )}
          {rules.map((rule) => (
            <Card key={rule.id}>
              <CardHeader>
                <CardTitle>{rule.name}</CardTitle>
                <div className="flex items-center gap-3">
                  <Link className="text-sm text-primary" href={`/rules/${rule.id}/edit`}>
                    編集する
                  </Link>
                  <button
                    onClick={() => handleDelete(rule.id, rule.name)}
                    disabled={deleting === rule.id}
                    className="text-sm text-red-600 hover:text-red-800 disabled:opacity-50"
                  >
                    {deleting === rule.id ? '削除中...' : '削除'}
                  </button>
                </div>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-slate-600">{rule.description || '説明は未設定です。'}</p>
                <div className="mt-3 flex flex-wrap items-center gap-2 text-xs">
                  <span
                    className={`inline-flex items-center rounded-full px-2 py-1 font-medium ${
                      rule.isActive ? 'bg-green-100 text-green-700' : 'bg-slate-200 text-slate-600'
                    }`}
                  >
                    {rule.isActive ? '有効' : '無効'}
                  </span>
                  <span className="inline-flex items-center rounded-full bg-slate-100 px-2 py-1 text-slate-600">
                    #{rule.channelName}
                  </span>
                </div>
                <dl className="mt-4 grid gap-2 text-xs text-slate-500 md:grid-cols-2">
                  <div>
                    <dt className="font-medium text-slate-600">キーワード</dt>
                    <dd>{rule.keywords.length > 0 ? rule.keywords.join(', ') : '指定なし'}</dd>
                  </div>
                  <div>
                    <dt className="font-medium text-slate-600">通知タイミング</dt>
                    <dd>
                      {rule.notifyTypes.length > 0
                        ? rule.notifyTypes.map((type) => notifyTypeLabel(type)).join(', ')
                        : '指定なし'}
                    </dd>
                  </div>
                  <div>
                    <dt className="font-medium text-slate-600">開催地域</dt>
                    <dd>{rule.location || '指定なし'}</dd>
                  </div>
                  <div>
                    <dt className="font-medium text-slate-600">最終更新</dt>
                    <dd>{new Date(rule.updatedAt).toLocaleString('ja-JP')}</dd>
                  </div>
                </dl>
              </CardContent>
            </Card>
          ))}
          </div>
        )}
      </main>
    </>
  );
}
