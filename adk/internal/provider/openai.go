package provider

import (
	"context"
	"iter"
	"log"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// AI-NOTE: OpenAI GPT 모델을 ADK의 model.LLM 인터페이스에 맞게 연결하는 실제 어댑터입니다.
// Claude 어댑터와 동일한 구조로, genai.Content ↔ OpenAI Chat Completions API 형식 간
// 양방향 변환을 수행하며, KeyManager를 통한 자동 Fallback 기능이 포함되어 있습니다.
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

func (p *OpenAIProvider) Name() string { return p.modelName }

func (p *OpenAIProvider) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		for {
			currentKey := p.km.GetCurrentKey()
			client := openai.NewClient(option.WithAPIKey(currentKey))

			// -------------------------------------------------------
			// STEP 1. ADK의 genai.Content 배열 → OpenAI Messages 형식 변환
			// -------------------------------------------------------
			var messages []openai.ChatCompletionMessageParamUnion

			// AI-NOTE: llmagent의 Instruction은 req.Config.SystemInstruction에 담겨 옵니다.
			if req.Config != nil && req.Config.SystemInstruction != nil {
				systemText := extractContentText(req.Config.SystemInstruction)
				if systemText != "" {
					messages = append(messages, openai.SystemMessage(systemText))
				}
			}

			for _, content := range req.Contents {
				text := extractContentText(content)
				if text == "" {
					continue
				}
				switch content.Role {
				case "user":
					messages = append(messages, openai.UserMessage(text))
				case "model", "assistant":
					messages = append(messages, openai.AssistantMessage(text))
				}
			}

			// -------------------------------------------------------
			// STEP 2. OpenAI Chat Completions API 호출
			// -------------------------------------------------------
			resp, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
				Model:    openai.ChatModel(p.modelName),
				Messages: messages,
			})
			if err != nil {
				apiErr := mapOpenAIError(err)
				handledErr := p.km.HandleError(apiErr)
				if handledErr != nil && handledErr.Error() == "fallback_triggered" {
					log.Printf("AI-NOTE: OpenAI Fallback 시도 (다음 API 키 사용)")
					continue
				}
				yield(nil, apiErr)
				return
			}

			if len(resp.Choices) == 0 {
				yield(nil, &APIError{Type: ErrUnknown, Message: "OpenAI returned empty choices"})
				return
			}

			// -------------------------------------------------------
			// STEP 3. OpenAI 응답 → ADK의 genai.Content 형식으로 역변환
			// -------------------------------------------------------
			text := resp.Choices[0].Message.Content
			adkContent := &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: text}},
			}

			yield(&model.LLMResponse{Content: adkContent}, nil)
			return
		}
	}
}

// AI-NOTE: OpenAI SDK 에러 메시지를 내부 Enum(APIError)으로 분류합니다.
func mapOpenAIError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "429") || strings.Contains(msg, "rate limit") || strings.Contains(msg, "too many requests") {
		return &APIError{Type: ErrRateLimit, Message: "OpenAI Rate Limit Exceeded", Cause: err}
	} else if strings.Contains(msg, "401") || strings.Contains(msg, "unauthorized") || strings.Contains(msg, "invalid api key") {
		return &APIError{Type: ErrUnauthorized, Message: "OpenAI Unauthorized API Key", Cause: err}
	} else if strings.Contains(msg, "context_length") || strings.Contains(msg, "maximum context") || strings.Contains(msg, "token") {
		return &APIError{Type: ErrTokenExceeded, Message: "OpenAI Token Limit Exceeded", Cause: err}
	}
	return &APIError{Type: ErrUnknown, Message: "Unknown OpenAI Error", Cause: err}
}
