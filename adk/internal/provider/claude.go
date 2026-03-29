package provider

import (
	"context"
	"iter"
	"log"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// AI-NOTE: Anthropic Claude 모델을 ADK의 model.LLM 인터페이스에 맞게 연결하는 실제 어댑터입니다.
// Google ADK Go에는 공식 Claude 패키지가 없기 때문에(Java ADK에만 존재),
// Anthropic 공식 Go SDK(github.com/anthropics/anthropic-sdk-go)를 직접 연동하여
// genai.Content ↔ Anthropic Messages API 형식 간 양방향 변환을 수행합니다.
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

func (p *ClaudeProvider) Name() string { return p.modelName }

func (p *ClaudeProvider) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		for {
			currentKey := p.km.GetCurrentKey()
			client := anthropic.NewClient(option.WithAPIKey(currentKey))

			// -------------------------------------------------------
			// STEP 1. ADK의 genai.Content 배열 → Anthropic Messages 형식 변환
			// -------------------------------------------------------
			var systemPrompt string
			var messages []anthropic.MessageParam

			// AI-NOTE: llmagent의 Instruction은 req.Config.SystemInstruction에 담겨 옵니다.
			if req.Config != nil && req.Config.SystemInstruction != nil {
				systemPrompt = extractContentText(req.Config.SystemInstruction)
			}

			for _, content := range req.Contents {
				text := extractContentText(content)
				if text == "" {
					continue
				}
				switch content.Role {
				case "user":
					messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(text)))
				case "model", "assistant":
					messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(text)))
				}
			}

			// -------------------------------------------------------
			// STEP 2. Anthropic API 파라미터 구성
			// -------------------------------------------------------
			params := anthropic.MessageNewParams{
				Model:     anthropic.Model(p.modelName),
				MaxTokens: 4096,
				Messages:  messages,
			}
			if systemPrompt != "" {
				params.System = []anthropic.TextBlockParam{
					{Type: "text", Text: systemPrompt},
				}
			}

			// -------------------------------------------------------
			// STEP 3. Anthropic API 호출
			// -------------------------------------------------------
			resp, err := client.Messages.New(ctx, params)
			if err != nil {
				apiErr := mapClaudeError(err)
				handledErr := p.km.HandleError(apiErr)
				if handledErr != nil && handledErr.Error() == "fallback_triggered" {
					log.Printf("AI-NOTE: Claude Fallback 시도 (다음 API 키 사용)")
					continue
				}
				yield(nil, apiErr)
				return
			}

			// -------------------------------------------------------
			// STEP 4. Anthropic 응답 → ADK의 genai.Content 형식으로 역변환
			// -------------------------------------------------------
			var responseText strings.Builder
			for _, block := range resp.Content {
				if block.Type == "text" {
					responseText.WriteString(block.Text)
				}
			}

			text := responseText.String()
			adkContent := &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: text}},
			}

			yield(&model.LLMResponse{Content: adkContent}, nil)
			return
		}
	}
}

// extractContentText는 genai.Content 내부 Parts에서 텍스트를 이어붙여 반환합니다.
func extractContentText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var sb strings.Builder
	for _, part := range content.Parts {
		sb.WriteString(part.Text)
	}
	return sb.String()
}

// AI-NOTE: Anthropic SDK 에러 메시지를 내부 Enum(APIError)으로 분류합니다.
func mapClaudeError(err error) error {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "429") || strings.Contains(msg, "rate_limit") {
		return &APIError{Type: ErrRateLimit, Message: "Claude Rate Limit Exceeded", Cause: err}
	} else if strings.Contains(msg, "401") || strings.Contains(msg, "authentication") {
		return &APIError{Type: ErrUnauthorized, Message: "Claude Unauthorized API Key", Cause: err}
	} else if strings.Contains(msg, "token") || strings.Contains(msg, "context_length") {
		return &APIError{Type: ErrTokenExceeded, Message: "Claude Token Limit Exceeded", Cause: err}
	}
	return &APIError{Type: ErrUnknown, Message: "Unknown Claude Error", Cause: err}
}
