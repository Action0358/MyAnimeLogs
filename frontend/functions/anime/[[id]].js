// Cloudflare Pages Function for SPA fallback
// /anime/:id へのリクエストを /anime/_/index.html にリライト（200ステータス）

export async function onRequest(context) {
  const url = new URL(context.request.url);
  const pathname = url.pathname;
  
  // /anime/_/ へのリクエストはそのまま通す
  if (pathname === '/anime/_' || pathname === '/anime/_/') {
    return context.next();
  }
  
  // /anime/:id を /anime/_/index.html にリライト
  url.pathname = '/anime/_/';
  
  return context.env.ASSETS.fetch(new Request(url, context.request));
}
