package sriov

import (
	"context"
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

// Run executes the state machine loop, simulating the SR-IOV configuration phases.
func (sm *StateMachine) Run(ctx context.Context) error {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Initial state transition
	sm.currentState = StateDraining
	time.Sleep(100 * time.Millisecond) // micro sleep to let status settle

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			switch sm.currentState {
			case StateIdle:
				sm.currentState = StateDraining
			case StateDraining:
				// Simulate drain complete
				sm.currentState = StateInProgress
			case StateInProgress:
				// Simulate configuration success or failure
				if sm.failureRate > 0 && time.Now().UnixNano()%100 < int64(sm.failureRate*100) {
					sm.currentState = StateFailed
				} else {
					sm.currentState = StateSucceeded
				}
			case StateFailed:
				// Retry config
				sm.currentState = StateInProgress
			case StateSucceeded:
				// Configuration is stable
			}
		}
	}
}

// GetState returns the current state
func (sm *StateMachine) GetState() State {
	return sm.currentState
}
