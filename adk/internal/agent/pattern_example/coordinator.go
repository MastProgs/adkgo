package pattern_example

import (
	"log"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"

	"adk/internal/provider"
)

// AI-NOTE: [5 & 6. 코디네이터 및 도구로서의 에이전트 패턴]
func BuildCoordinatorPattern(reg *provider.ModelRegistry) agent.Agent {
	modelMath, _ := reg.GetModel(provider.ModelGemini)
	mathAgent, _ := llmagent.New(llmagent.Config{
		Name:        "math_expert",
		Description: "수학 계산이나 논리 퍼즐을 전문으로 해결합니다.",
		Model:       modelMath,
		Instruction: "당신은 세계 최고의 수학자입니다. 과정과 정답을 명확히 풀어주세요.",
	})

	modelTrans, _ := reg.GetModel(provider.ModelGPT)
	transAgent, _ := llmagent.New(llmagent.Config{
		Name:        "translator",
		Description: "문장을 다른 언어로 전문적으로 번역합니다.",
		Model:       modelTrans,
		Instruction: "주어진 텍스트를 가장 자연스러운 한국어로 번역하세요.",
	})

	mathTool := agenttool.New(mathAgent, &agenttool.Config{})
	transTool := agenttool.New(transAgent, &agenttool.Config{})

	modelCoordinator, _ := reg.GetModel(provider.ModelClaude)
	coordinatorAgent, err := llmagent.New(llmagent.Config{
		Name:        "coordinator_example",
		Description: "요청을 분석하고 적절한 전문가(도구)에게 작업을 위임하는 매니저",
		Model:       modelCoordinator,
		Instruction: `당신은 총괄 매니저입니다. 
사용자의 질문이 수학이라면 math_expert 도구를 사용하시고, 
번역이 필요하다면 translator 도구를 사용해서 그들의 답변을 반환하세요.
직접 대답하지 말고 반드시 도구를 사용해야 합니다.`,
		Tools: []tool.Tool{mathTool, transTool},
	})

	if err != nil {
		log.Fatalf("코디네이터 패턴 생성 실패: %v", err)
	}

	return coordinatorAgent
}
