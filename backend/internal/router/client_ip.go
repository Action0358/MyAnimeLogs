package router

import "github.com/gin-gonic/gin"

// flyClientIPHeader は Fly.io のプロキシが付けるクライアントIPのヘッダー。
// Fly のプロキシは受け取った値に関係なく自分で上書きするため、利用者は偽装できない。
const flyClientIPHeader = "Fly-Client-IP"

// configureClientIP は c.ClientIP() が返すIPの決め方を設定する。
//
// レート制限はIPアドレスごとに数えるため、IPを利用者が偽装できると制限を回避される。
// Gin の既定は全プロキシを信頼し、利用者が自由に付けられる X-Forwarded-For の
// 左端の値を採用してしまう。そこで:
//   - X-Forwarded-For / X-Real-IP は信頼しない（SetTrustedProxies(nil)）
//   - Fly.io 上（staging・production）だけ、Fly-Client-IP を信頼する
//   - それ以外（ローカルなど）はプロキシを介さないので、接続元のアドレスをそのまま使う
func configureClientIP(r *gin.Engine, environment string) {
	// nil を渡すとプロキシを一切信頼しない設定になる。エラーは返らない
	_ = r.SetTrustedProxies(nil)

	if environment == "production" || environment == "staging" {
		r.TrustedPlatform = flyClientIPHeader
	}
}
