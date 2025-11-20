import { NextResponse } from 'next/server';

export async function GET(request: Request) {
  const url = new URL(request.url);
  const code = url.searchParams.get('code');
  const redirect = new URL('/login', url.origin);
  if (code) {
    redirect.searchParams.set('code', code);
  }
  return NextResponse.redirect(redirect);
}
