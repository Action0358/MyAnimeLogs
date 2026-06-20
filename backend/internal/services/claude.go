package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"myanimelogs/internal/common"
)

// ClaudeClient はあらすじの日本語翻訳・整形に Claude (Anthropic API) を利用するクライアント。
type ClaudeClient struct {
	client  anthropic.Client
	logger  *common.Logger
	enabled bool
}

// NewClaudeClient は ANTHROPIC_API_KEY が設定されていれば有効なクライアントを返す。
// 未設定の場合は enabled=false となり、翻訳はスキップされる（英語あらすじはそのまま使用）。
func NewClaudeClient(logger *common.Logger) *ClaudeClient {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		if logger != nil {
			logger.Warn("ANTHROPIC_API_KEY not set, Claude synopsis translation will be disabled")
		}
		return &ClaudeClient{logger: logger, enabled: false}
	}
	return &ClaudeClient{
		client:  anthropic.NewClient(option.WithAPIKey(apiKey)),
		logger:  logger,
		enabled: true,
	}
}

// IsEnabled は Claude API が利用可能かを返す。
func (c *ClaudeClient) IsEnabled() bool {
	return c.enabled
}

const synopsisTranslationPrompt = `あなたはアニメのあらすじを日本語に翻訳・整形する編集者です。
次の<synopsis>のあらすじを、自然で読みやすい日本語に翻訳してください。すでに日本語の場合は翻訳せず整形のみ行ってください。

出力ルール（厳守）:
- 出力は翻訳・整形後のあらすじ本文のみ。前置き・後書き・補足・引用符は一切付けないこと。
- HTMLタグや "(Source: ...)" などの出典注記は除去すること。
- 段落構成は維持し、余分な空行や行頭・行末の空白は除去すること。
- 作品名・固有名詞は一般的な日本語表記に合わせること。

<synopsis>
%s
</synopsis>`

// TranslateAndFormatSynopsis は英語などのあらすじを日本語へ翻訳し、書式を整える。
func (c *ClaudeClient) TranslateAndFormatSynopsis(ctx context.Context, text string) (string, error) {
	if !c.enabled {
		return "", fmt.Errorf("Claude API is not configured")
	}
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5_20251001,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(fmt.Sprintf(synopsisTranslationPrompt, text))),
		},
	})
	if err != nil {
		return "", fmt.Errorf("claude synopsis translation failed: %w", err)
	}

	// ContentBlockUnion は Text/Type を直接フィールドに持つため、型アサーション不要で抽出できる
	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return strings.TrimSpace(sb.String()), nil
}
