package provider

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/adk/model"
)

// AI-NOTE: OpenAI 추상화 레이어로 사용할 빈 껍데기(Stub) 클래스입니다.
// 실제 사용 시 go-openai 모듈 연동 및 genai.Content 변환 작업이 추가됩니다.
type OpenAIProvider struct {
	modelName string
	km        *KeyManager
}

func NewOpenAIProvider(modelName string, keys []string) *OpenAIProvider {
	return &OpenAIProvider{
		modelName: modelName,
		km:        NewKeyManager(keys),
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(nil, fmt.Errorf("OpenAI API is currently a stub structure. Not fully implemented."))
	}
}
