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

// MaxRelatedSeasons - 詳細ページに返す関連シーズン（同シリーズの他作品）の最大件数
//
// 48件あれば実データ8,101作品のうち8,100作品を完全に表示できる
// （超えるのは「それいけ！アンパンマン」の58件のみ）。
// クエリは放送時期の昇順で並べてから打ち切るため、上限が低いと
// 常に新しい作品から落ちる。12件だった頃は名探偵コナン（36作品）で
// 1997〜2008年頃の劇場版しか出なかった。
//
// 1件あたりはid・タイトル・放送時期・画像URLのみでペイロードは小さい。
// 画面が縦に伸びないよう、初期表示を絞るのはフロントエンド側の責務とする。
const MaxRelatedSeasons = 48

// MinRelatedTitlePrefixLength - 関連作品をタイトルの前方一致で判定する際、
// 一致対象双方に要求する最小文字数（文字数、バイト数ではない）。
// 短すぎるタイトル同士（例: 1〜3文字）の前方一致は無関係な作品を大量に
// 巻き込むため、実データ計測（サンプル調査）に基づきこの閾値で除外する。
//
// 「アリス」を2文字まで許すと「アリス・ギア・アイギス」「アリスSOS」
// 「アリスと蔵六」「アリス探偵局」を巻き込むことを実データで確認しているため、
// 前方一致についてはこの閾値を下げない。
const MinRelatedTitlePrefixLength = 4

// MinRelatedTitleHeadLength - 関連作品をタイトルの先頭語一致で判定する際、
// 先頭語に要求する最小文字数（文字数、バイト数ではない）。
//
// 前方一致より短くてよいのは、先頭語一致が「先頭語**全体**の一致」を要求し、
// 本質的に safer なため。「アリス」で照合しても、先頭語が「アリス探偵局(第1期)」
// 「アリスと蔵六」である作品には一致しない。
//
// この閾値が4だった頃は「銀魂」（2文字）が全13作品で関連0件だった（#837）。
// 2まで下げた場合に新たにできる先頭語グループを実データで全数目視し、
// 誤検知は「アニメ」1件のみだったため、それは relatedTitleHeadExpr の
// 除去対象（総称語）に加えて解消している。
//
// 先頭語一致は非ASCII（日本語を含む）先頭語にのみ適用されるため、
// 英語の短い単語がここに乗ることはない。
const MinRelatedTitleHeadLength = 2

// String returns the string representation of AnimeStatus
func (s AnimeStatus) String() string {
	return string(s)
}
