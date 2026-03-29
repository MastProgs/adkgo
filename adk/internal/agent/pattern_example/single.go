package pattern_example

import (
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"

	"adk/internal/provider"
)

// AI-NOTE: [1. 단일(Single) 패턴]
func BuildSinglePattern(reg *provider.ModelRegistry) agent.Agent {
	m, err := reg.GetModel(provider.ModelGemini)
	if err != nil {
		log.Fatalf("모델 로드 실패: %v", err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "single_agent_example",
		Description: "단일 챗봇 에이전트입니다.",
		Model:       m,
		Instruction: "당신은 사용자에게 가장 기본적이고 친절한 답변을 제공하는 단일 에이전트입니다.",
	})
	if err != nil {
		log.Fatalf("단일 에이전트 생성 실패: %v", err)
	}

	return a
}
