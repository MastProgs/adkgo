package provider

import (
	"context"
	"iter"
	"strings"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// AI-NOTE: Gemini 모델 전용 Provider 래퍼입니다.
// 키 매니저를 가지고 있으며, API 호출 도중 토큰 초과 등의 오류 발생 시 예비 키로 Fallback을 수행합니다.
type GeminiProvider struct {
	modelName string
	km        *KeyManager
}

func NewGeminiProvider(modelName string, keys []string) *GeminiProvider {
	return &GeminiProvider{
		modelName: modelName,
		km:        NewKeyManager(keys), // AI-NOTE: 넘겨받은 배열로 Fallback 관리 시작
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		for {
			currentKey := p.km.GetCurrentKey()
			// 현재 키 기반으로 모델 객체 조립
			m, err := gemini.NewModel(ctx, p.modelName, &genai.ClientConfig{APIKey: currentKey})
			if err != nil {
				yield(nil, err)
				return
			}
			
			reqIter := m.GenerateContent(ctx, req, stream)
			var iterError error
			
			// 이터레이터를 순회하며 ADK로 응답 흘려주기
			for resp, rErr := range reqIter {
				if rErr != nil {
					iterError = rErr
					break
				}
				// yield는 계속 값을 받아올지 여부를 반환합니다. false면 스톱.
				if !yield(resp, nil) {
					return 
				}
			}

			// 에러가 있다면 맵핑 및 Fallback 판단
			if iterError != nil {
				apiErr := mapGeminiError(iterError) // Enum 변환
				
				// 에러 핸들러 호출: 복구 가능하면 Index 증가시키고 "fallback_triggered" 리턴
				handledErr := p.km.HandleError(apiErr)
				if handledErr != nil && handledErr.Error() == "fallback_triggered" {
					// 루프를 돌며 새로운 키(변경된 index)로 다시 GenerateContent를 시도합니다.
					continue
				}
				
				// Fallback 불가 시엔 원본 오류 던짐
				yield(nil, apiErr)
				return
			}
			
			// 성공 시 루프 종료
			return 
		}
	}
}

// AI-NOTE: Google genai 오류 문구를 기반으로, Enum값(오류 종류)을 분석하여 APIError로 감쌉니다.
func mapGeminiError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "quota") || strings.Contains(msg, "429") || strings.Contains(msg, "too many requests") {
		return &APIError{Type: ErrRateLimit, Message: "Quota exceeded for current key", Cause: err}
	} else if strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized") {
		return &APIError{Type: ErrUnauthorized, Message: "Unauthorized API Key", Cause: err}
	} else if strings.Contains(msg, "token limit") || strings.Contains(msg, "maximum context length") {
		return &APIError{Type: ErrTokenExceeded, Message: "Token limit exceeded", Cause: err}
	}
	return &APIError{Type: ErrUnknown, Message: "Unknown Gemini Error", Cause: err}
}
