'use client';

import { apiClient } from './api';

const clientId = process.env.NEXT_PUBLIC_DISCORD_CLIENT_ID ?? '';
const redirectUri = process.env.NEXT_PUBLIC_DISCORD_REDIRECT_URI ?? '';
const oauthBase = process.env.NEXT_PUBLIC_DISCORD_OAUTH_URL ?? 'https://discord.com/oauth2/authorize';

export function initiateLogin() {
  if (!clientId || !redirectUri) {
    throw new Error('Discord OAuthの環境変数が設定されていません');
  }

  const params = new URLSearchParams({
    client_id: clientId,
    redirect_uri: redirectUri,
    response_type: 'code',
    scope: 'identify guilds'
  });

  window.location.href = `${oauthBase}?${params.toString()}`;
}

export async function exchangeCode(code: string) {
  const { data } = await apiClient.post('/auth/callback', { code });
  return data;
}

export function parseOAuthCode(url: string) {
  const u = new URL(url);
  return u.searchParams.get('code');
}

export async function checkAuth(): Promise<boolean> {
  try {
    const { apiClient } = await import('./api');
    await apiClient.get('/auth/me');
    return true;
  } catch {
    return false;
  }
}

export async function logout(): Promise<void> {
  await apiClient.post('/auth/logout');
}
