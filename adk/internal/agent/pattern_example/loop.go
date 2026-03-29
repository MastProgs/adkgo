package pattern_example

import (
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/loopagent"

	"adk/internal/provider"
)

// AI-NOTE: [4. 루프(Loop) 에이전트 패턴]
func BuildLoopPattern(reg *provider.ModelRegistry) agent.Agent {
	modelDrafter, _ := reg.GetModel(provider.ModelGemini)
	drafterAgent, _ := llmagent.New(llmagent.Config{
		Name:        "drafter",
		Description: "문서 초안을 작성하는 에이전트",
		Model:       modelDrafter,
		Instruction: "사용자의 요청이나 이전 비판을 듣고 문장을 새롭게 작성하세요.",
	})

	modelCritic, _ := reg.GetModel(provider.ModelClaude)
	criticAgent, _ := llmagent.New(llmagent.Config{
		Name:        "critic",
		Description: "초안을 검토하고 날카롭게 비판하는 에이전트",
		Model:       modelCritic,
		Instruction: "이전에 작성된 초안을 읽고, 부족한 점을 꼬집어 더 나은 방향으로 수정 지시를 내리세요.",
	})

	loopAgent, err := loopagent.New(loopagent.Config{
		MaxIterations: 3,
		AgentConfig: agent.Config{
			Name:        "loop_workflow_example",
			Description: "작성자와 검토자가 만족할 때까지 문서 품질을 갈고 닦습니다.",
			SubAgents:   []agent.Agent{drafterAgent, criticAgent},
		},
	})

	if err != nil {
		log.Fatalf("루프 패턴 생성 실패: %v", err)
	}

	return loopAgent
}
