import axios from 'axios';
import { ChangeEvent, useEffect, useMemo, useState } from 'react';
import { apiClient } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Select } from '@/components/ui/select';
import { Input } from '@/components/ui/input';

type ChannelOption = {
  id: string;
  name: string;
  categoryId?: string;
  categoryName?: string;
};

type CategoryOption = {
  id: string;
  name: string;
};

type Props = {
  guildId: string;
  value: ChannelOption | null;
  onChange: (channel: ChannelOption | null) => void;
};

const NEW_CATEGORY_VALUE = '__new__';

export function ChannelSelector({ guildId, value, onChange }: Props) {
  const [channels, setChannels] = useState<ChannelOption[]>([]);
  const [categories, setCategories] = useState<CategoryOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [newChannelName, setNewChannelName] = useState('');
  const [selectedCategoryId, setSelectedCategoryId] = useState('');
  const [newCategoryName, setNewCategoryName] = useState('');
  const [error, setError] = useState<string | null>(null);

  const currentValue = useMemo(() => value?.id ?? '', [value?.id]);

  const selectedChannel = useMemo(() => {
    if (!value?.id) {
      return null;
    }
    return channels.find((channel) => channel.id === value.id) ?? value;
  }, [channels, value]);

  useEffect(() => {
    setError(null);
    if (!guildId) {
      setChannels([]);
      setCategories([]);
      setShowCreateForm(false);
      setNewChannelName('');
      setSelectedCategoryId('');
      setNewCategoryName('');
      return;
    }
    const run = async () => {
      try {
        setLoading(true);
        const { data } = await apiClient.get(`/guilds/${guildId}/channels`);
        const fetchedCategories: CategoryOption[] = (data.categories ?? []).map((category: any) => ({
          id: category.id,
          name: category.name
        }));
        const fetchedChannels: ChannelOption[] = (data.channels ?? []).map((channel: any) => ({
          id: channel.id,
          name: channel.name,
          categoryId: channel.categoryId ?? '',
          categoryName: channel.categoryName ?? ''
        }));
        setCategories(fetchedCategories);
        setChannels(() => {
          const merged = [...fetchedChannels];
          if (value && value.id && !merged.some((ch) => ch.id === value.id)) {
            merged.push(value);
          }
          return merged;
        });
      } catch (err) {
        console.warn('チャンネル一覧の取得に失敗しました', err);
        setError('チャンネル一覧の取得に失敗しました。Botの権限を確認してください。');
        setCategories([]);
        setChannels(value ? [value] : []);
      } finally {
        setLoading(false);
      }
    };
    void run();
  }, [guildId, value?.id, value?.name, value?.categoryId, value?.categoryName]);

  const handleSelect = (event: ChangeEvent<HTMLSelectElement>) => {
    const selectedId = event.target.value;
    if (!selectedId) {
      onChange(null);
      return;
    }
    const selected = channels.find((channel) => channel.id === selectedId);
    if (selected) {
      onChange(selected);
    } else {
      onChange({ id: selectedId, name: '' });
    }
  };

  const handleToggleCreateForm = () => {
    if (showCreateForm) {
      setShowCreateForm(false);
      setNewChannelName('');
      setSelectedCategoryId('');
      setNewCategoryName('');
    } else {
      setError(null);
      setShowCreateForm(true);
      setNewChannelName('');
      setSelectedCategoryId('');
      setNewCategoryName('');
    }
  };

  const handleCategorySelect = (event: ChangeEvent<HTMLSelectElement>) => {
    const category = event.target.value;
    setSelectedCategoryId(category);
    if (category !== NEW_CATEGORY_VALUE) {
      setNewCategoryName('');
    }
  };

  const submitNewChannel = async () => {
    if (!guildId || creating) {
      return;
    }
    const trimmedName = newChannelName.trim();
    if (!trimmedName) {
      setError('チャンネル名を入力してください。');
      return;
    }
    const isNewCategory = selectedCategoryId === NEW_CATEGORY_VALUE;
    const trimmedCategoryName = newCategoryName.trim();
    if (isNewCategory && !trimmedCategoryName) {
      setError('新しいカテゴリ名を入力してください。');
      return;
    }

    const payload: { name: string; categoryId?: string; categoryName?: string } = { name: trimmedName };
    if (!isNewCategory && selectedCategoryId) {
      payload.categoryId = selectedCategoryId;
    }
    if (isNewCategory) {
      payload.categoryName = trimmedCategoryName;
    }

    setCreating(true);
    setError(null);
    try {
      const { data } = await apiClient.post(`/guilds/${guildId}/channels`, payload);
      const createdCategory: CategoryOption | null = data.category
        ? { id: data.category.id, name: data.category.name }
        : null;

      let nextCategories = categories;
      if (createdCategory && !categories.some((category) => category.id === createdCategory.id)) {
        nextCategories = [...categories, createdCategory];
        setCategories(nextCategories);
      }

      const categoryIdFromResponse: string = data.categoryId ?? createdCategory?.id ?? '';
      let categoryNameFromResponse: string =
        createdCategory?.name ??
        (categoryIdFromResponse ? nextCategories.find((category) => category.id === categoryIdFromResponse)?.name ?? '' : '');

      const newChannel: ChannelOption = {
        id: data.id,
        name: data.name,
        categoryId: categoryIdFromResponse || undefined,
        categoryName: categoryNameFromResponse || undefined
      };

      setChannels((prev) => {
        const index = prev.findIndex((channel) => channel.id === newChannel.id);
        if (index >= 0) {
          const next = [...prev];
          next[index] = newChannel;
          return next;
        }
        return [...prev, newChannel];
      });
      onChange(newChannel);
      setShowCreateForm(false);
      setNewChannelName('');
      setSelectedCategoryId('');
      setNewCategoryName('');
    } catch (err) {
      console.error('チャンネルの作成に失敗しました', err);
      if (axios.isAxiosError(err)) {
        if (err.response?.status === 403) {
          setError('Botにチャンネル作成権限がありません。Discord側の権限を確認してください。');
        } else if (typeof err.response?.data?.message === 'string') {
          setError(err.response.data.message);
        } else if (err.response?.status === 400) {
          setError('入力内容を確認してください。');
        } else {
          setError('チャンネルの作成に失敗しました。時間をおいて再度お試しください。');
        }
      } else {
        setError('チャンネルの作成に失敗しました。時間をおいて再度お試しください。');
      }
    } finally {
      setCreating(false);
    }
  };

  const isNewCategory = selectedCategoryId === NEW_CATEGORY_VALUE;
  const canSubmit =
    newChannelName.trim().length > 0 &&
    (!isNewCategory || newCategoryName.trim().length > 0);

  return (
    <div className="space-y-2">
      <label className="text-sm font-medium text-slate-600">通知チャンネル</label>
      <div className="flex flex-col gap-3">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Select
              value={currentValue}
              onChange={handleSelect}
              disabled={loading || !guildId}
              required
              className={loading ? 'opacity-50' : ''}
            >
              <option value="">{loading ? '読み込み中...' : '選択してください'}</option>
              {channels.map((channel) => (
                <option key={channel.id} value={channel.id}>
                  {channel.categoryName ? `${channel.categoryName} / ` : ''}
                  #{channel.name}
                </option>
              ))}
            </Select>
            {loading && (
              <div className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2">
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-slate-600"></div>
              </div>
            )}
          </div>
          <Button
            type="button"
            variant={showCreateForm ? 'ghost' : 'outline'}
            onClick={handleToggleCreateForm}
            disabled={!guildId || loading || creating}
          >
            {showCreateForm ? 'フォームを閉じる' : '新規作成'}
          </Button>
        </div>
        {showCreateForm && (
          <div className="space-y-3 rounded border border-slate-200 p-3">
            <div className="grid gap-3">
              <div className="space-y-1">
                <label htmlFor="newChannelName" className="text-xs font-medium text-slate-600">
                  チャンネル名
                </label>
                <Input
                  id="newChannelName"
                  value={newChannelName}
                  onChange={(event) => setNewChannelName(event.target.value)}
                  disabled={creating}
                />
              </div>
              <div className="space-y-1">
                <label htmlFor="newChannelCategory" className="text-xs font-medium text-slate-600">
                  カテゴリ
                </label>
                <Select
                  id="newChannelCategory"
                  value={selectedCategoryId}
                  onChange={handleCategorySelect}
                  disabled={creating}
                >
                  <option value="">カテゴリなし</option>
                  {categories.map((category) => (
                    <option key={category.id} value={category.id}>
                      {category.name}
                    </option>
                  ))}
                  <option value={NEW_CATEGORY_VALUE}>+ 新しいカテゴリを作成</option>
                </Select>
              </div>
              {isNewCategory && (
                <div className="space-y-1">
                  <label htmlFor="newCategoryName" className="text-xs font-medium text-slate-600">
                    新規カテゴリ名
                  </label>
                  <Input
                    id="newCategoryName"
                    value={newCategoryName}
                    onChange={(event) => setNewCategoryName(event.target.value)}
                    disabled={creating}
                  />
                </div>
              )}
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="ghost" onClick={handleToggleCreateForm} disabled={creating}>
                キャンセル
              </Button>
              <Button type="button" onClick={submitNewChannel} disabled={creating || !canSubmit}>
                {creating ? '作成中...' : '作成する'}
              </Button>
            </div>
          </div>
        )}
      </div>
      {error && <p className="text-xs text-red-600">{error}</p>}
      {selectedChannel && selectedChannel.id && (
        <p className="text-xs text-slate-500">
          選択中: {selectedChannel.categoryName ? `${selectedChannel.categoryName} / ` : ''}
          #{selectedChannel.name || '（名称未取得）'} （ID: {selectedChannel.id}）
        </p>
      )}
      <p className="text-xs text-slate-500">
        チャンネルAPIはBotの権限が必要です。権限が不足している場合は未表示となります。
      </p>
    </div>
  );
}
