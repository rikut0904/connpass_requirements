'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { RuleForm, RuleFormValues } from '@/components/features/rule-form';
import { GuildPermission } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';

export default function NewRulePage() {
  const router = useRouter();
  const [guilds, setGuilds] = useState<GuildPermission[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const run = async () => {
      try {
        const { data } = await apiClient.get('/me/guilds');
        setGuilds(data);
      } finally {
        setLoading(false);
      }
    };
    void run();
  }, []);

  const handleSubmit = async (values: RuleFormValues) => {
    await apiClient.post('/rules', {
      guildId: values.guildId,
      channelId: values.channelId,
      channelName: values.channelName,
      name: values.name,
      description: values.description,
      location: values.location,
      capacityThreshold: values.capacityThreshold,
      keywords: values.keywords,
      notifyTypes: values.notifyTypes,
      isActive: values.isActive
    });
    router.push('/rules');
  };

  if (loading) {
    return (
      <main className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <div className="space-y-4 rounded-lg border border-slate-200 bg-white p-6">
          <div className="flex items-center justify-center py-8">
            <div className="text-center">
              <div className="mx-auto mb-4 h-8 w-8 animate-spin rounded-full border-4 border-slate-200 border-t-slate-600"></div>
              <p className="text-sm text-slate-600">サーバー情報を読み込んでいます...</p>
            </div>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="space-y-6">
      <h1 className="text-2xl font-bold">ルールを作成</h1>
      <RuleForm guilds={guilds} onSubmit={handleSubmit} />
    </main>
  );
}
