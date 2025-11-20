'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { RuleForm, RuleFormValues } from '@/components/features/rule-form';
import { GuildPermission, Rule } from '@/lib/types';
import { apiClient } from '@/lib/api';
import { Skeleton } from '@/components/ui/skeleton';

export default function EditRulePage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const [rule, setRule] = useState<Rule | null>(null);
  const [guilds, setGuilds] = useState<GuildPermission[]>([]);
  const [loading, setLoading] = useState(true);
  const ruleId = Number(params?.id);

  useEffect(() => {
    const run = async () => {
      try {
        const [{ data: ruleData }, { data: guildData }] = await Promise.all([
          apiClient.get(`/rules/${ruleId}`),
          apiClient.get('/me/guilds')
        ]);
        setRule(ruleData);
        setGuilds(guildData);
      } finally {
        setLoading(false);
      }
    };
    if (!Number.isNaN(ruleId)) {
      void run();
    }
  }, [ruleId]);

  const handleSubmit = async (values: RuleFormValues) => {
    await apiClient.put(`/rules/${ruleId}`, {
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
              <p className="text-sm text-slate-600">ルール情報を読み込んでいます...</p>
            </div>
          </div>
        </div>
      </main>
    );
  }

  if (!rule) {
    return <p className="text-sm text-red-600">ルールが見つかりませんでした。</p>;
  }

  return (
    <main className="space-y-6">
      <h1 className="text-2xl font-bold">ルールを編集</h1>
      <RuleForm guilds={guilds} initialValue={rule} onSubmit={handleSubmit} />
    </main>
  );
}
