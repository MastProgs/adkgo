package provider

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/adk/model"
)

// AI-NOTE: Anthropic(Claude) 추상화 레이어로 사용할 빈 껍데기(Stub) 클래스입니다.
type ClaudeProvider struct {
	modelName string
	km        *KeyManager
}

func NewClaudeProvider(modelName string, keys []string) *ClaudeProvider {
	return &ClaudeProvider{
		modelName: modelName,
		km:        NewKeyManager(keys),
	}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(nil, fmt.Errorf("Claude API is currently a stub structure. Not fully implemented."))
	}
}
