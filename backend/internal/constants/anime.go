package constants

import "time"

// 季節の定数
const (
	// SeasonSpring - 春
	SeasonSpring = "spring"
	// SeasonSummer - 夏
	SeasonSummer = "summer"
	// SeasonAutumn - 秋
	SeasonAutumn = "autumn"
	// SeasonWinter - 冬
	SeasonWinter = "winter"
)

// デフォルト値の定数

// DefaultPopularSearches - デフォルト人気検索キーワード
// エラー時や検索ログが少ない場合に使用
var DefaultPopularSearches = []string{"進撃の巨人", "鬼滅の刃", "呪術廻戦", "ワンピース", "ナルト"}

// DefaultSeasons - デフォルトシーズン一覧
// エラー時に使用
var DefaultSeasons = []string{SeasonSpring, SeasonSummer, SeasonAutumn, SeasonWinter}

// SyncStatus - 同期ステータスの型
type SyncStatus string

const (
	// SyncStatusPending - 同期待ち
	SyncStatusPending SyncStatus = "pending"
	// SyncStatusSynced - 同期済み
	SyncStatusSynced SyncStatus = "synced"
	// SyncStatusFailed - 同期失敗
	SyncStatusFailed SyncStatus = "failed"
)

// StaleDataThreshold - データが古いと判定する閾値（30日）
const StaleDataThreshold = 30 * 24 * time.Hour

// AnimeStatus - アニメ放送ステータスの型
type AnimeStatus string

const (
	// AnimeStatusAiring - 放送中
	AnimeStatusAiring AnimeStatus = "airing"
	// AnimeStatusFinished - 完結
	AnimeStatusFinished AnimeStatus = "finished"
	// AnimeStatusUpcoming - 放送予定
	AnimeStatusUpcoming AnimeStatus = "upcoming"
)

// MaxRelatedSeasons - 詳細ページに表示する関連シーズン（同シリーズの他作品）の最大件数
const MaxRelatedSeasons = 12

// MinRelatedTitlePrefixLength - 関連作品をタイトルの前方一致で判定する際、
// 一致対象双方に要求する最小文字数（文字数、バイト数ではない）。
// 短すぎるタイトル同士（例: 1〜3文字）の前方一致は無関係な作品を大量に
// 巻き込むため、実データ計測（サンプル調査）に基づきこの閾値で除外する。
const MinRelatedTitlePrefixLength = 4

// String returns the string representation of AnimeStatus
func (s AnimeStatus) String() string {
	return string(s)
}
