package constants

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
