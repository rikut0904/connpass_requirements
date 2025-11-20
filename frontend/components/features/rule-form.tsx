'use client';

import { FormEvent, useEffect, useRef, useState } from 'react';
import { GuildPermission, Rule } from '@/lib/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select } from '@/components/ui/select';
import { ChannelSelector } from './channel-selector';

const NOTIFY_OPTIONS = [
  { key: 'open', label: '新規公開' },
  { key: 'start', label: '申込開始' },
  { key: 'almost_full', label: '残席わずか' },
  { key: 'before_deadline', label: '締切前' }
];

export type RuleFormValues = {
  guildId: string;
  channelId: string;
  channelName: string;
  name: string;
  description: string;
  location: string;
  capacityThreshold: number;
  keywords: string[];
  notifyTypes: string[];
  isActive: boolean;
};

type Props = {
  guilds: GuildPermission[];
  initialValue?: Partial<Rule>;
  onSubmit: (values: RuleFormValues) => Promise<void>;
};

export function RuleForm({ guilds, initialValue, onSubmit }: Props) {
  const [form, setForm] = useState<RuleFormValues>({
    guildId: initialValue?.guildId ?? '',
    channelId: initialValue?.channelId ?? '',
    channelName: initialValue?.channelName ?? '',
    name: initialValue?.name ?? '',
    description: initialValue?.description ?? '',
    location: initialValue?.location ?? '',
    capacityThreshold: initialValue?.capacityThreshold ?? 80,
    keywords: initialValue?.keywords ?? [],
    notifyTypes: initialValue?.notifyTypes ?? ['open'],
    isActive: initialValue?.isActive ?? true
  });
  const keywordsRef = useRef<HTMLTextAreaElement | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!initialValue) {
      if (keywordsRef.current) {
        keywordsRef.current.value = '';
      }
      return;
    }
    setForm({
      guildId: initialValue.guildId ?? '',
      channelId: initialValue.channelId ?? '',
      channelName: initialValue.channelName ?? '',
      name: initialValue.name ?? '',
      description: initialValue.description ?? '',
      location: initialValue.location ?? '',
      capacityThreshold: initialValue.capacityThreshold ?? 80,
      keywords: initialValue.keywords ?? [],
      notifyTypes: initialValue.notifyTypes ?? ['open'],
      isActive: initialValue.isActive ?? true
    });
    if (keywordsRef.current) {
      keywordsRef.current.value = initialValue.keywords?.join('\n') ?? '';
    }
  }, [initialValue]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setLoading(true);
    setError(null);
    try {
      if (!form.channelId) {
        setError('通知チャンネルを選択してください');
        setLoading(false);
        return;
      }
      const rawKeywords = keywordsRef.current?.value ?? '';
      const parsedKeywords = rawKeywords
        .split(/\r?\n/)
        .map((keyword) => keyword.trim())
        .filter((keyword) => keyword.length > 0);
      setForm((prev) => ({ ...prev, keywords: parsedKeywords }));
      await onSubmit({
        ...form,
        keywords: parsedKeywords
      });
    } catch (err) {
      console.error(err);
      setError('ルールの保存に失敗しました');
    } finally {
      setLoading(false);
    }
  };

  const updateField = <K extends keyof RuleFormValues>(key: K, value: RuleFormValues[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
  };

  const toggleNotifyType = (key: string) => {
    setForm((prev) => {
      const exists = prev.notifyTypes.includes(key);
      return {
        ...prev,
        notifyTypes: exists ? prev.notifyTypes.filter((n) => n !== key) : [...prev.notifyTypes, key]
      };
    });
  };

  return (
    <form className="space-y-6" onSubmit={handleSubmit}>
      {error && <p className="rounded bg-red-100 p-2 text-sm text-red-700">{error}</p>}
      <div className="grid gap-4 md:grid-cols-2">
        <div>
          <label className="text-sm font-medium text-slate-600">対象ギルド</label>
          <Select
            value={form.guildId}
            onChange={(e) =>
              setForm((prev) => ({
                ...prev,
                guildId: e.target.value,
                channelId: '',
                channelName: ''
              }))
            }
            required
          >
            <option value="" disabled>
              選択してください
            </option>
            {guilds.map((guild) => (
              <option key={guild.guildId} value={guild.guildId}>
                {guild.guildName}
              </option>
            ))}
          </Select>
        </div>
        <div className="space-y-2">
          <ChannelSelector
            guildId={form.guildId}
            value={form.channelId ? { id: form.channelId, name: form.channelName } : null}
            onChange={(channel) =>
              setForm((prev) => ({
                ...prev,
                channelId: channel?.id ?? '',
                channelName: channel?.name ?? ''
              }))
            }
          />
        </div>
        <div>
          <label className="text-sm font-medium text-slate-600">開催地域 (任意)</label>
          <Input value={form.location} onChange={(e) => updateField('location', e.target.value)} />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2">
        <div>
          <label className="text-sm font-medium text-slate-600">ルール名</label>
          <Input value={form.name} onChange={(e) => updateField('name', e.target.value)} required />
        </div>
        <div>
          <label className="text-sm font-medium text-slate-600">残席閾値 (%)</label>
          <Input
            type="number"
            min={10}
            max={100}
            value={form.capacityThreshold}
            onChange={(e) => updateField('capacityThreshold', Number(e.target.value))}
          />
        </div>
      </div>
      <div>
        <label className="text-sm font-medium text-slate-600">説明</label>
        <textarea
          className="w-full rounded border border-slate-300 px-3 py-2 text-sm"
          rows={3}
          value={form.description}
          onChange={(e) => updateField('description', e.target.value)}
        />
      </div>
      <div>
        <label className="text-sm font-medium text-slate-600">キーワード (改行区切り)</label>
        <textarea
          ref={keywordsRef}
          className="w-full rounded border border-slate-300 px-3 py-2 text-sm"
          rows={3}
          defaultValue={initialValue?.keywords?.join('\n') ?? ''}
        />
      </div>
      <div>
        <p className="text-sm font-medium text-slate-600">通知タイミング</p>
        <div className="mt-2 flex flex-wrap gap-2">
          {NOTIFY_OPTIONS.map((option) => (
            <button
              key={option.key}
              type="button"
              className={`rounded-full border px-3 py-1 text-sm ${
                form.notifyTypes.includes(option.key) ? 'border-primary bg-primary/10 text-primary' : 'border-slate-300 text-slate-600'
              }`}
              onClick={() => toggleNotifyType(option.key)}
            >
              {option.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-2">
        <input
          id="isActive"
          type="checkbox"
          checked={form.isActive}
          onChange={(e) => updateField('isActive', e.target.checked)}
        />
        <label htmlFor="isActive" className="text-sm text-slate-600">
          ルールを有効化する
        </label>
      </div>
      <div className="flex justify-end gap-3">
        <Button type="submit" disabled={loading}>
          {loading ? '保存中...' : '保存する'}
        </Button>
      </div>
    </form>
  );
}
