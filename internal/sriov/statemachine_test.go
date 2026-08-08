package sriov

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStateMachine_Lifecycle(t *testing.T) {
	// Create state machine with 0 failure rate to ensure successful transitions
	sm := NewStateMachine("node-1", 10*time.Millisecond, 10*time.Millisecond, 0.0)

	assert.Equal(t, StateIdle, sm.GetState())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run state machine in background
	go func() {
		err := sm.Run(ctx)
		assert.NoError(t, err)
	}()

	// Run immediately sets state to StateDraining and sleeps 100ms.
	time.Sleep(200 * time.Millisecond)

	// Verify state transitioned out of Idle thread-safely
	currentState := sm.GetState()
	assert.NotEqual(t, StateIdle, currentState)
}

func TestStateMachine_ManualTransitions(t *testing.T) {
	sm := NewStateMachine("node-1", time.Millisecond, time.Millisecond, 0.0)

	// Simulate Idle -> Draining
	assert.Equal(t, StateIdle, sm.GetState())
	sm.SetState(StateDraining)

	// Simulate Draining -> InProgress
	assert.Equal(t, StateDraining, sm.GetState())
	sm.SetState(StateInProgress)

	// Simulate InProgress -> Succeeded
	assert.Equal(t, StateInProgress, sm.GetState())
	sm.SetState(StateSucceeded)

	assert.Equal(t, StateSucceeded, sm.GetState())
}
