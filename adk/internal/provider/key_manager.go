package provider

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// AI-NOTE: API 키 배열을 관리하고, 오류 발생 시 다음 키로 안전하게 넘어가기 위한 매니저입니다.
type KeyManager struct {
	keys       []string
	currentIndex int
	mu         sync.Mutex
}

func NewKeyManager(keys []string) *KeyManager {
	if len(keys) == 0 {
		return &KeyManager{keys: []string{""}} // 기본 빈 키 방어
	}
	return &KeyManager{
		keys:       keys,
		currentIndex: 0,
	}
}

// GetCurrentKey는 현재 활성화된 API 키를 반환합니다.
func (k *KeyManager) GetCurrentKey() string {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.keys[k.currentIndex]
}

// HandleError는 발생한 에러를 분석하여 예비 키로 교체(Fallback)할지 결정합니다.
func (k *KeyManager) HandleError(err error) error {
	if err == nil {
		return nil
	}

	// 에러 타입 분석 후 복구 가능한지 확인
	if IsRecoverable(err) {
		k.mu.Lock()
		defer k.mu.Unlock()

		if k.currentIndex+1 < len(k.keys) {
			k.currentIndex++
			log.Printf("AI-NOTE: API 오류 복구 시도 (Fallback 시작). 새로운 키를 사용합니다 (Index: %d)", k.currentIndex)
			// 실패를 알리되 복구를 시도할 수 있도록 특별한 신호 반환
			return fmt.Errorf("fallback_triggered") 
		} else {
			log.Println("AI-NOTE: 모든 예비 API 키가 소진되었습니다.")
		}
	}
	
	return err
}

// WithFallback(Action) 래퍼를 만들어, 오류 시 자동으로 다시 시도하는 역할을 수행
func (k *KeyManager) WithFallback(action func(key string) error) error {
	var lastErr error
	for {
		currentKey := k.GetCurrentKey()
		err := action(currentKey)
		
		if err == nil {
			return nil
		}

		// 오류를 KeyManager에 판단 맡김
		handledErr := k.HandleError(err)
		
		if handledErr != nil && handledErr.Error() == "fallback_triggered" {
			// 다음 키로 넘어갔으므로 다시 루프(재시도)
			continue
		}
		
		// 복구할 수 없는 오류이거나 키가 소진됨
		lastErr = handledErr
		break
	}
	
	if lastErr != nil {
		return fmt.Errorf("모든 키 재시도 실패 또는 복구불가 에러: %w", lastErr)
	}
	return errors.New("알 수 없는 에러")
}
