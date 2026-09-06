// Cloudflare Pages Function for SPA fallback
// /anime/:id へのリクエストを /anime/_/index.html にリライト（200ステータス）
// 注意: [id].js は単一セグメント必須のため、/anime/ へのリクエストはマッチしない

/**
 * /anime/ 配下で静的生成されるページ。リライトの対象外にする。
 *
 * この関数は /anime/ 配下の単一セグメントすべてにマッチするため、
 * 除外しないと /anime/popular/ /anime/current/ まで詳細ページのシェルへ
 * リライトされてしまう。詳細ページは URL から作品IDを取り出すので
 * parseInt('popular') が NaN になり、ローディング表示から抜けられない（#842）。
 *
 * サイト内リンクではクライアントサイドルーティングが働きこの関数を通らないため、
 * 直接アクセス・リロード・ブックマーク・SNS共有のときだけ壊れる。
 *
 * src/app/anime/ にページを追加したらここにも追加すること。
 * 追随漏れは functions/anime/id.test.js が検出する。
 */
export const STATIC_ANIME_PATHS = ['popular', 'current'];

/** 詳細ページのシェル（Next.js が出力する /anime/_/index.html） */
const DETAIL_SHELL_SEGMENT = '_';

export async function onRequest(context) {
  const url = new URL(context.request.url);

  // 末尾スラッシュの有無を吸収して /anime/<segment> の <segment> を取り出す
  const segment = url.pathname.replace(/^\/anime\//, '').replace(/\/$/, '');

  // 実体のあるページ（詳細シェル自身と静的ページ）はそのまま通す
  if (segment === DETAIL_SHELL_SEGMENT || STATIC_ANIME_PATHS.includes(segment)) {
    return context.next();
  }

  // それ以外は作品IDとみなし、詳細ページのシェルへリライト
  url.pathname = `/anime/${DETAIL_SHELL_SEGMENT}/`;

  return context.env.ASSETS.fetch(new Request(url, context.request));
}
