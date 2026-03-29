package pattern_example

import (
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"

	"adk/internal/provider"
)

// AI-NOTE: [2. 순차(Sequential) 에이전트 패턴]
func BuildSequentialPattern(reg *provider.ModelRegistry) agent.Agent {
	modelA, _ := reg.GetModel(provider.ModelClaude)
	agentA, _ := llmagent.New(llmagent.Config{
		Name:        "extractor",
		Description: "주어진 정보에서 핵심만 추출하는 에이전트입니다.",
		Model:       modelA,
		Instruction: "사용자의 입력에서 가장 핵심적인 키워드 3가지만 추출해 주세요.",
	})

	modelB, _ := reg.GetModel(provider.ModelGemini)
	agentB, _ := llmagent.New(llmagent.Config{
		Name:        "writer",
		Description: "키워드를 바탕으로 최종 글을 짓는 에이전트입니다.",
		Model:       modelB,
		Instruction: "이전에 추출된 키워드를 바탕으로 3줄짜리 시를 지어주세요.",
	})

	seqAgent, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:        "sequential_workflow_example",
			Description: "순차적으로 리서처와 작성자를 거치는 워크플로우",
			SubAgents:   []agent.Agent{agentA, agentB},
		},
	})

	if err != nil {
		log.Fatalf("순차 패턴 생성 실패: %v", err)
	}

	return seqAgent
}
