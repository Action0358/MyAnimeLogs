# メンテナンスモード運用手順

メンテナンス中に全ページへ「メンテナンス中です」画面を出す仕組みの切り替え手順。
経路は2つあり、状況に応じて使い分ける（併用できる）。

| 経路 | 使う場面 | 切り替えの手段 | 反映までの時間 |
|---|---|---|---|
| **A. バックエンドのフラグ** | バックエンドが動いていて、メンテナンスを宣言したいとき（DBマイグレーション・外部API障害など） | Fly の secret | 再起動の間（数十秒〜） |
| **B. `maintenance.json`** | バックエンドが止まる・入れ替わるとき、またはフロント単独で宣言したいとき | `frontend/public/maintenance.json` を変更してフロントをデプロイ | フロントのデプロイ時間（数分） |

どちらかが有効なら、フロントは全ページでメンテナンス画面を出す。ただし **`/news`（お知らせ）は対象外**で、メンテナンス中も開ける（静的生成でAPIに依存しないため）。メンテナンス画面の「お知らせを見る」の行き先になっている。

## A. バックエンドのフラグ

有効にすると、`/health` と `/api/admin/` と CORS のプリフライト（OPTIONS）を除く全リクエストが `503` + `code: "MAINTENANCE"` を返す。`ends_at` と `Retry-After` も付く。

```bash
# 開始（終了予定は任意。RFC3339、JSTなら +09:00）
fly secrets set MAINTENANCE_MODE=true MAINTENANCE_ENDS_AT=2026-10-25T15:00:00+09:00 -a myanimelogs-production

# 終了
fly secrets unset MAINTENANCE_MODE MAINTENANCE_ENDS_AT -a myanimelogs-production
```

- STG は `-a myanimelogs-staging`
- `fly secrets set` / `unset` は再起動を伴う。反映を待ってから確認する
- `MAINTENANCE_ENDS_AT` が不正な形式でも起動は止まらない（終了予定が出ないだけ）。形式を間違えていないか、画面の終了予定表示で確認する
- フロントは API が `503` かつ `MAINTENANCE` を返したときだけメンテナンス画面にする。外部API障害などの `SERVICE_UNAVAILABLE` は対象外（通常のエラー表示のまま）

## B. maintenance.json

```json
{
  "enabled": true,
  "endsAt": "2026-10-25T15:00:00+09:00"
}
```

- `enabled: true` でメンテナンス画面になる。`endsAt` は任意（`null` なら終了予定を出さない）
- 変更をコミットして、通常どおりフロントをデプロイする（develop → staging → main）
- 終了するときは `"enabled": false` に戻して再デプロイする
- ファイルは `Cache-Control: no-store`（`frontend/public/_headers`）で配信され、ブラウザは毎回取得する
- このファイルを取得できない・形式が違う場合は、何もせず通常どおり動く

## お知らせへの告知

メンテナンス画面の「お知らせを見る」は `/news` に行く。そこに知りたい情報がないと行き止まりになるため、**メンテナンスの前に告知記事を足しておく**。

`frontend/src/constants/announcements.ts` の `ANNOUNCEMENTS` に、カテゴリ `maintenance` の記事を足して develop → staging → main とマージする（マージした時点で公開される）。

- タイトル例: `10月25日 メンテナンスのお知らせ`
- `body` の1文目に終了予定を入れる（一覧では2行に省略されるため）。例: `終了予定：2026年10月25日 15:00ごろ。作業中はすべての機能をご利用いただけません。視聴記録・お気に入りは保持されます。`
- カテゴリ `maintenance` のバッジはスレート色（赤・琥珀は使わない）
- 終了後も記事は残してよい（履歴として）

メンテナンス中の `/news` は、ヘッダーがロゴのみ、フッターがコピーライトのみになり、タイトルの上に「‹ 戻る」（メンテナンス画面へ）が出る。

## 動作確認（STG）

1. STG で A または B を有効にする
2. 次を確認する
   - トップ・マイページなど任意のページでメンテナンス画面が出る（ナビ・ログインは出ない）
   - 終了予定を付けたときだけ「終了予定：…ごろ」が出る
   - 「お知らせを見る」から `/news` が開ける
   - 「再読み込み」で画面が再取得される
   - `curl https://myanimelogs-staging.fly.dev/health` が 200（A の場合）
3. 解除して、通常表示に戻ることを確認する

## 注意

- フロントは静的サイトのため、判定が付くまでの一瞬は通常画面が見えることがある（メンテナンス開始直後）
- 一度メンテナンスと判定したタブは、再読み込みするまで解除されない。解除の確認は「再読み込み」で行う
- 管理者の API（`/api/admin/`）はメンテナンス中も使える。同期などの作業を続けられる
- バックエンドが完全に落ちていて B も有効でない場合は、メンテナンス画面ではなく通信エラー画面になる。その画面には「メンテナンス中は、お知らせをご確認ください」と案内が出る
- 画面のデザインは `frontend/src/components/maintenance/MaintenanceScreen.tsx`、判定は `MaintenanceGate.tsx` を参照
