package common

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

// RecoverGoroutine は goroutine の先頭で defer して使い、panic を記録して止める。
//
// HTTP リクエストの panic は Gin の Recovery が 500 にするが、go 文で起動した
// 別の goroutine の panic は対象外で、プロセスごと落ちる。本番のマシンは1台なので、
// 1回の panic で全リクエストが失敗する。それを防ぐため、バックグラウンドの
// goroutine ではこの関数で panic を止める。
//
// recover は defer された関数の中で直接呼ばないと効かないため、必ず
// `defer common.RecoverGoroutine(logger, "name")` の形で呼ぶこと。
// onPanic を渡すと、記録の後に呼ぶ（結果の送信や状態の後始末に使う）。
func RecoverGoroutine(logger *Logger, name string, onPanic ...func(recovered any)) {
	r := recover()
	if r == nil {
		return
	}
	LogPanic(logger, name, r)
	for _, f := range onPanic {
		f(r)
	}
}

// LogPanic は recover した値とスタックトレースをエラーログに残す。
// 自前で recover する箇所（後始末が必要な場合など）から呼ぶ。
func LogPanic(logger *Logger, name string, recovered any) {
	l := slog.Default()
	if logger != nil && logger.Logger != nil {
		l = logger.Logger
	}
	l.Error("Panic recovered in goroutine",
		"goroutine", name,
		"panic", fmt.Sprint(recovered),
		"stack", string(debug.Stack()))
}
