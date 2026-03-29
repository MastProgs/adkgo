package provider

import (
	"fmt"
	"log"

	"google.golang.org/adk/model"
	"adk/internal/config"
)

// AI-NOTE: 애플리케이션에서 사용하는 모든 외부 LLM(Gemini, GPT, Claude 등)을 
// 초기화하여 메모리에 보관하고, 특정 에이전트가 요구할 때 꺼내어 주는 전역 보관소(Registry)입니다.
type ModelRegistry struct {
	models map[ModelID]model.LLM
}

// NewModelRegistry는 TOML 컨피그에 적힌 모든 모델을 초기화하여 레지스트리에 저장합니다.
func NewModelRegistry(cfgMap map[string]config.ModelConfig) *ModelRegistry {
	reg := &ModelRegistry{
		models: make(map[ModelID]model.LLM),
	}

	for id, mdlCfg := range cfgMap {
		var m model.LLM
		pType := ProviderType(mdlCfg.Provider) // Enum 변환
		
		switch pType {
		case ProviderGemini:
			m = NewGeminiProvider(mdlCfg.ModelName, mdlCfg.APIKeys)
		case ProviderOpenAI:
			m = NewOpenAIProvider(mdlCfg.ModelName, mdlCfg.APIKeys)
		case ProviderAnthropic:
			m = NewClaudeProvider(mdlCfg.ModelName, mdlCfg.APIKeys)
		default:
			log.Printf("경고: 지원하지 않는 프로바이더가 감지되었습니다 (%s), 건너뜁니다.", mdlCfg.Provider)
			continue
		}
		
		reg.models[ModelID(id)] = m
		log.Printf("AI-NOTE: 모델 레지스트리에 등록됨 - [%s] (Provider: %s)", id, mdlCfg.Provider)
	}

	return reg
}

// GetModel은 에이전트가 Enum을 사용해 안전하게 모델을 꺼내갈 수 있도록 보장합니다.
func (r *ModelRegistry) GetModel(id ModelID) (model.LLM, error) {
	if m, ok := r.models[id]; ok {
		return m, nil
	}
	// Fallback 정책: 찾지 못하면 레지스트리에 있는 아무 모델이나 반환합니다.
	for fallbackId, m := range r.models {
		log.Printf("AI-NOTE: 요청한 모델 [%s]을 찾을 수 없어 기본 모델 [%s]로 대체합니다.", id, fallbackId)
		return m, nil
	}
	return nil, fmt.Errorf("레지스트리에 등록된 모델이 하나도 없습니다")
}
