# ADK-Go 멀티 에이전트 프레임워크

> **Google ADK (Agent Development Kit) for Go**를 기반으로, Gemini · GPT · Claude 등 여러 LLM을 조합하여 복잡한 멀티 에이전트 워크플로우를 구성하는 방법을 보여주는 레퍼런스 프로젝트입니다.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![ADK](https://img.shields.io/badge/Google%20ADK-Go-4285F4?logo=google)](https://github.com/google/adk-go)

---

## ✨ 주요 특징

| 기능 | 설명 |
|---|---|
| **멀티 프로바이더** | Gemini, OpenAI(GPT), Anthropic(Claude) 모델을 에이전트별로 독립 할당 |
| **TOML 기반 설정** | `config.toml` 한 파일로 모든 API 키와 모델 리소스를 중앙 관리 |
| **자동 Fallback** | 쿼터 초과(429) · 인증 오류(401) 발생 시 예비 API 키로 자동 전환 |
| **Enum 타입 안전** | 모델 ID와 공급자를 Enum 상수로 관리하여 오타 등의 런타임 에러를 컴파일 시점에서 차단 |
| **6가지 디자인 패턴** | Single · Sequential · Parallel · Loop · Coordinator(Agent-as-Tool) 예시 제공 |
| **멀티 세션 지원** | 하나의 서버에 여러 에이전트를 동시 등록, 세션마다 독립된 워크플로우 운용 |
| **Web UI 내장** | ADK의 React 기반 개발 UI를 통해 브라우저에서 바로 에이전트와 대화 가능 |

---

## 📂 프로젝트 구조

```
adk/
├── main_adkgo.go                   # 진입점: 에이전트 조립 및 서버 가동
├── config.toml                     # LLM 공급자별 API 키 및 모델 설정
├── go.mod / go.sum
└── internal/
    ├── config/
    │   └── config.go               # TOML 파싱 → Config 구조체
    ├── provider/
    │   ├── enums.go                # ModelID, ProviderType Enum 상수 정의
    │   ├── errors.go               # API 에러 분류 Enum (ErrRateLimit, ErrUnauthorized...)
    │   ├── key_manager.go          # API 키 로테이션 및 Fallback 로직
    │   ├── registry.go             # 모델 레지스트리 (ModelID → model.LLM 매핑)
    │   ├── gemini.go               # Gemini 모델 어댑터 (Fallback 탑재)
    │   ├── openai.go               # OpenAI 모델 어댑터 (Stub - 확장 예정)
    │   └── claude.go               # Anthropic 모델 어댑터 (Stub - 확장 예정)
    └── agent/
        └── pattern_example/        # ⬇ 6가지 디자인 패턴 예시 (복사하여 활용)
            ├── single.go           # 단일 에이전트 패턴
            ├── sequential.go       # 순차 에이전트 패턴
            ├── parallel.go         # 병렬 에이전트 패턴
            ├── loop.go             # 루프 에이전트 패턴
            └── coordinator.go      # 코디네이터 + Agent-as-Tool 패턴
```

---

## 🛠️ 시작하기

### 사전 요구 사항

- **Go 1.25 이상** ([다운로드](https://go.dev/dl/))
- 사용할 LLM의 API 키 (Gemini, OpenAI, Anthropic 중 하나 이상)

### 1단계: 저장소 클론 및 의존성 설치

```bash
git clone <this-repo-url>
cd adk
go mod tidy
```

### 2단계: API 키 설정

`config.toml`을 열고 `api_keys` 배열에 실제 발급받은 API 키를 입력합니다.

```toml
[models]
  [models.gemini]
  provider = "gemini"
  model_name = "gemini-2.5-flash"
  # api_keys 배열: 첫 번째 키가 소진되면 두 번째 키를 자동으로 사용합니다.
  api_keys = [
    "AIzaSy-PRIMARY-KEY",
    "AIzaSy-FALLBACK-KEY"    # 예비 키 (선택 사항)
  ]

  [models.gpt]
  provider = "openai"
  model_name = "gpt-4o"
  api_keys = ["sk-proj-YOUR-OPENAI-KEY"]

  [models.claude]
  provider = "anthropic"
  model_name = "claude-3-5-sonnet-20240620"
  api_keys = ["sk-ant-YOUR-ANTHROPIC-KEY"]
```

> **팁**: 현재 OpenAI와 Anthropic 어댑터는 Stub 상태입니다. `internal/provider/openai.go`와 `claude.go`에 실제 SDK 연동 코드를 추가하여 확장하세요.

### 3단계: 서버 실행

```bash
go run main_adkgo.go web -port 8000 webui -api_server_address http://localhost:8000/api api -webui_address localhost:8000
```

실행에 성공하면 다음과 같이 출력됩니다.

```
Web servers starts on http://localhost:8000
       webui:  you can access API using http://localhost:8000/ui/
       api:    you can access API using http://localhost:8000/api
```

### 4단계: Web UI 접속

브라우저에서 **http://localhost:8000/ui/** 로 접속합니다.  
좌측 사이드바에 등록된 에이전트 목록이 나열되며, 클릭하여 해당 에이전트와 독립 세션으로 대화할 수 있습니다.

---

## 💡 에이전트 디자인 패턴

이 프로젝트의 핵심은 하나의 서버에 **여러 목적의 에이전트를 동시에 등록하는 것**입니다.  
모든 예시는 `internal/agent/pattern_example/` 폴더에 있으며, 복사하여 자유롭게 커스터마이징하시면 됩니다.

### 1. 🧩 Single — 단일 에이전트

가장 기본적인 형태입니다. 하나의 LLM 모델로 단일 목적 챗봇을 만들 때 사용합니다.

```
사용자 입력 ──▶ [Gemini] ──▶ 응답
```

> **사용 예시**: FAQ 챗봇, 기본 Q&A

---

### 2. 🔗 Sequential — 순차 에이전트

앞 에이전트의 출력이 다음 에이전트의 입력으로 전달되는 파이프라인 구조입니다.

```
사용자 입력 ──▶ [Claude: 키워드 추출] ──▶ [Gemini: 시 작성] ──▶ 응답
```

> **사용 예시**: 리서치 → 요약 → 번역 파이프라인, 데이터 전처리 체인

---

### 3. ⚡ Parallel — 병렬 에이전트

여러 에이전트가 **동시에** 같은 입력을 처리하고, 결과를 한꺼번에 반환합니다.

```
                  ┌─▶ [Gemini: 낙관적 의견]
사용자 입력 ──────┤
                  └─▶ [GPT: 비관적 의견]
                  
두 결과를 취합하여 응답
```

> **사용 예시**: 아이디어 브레인스토밍, 다방면 리서치, A/B 의견 수집

---

### 4. 🔄 Loop — 루프 에이전트

원하는 품질에 도달할 때까지 서브 에이전트 파이프라인을 반복 실행합니다.

```
사용자 입력 ──▶ [Gemini: 초안 작성] ──▶ [Claude: 비판] ──┐
                         ▲                                          │ MaxIterations(3회) 반복
                         └──────────────────────────────────────────┘
```

> **사용 예시**: 코드 자동 검토 및 개선, 문서 품질 향상, 반복 최적화

---

### 5. 🎯 Coordinator — 코디네이터 + Agent-as-Tool

**가장 강력한 패턴입니다.** 메인 관리자 에이전트가 상황을 분석하여, 등록된 전문 에이전트들 중 적합한 것을 도구(Tool)로 호출합니다.
`agenttool.New()`를 사용하여 에이전트 자체를 Function Call 도구로 변환하는 것이 핵심입니다.

```
                          ┌──▶ math_expert [Gemini] (수학 질문일 때)
사용자 입력 ──▶ [Claude: 매니저]
                          └──▶ translator [GPT] (번역 요청일 때)
```

> **사용 예시**: 지능형 라우팅 시스템, 전문 분야 위임, Multi-Specialist AI

---

## 🔑 Fallback 동작 방식

`config.toml`에 API 키를 여러 개 등록해 두면, 아래의 에러 상황에서 자동으로 다음 키를 사용합니다.

| 에러 타입 | HTTP 코드 | Fallback 여부 |
|---|---|---|
| Rate Limit (쿼터 초과) | 429 | ✅ 예비 키로 전환 |
| Unauthorized (인증 실패) | 401 | ✅ 예비 키로 전환 |
| Token Limit 초과 | — | ✅ 예비 키로 전환 |
| 네트워크 오류 | — | ❌ 즉시 에러 반환 |
| 서버 오류 | 5xx | ❌ 즉시 에러 반환 |

---

## 🔧 나만의 에이전트 만들기

`pattern_example` 폴더의 파일을 참고하여 새로운 에이전트를 만들고 `main_adkgo.go`에 등록합니다.

```go
// 1. 레지스트리에서 원하는 모델을 Enum으로 안전하게 가져옵니다.
myModel, _ := registry.GetModel(provider.ModelGemini)

// 2. 에이전트 생성
myAgent, _ := llmagent.New(llmagent.Config{
    Name:        "my_custom_agent",  // Web UI에 표시되는 이름 (고유해야 합니다)
    Description: "나만의 에이전트",
    Model:       myModel,
    Instruction: "당신은 ... 역할을 하는 에이전트입니다.",
})

// 3. main_adkgo.go의 NewMultiLoader에 추가
multiLoader, _ := adkagent.NewMultiLoader(
    rootAgent,
    myAgent,  // 여기에 추가
    // ...
)
```

---

## 📌 자주 묻는 질문 (FAQ)

**Q. 포트를 변경하고 싶어요.**  
A. 실행 명령어에서 `-port 8000` 부분의 숫자를 변경하세요. 단, 세 곳(`-port`, `-api_server_address`, `-webui_address`)의 포트 번호를 모두 동일하게 맞춰야 합니다.

```bash
go run main_adkgo.go web -port 9090 webui -api_server_address http://localhost:9090/api api -webui_address localhost:9090
```

**Q. 에이전트 이름이 중복되면 어떻게 되나요?**  
A. `NewMultiLoader` 호출 시 즉시 에러가 발생하고 서버가 시작되지 않습니다. 각 `Build*Pattern()` 함수 내부의 `Name` 필드를 고유하게 설정하세요.

**Q. OpenAI와 Claude는 아직 지원이 안 되나요?**  
A. `internal/provider/openai.go`와 `claude.go`는 현재 Stub 구조로 작성되어 있습니다. 각 파일에 실제 SDK(`go-openai`, `anthropic-sdk-go` 등)를 연동하고 `genai.Content` 스키마를 각 API 형식으로 변환하는 어댑터 코드를 추가하면 완성됩니다.

---

## 🔗 참고 링크

- [Google ADK for Go 공식 저장소](https://github.com/google/adk-go)
- [Google Gen AI Go SDK](https://github.com/google/generative-ai-go)
- [Google AI Studio (API 키 발급)](https://aistudio.google.com/apikey)
- [OpenAI API Keys](https://platform.openai.com/api-keys)
- [Anthropic API Keys](https://console.anthropic.com/)
