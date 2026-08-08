package sriov

import (
	"context"
	"sync"
	"time"
)

// State represents the state of SR-IOV configuration.
type State string

const (
	StateIdle       State = "Idle"
	StateInProgress State = "InProgress"
	StateDraining   State = "Draining"
	StateSucceeded  State = "Succeeded"
	StateFailed     State = "Failed"
)

// StateMachine manages the transitions for SR-IOV config.
type StateMachine struct {
	mu           sync.RWMutex
	nodeName     string
	drainDelay   time.Duration
	configDelay  time.Duration
	failureRate  float64
	currentState State
}

// NewStateMachine creates a new StateMachine.
func NewStateMachine(nodeName string, drainDelay, configDelay time.Duration, failureRate float64) *StateMachine {
	return &StateMachine{
		nodeName:     nodeName,
		drainDelay:   drainDelay,
		configDelay:  configDelay,
		failureRate:  failureRate,
		currentState: StateIdle,
	}
}

// SetState safely updates the current state (thread-safe).
func (sm *StateMachine) SetState(s State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.currentState = s
}

// GetState returns the current state (thread-safe).
func (sm *StateMachine) GetState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

// Run executes the state machine loop, simulating the SR-IOV configuration phases.
func (sm *StateMachine) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial state transition
	sm.SetState(StateDraining)
	time.Sleep(100 * time.Millisecond) // micro sleep to let status settle

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			current := sm.GetState()
			switch current {
			case StateIdle:
				sm.SetState(StateDraining)
			case StateDraining:
				// Simulate drain complete
				sm.SetState(StateInProgress)
			case StateInProgress:
				// Simulate configuration success or failure
				if sm.failureRate > 0 && time.Now().UnixNano()%100 < int64(sm.failureRate*100) {
					sm.SetState(StateFailed)
				} else {
					sm.SetState(StateSucceeded)
				}
			case StateFailed:
				// Retry config
				sm.SetState(StateInProgress)
			case StateSucceeded:
				// Configuration is stable
			}
		}
	}
}
