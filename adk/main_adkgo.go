package main

import (
	"context"
	"log"
	"os"

	"adk/internal/agent/pattern_example"
	"adk/internal/config"
	"adk/internal/provider"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
)

// AI-NOTE: ADK 다중 에이전트 등록 예시 (멀티 세션 지원)
// Web UI 좌측 리스트에 여러 에이전트가 나열되며, 세션마다 독립적인 에이전트와 대화할 수 있습니다.
func main() {
	ctx := context.Background()

	// 1. 설정 로드 (TOML은 모델 리소스 관리에만 집중)
	cfg, err := config.Load("config.toml")
	if err != nil {
		log.Fatalf("환경 설정 파일을 읽을 수 없습니다: %v", err)
	}

	// 2. 모델 레지스트리 가동 (Gemini, GPT, Claude 모두 장전)
	registry := provider.NewModelRegistry(cfg.Models)

	// 3. 각 용도에 맞는 독립적인 에이전트(워크플로우)를 자유롭게 조립합니다.
	// AI-NOTE: 아래 에이전트들은 서로 완전히 독립된 세션으로 Web UI에 개별 등록됩니다.
	// 하나의 서버에서 여러 목적의 에이전트를 동시에 운용할 수 있습니다.
	// -----------------------------------------------------------------------
	// [루트 에이전트] Web UI 접속 시 기본으로 선택되는 에이전트입니다.
	rootAgent := pattern_example.BuildSinglePattern(registry)

	// [추가 에이전트] 특수 목적 에이전트들을 필요한 수만큼 등록할 수 있습니다.
	sequentialAgent := pattern_example.BuildSequentialPattern(registry) // 릴레이형 파이프라인 워크플로우
	parallelAgent := pattern_example.BuildParallelPattern(registry)     // 동시 다발 브레인스토밍 워크플로우
	loopAgent := pattern_example.BuildLoopPattern(registry)             // 반복 품질 개선 워크플로우
	coordinatorAgent := pattern_example.BuildCoordinatorPattern(registry) // 전문가 위임 워크플로우
	// -----------------------------------------------------------------------

	// 4. NewMultiLoader로 모든 에이전트를 하나의 런처 Config에 한꺼번에 등록
	// AI-NOTE: NewSingleLoader 대신 NewMultiLoader를 사용하면,
	// Web UI 좌측 사이드바에 등록된 에이전트들이 모두 나열됩니다.
	// 첫 번째 인자(rootAgent)가 기본 선택 에이전트이며,
	// 이름이 중복된 에이전트가 있으면 에러가 발생하므로 주의하세요.
	multiLoader, err := adkagent.NewMultiLoader(
		rootAgent,
		sequentialAgent,
		parallelAgent,
		loopAgent,
		coordinatorAgent,
	)
	if err != nil {
		log.Fatalf("멀티 에이전트 로더 생성 실패 (에이전트 이름 중복 확인 필요): %v", err)
	}

	launcherCfg := &launcher.Config{
		AgentLoader: multiLoader,
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, launcherCfg, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

