import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Connpass Discord Notifier',
  description: 'connpassイベントをDiscordに自動通知する管理ダッシュボード'
};

export default function RootLayout({
  children
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body className="min-h-screen antialiased">
        <div className="mx-auto max-w-6xl px-4 py-8">{children}</div>
      </body>
    </html>
  );
}
