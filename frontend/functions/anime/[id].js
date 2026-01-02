// Cloudflare Pages Function
// /anime/:id へのリクエストを /anime/_/index.html にリライト

export async function onRequest(context) {
  const url = new URL(context.request.url);
  
  // /anime/_/ へのリクエストはそのまま通す
  if (url.pathname === '/anime/_' || url.pathname === '/anime/_/') {
    return context.next();
  }
  
  // /anime/:id を /anime/_/index.html にリライト
  url.pathname = '/anime/_/';
  
  return context.env.ASSETS.fetch(new Request(url, context.request));
}
