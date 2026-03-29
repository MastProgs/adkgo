package pattern_example

import (
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/agent/workflowagents/parallelagent"

	"adk/internal/provider"
)

// AI-NOTE: [3. 병렬(Parallel) 에이전트 패턴]
func BuildParallelPattern(reg *provider.ModelRegistry) agent.Agent {
	modelA, _ := reg.GetModel(provider.ModelGemini)
	agentA, _ := llmagent.New(llmagent.Config{
		Name:        "optimist",
		Description: "낙관적인 관점의 에이전트",
		Model:       modelA,
		Instruction: "주어진 안건에 대해 긍정적인 평가만 1줄로 제시하세요.",
	})

	modelB, _ := reg.GetModel(provider.ModelGPT)
	agentB, _ := llmagent.New(llmagent.Config{
		Name:        "pessimist",
		Description: "비관적인 관점의 에이전트",
		Model:       modelB,
		Instruction: "주어진 안건에 대해 부정적인 평가만 1줄로 제시하세요.",
	})

	parallelAgent, err := parallelagent.New(parallelagent.Config{
		AgentConfig: agent.Config{
			Name:        "parallel_workflow_example",
			Description: "낙관론자와 비관론자에게 동시에 브레인스토밍을 맡깁니다.",
			SubAgents:   []agent.Agent{agentA, agentB},
		},
	})

	if err != nil {
		log.Fatalf("병렬 패턴 생성 실패: %v", err)
	}

	return parallelAgent
}
