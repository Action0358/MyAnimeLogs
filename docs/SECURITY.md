# セキュリティガイドライン

## 画像URLのセキュリティ

### 許可されている画像ホスティングサービス

ユーザーアイコンには、以下の信頼できる画像ホスティングサービスのみが使用できます：

#### 一般的な画像ホスティング
- **Imgur**: `https://i.imgur.com/*`, `https://imgur.com/*`
- **Cloudinary**: `https://res.cloudinary.com/*`, `https://*.cloudinary.com/*`

#### 開発者向けサービス
- **GitHub**: `https://avatars.githubusercontent.com/*`, `https://raw.githubusercontent.com/*`
- **Gravatar**: `https://gravatar.com/*`, `https://secure.gravatar.com/*`

#### アニメ関連サービス
- **AniList**: `https://*.anilist.co/*`
- **MyAnimeList**: `https://cdn.myanimelist.net/*`
- **Twitter**: `https://pbs.twimg.com/*`, `https://*.twimg.com/*`

### セキュリティ対策

#### 1. フロントエンド
- **Content Security Policy (CSP)**: 信頼できるドメインのみを許可
- **Next.js Image Optimization**: remotePatterns で許可リストを管理

#### 2. バックエンド
- **URLバリデーション**: 
  - HTTPSのみ許可
  - 許可リストに含まれるドメインのみ受け入れ
  - URL長の制限（最大2048文字）
- **エラーメッセージ**: ユーザーに適切なフィードバックを提供

### 実装詳細

#### バックエンドバリデーション
```go
// backend/internal/utils/validation.go
func ValidateImageURL(imageURL string) error {
    // 1. HTTPSのみ許可
    // 2. 許可リストのドメインチェック
    // 3. 適切なエラーメッセージ
}
```

#### フロントエンド設定
```javascript
// frontend/next.config.js
images: {
  remotePatterns: [
    { protocol: 'https', hostname: 'i.imgur.com' },
    // ... 他の許可ドメイン
  ]
}
```

### リスクと対策

#### 防止されるリスク
- ✅ **SSRF攻撃**: 許可リストにより内部ネットワークへのアクセスを防止
- ✅ **フィッシング**: 悪意のあるサイトへのリンクを防止
- ✅ **XSS攻撃**: CSPにより不正なスクリプト実行を防止
- ✅ **不適切なコンテンツ**: 信頼できるサービスのみを許可

#### 残存するリスク
- ⚠️ 許可されたサービス上の不適切な画像
  - 対策: ユーザー報告機能の実装を検討
  - 対策: 画像モデレーション機能の追加を検討

### 新しいドメインの追加

新しい画像ホスティングサービスを追加する場合：

1. **セキュリティレビュー**
   - サービスの信頼性を確認
   - HTTPSサポートを確認
   - 利用規約を確認

2. **バックエンド更新**
   ```go
   // backend/internal/utils/validation.go
   var allowedImageDomains = []string{
       // ... 既存のドメイン
       "new-trusted-service.com",
   }
   ```

3. **フロントエンド更新**
   ```javascript
   // frontend/next.config.js
   remotePatterns: [
       // ... 既存のパターン
       { protocol: 'https', hostname: 'new-trusted-service.com' },
   ]
   ```

4. **テスト追加**
   ```go
   // backend/internal/utils/validation_test.go
   {
       name: "New Service URL",
       url: "https://new-trusted-service.com/image.png",
       wantErr: false,
   }
   ```

### モニタリング

以下の項目を定期的に監視することを推奨：
- 不正なURL試行のログ
- 画像読み込みエラーの頻度
- ユーザーからの報告

### 参考資料

- [OWASP - Server Side Request Forgery](https://owasp.org/www-community/attacks/Server_Side_Request_Forgery)
- [Content Security Policy Reference](https://content-security-policy.com/)
- [Next.js Image Optimization](https://nextjs.org/docs/app/building-your-application/optimizing/images)
