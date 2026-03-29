package provider

// AI-NOTE: 코드 전반에서 오타나 마법의 문자열(Magic String)을 방지하기 위한 상수 모음입니다.

// ModelID는 TOML의 [models.*] 블록 키와 매핑됩니다.
type ModelID string

const (
	ModelGemini ModelID = "gemini"
	ModelGPT    ModelID = "gpt"
	ModelClaude ModelID = "claude"
)

// ProviderType은 실제 API 백엔드 벤더사를 의미합니다.
type ProviderType string

const (
	ProviderGemini    ProviderType = "gemini"
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
)
