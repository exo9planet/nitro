// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package staker

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/offchainlabs/nitro/validator"
)

type ExecutionChallengeBackend struct {
	exec          validator.ExecutionRun
	lastStep      validator.MachineStep
	lastStepMutex sync.Mutex
}

// NewExecutionChallengeBackend creates a backend with the given arguments.
// Note: machineCache may be nil, but if present, it must not have a restricted range.
func NewExecutionChallengeBackend(executionRun validator.ExecutionRun) (*ExecutionChallengeBackend, error) {
	return &ExecutionChallengeBackend{
		exec: executionRun,
	}, nil
}

func (b *ExecutionChallengeBackend) SetRange(ctx context.Context, start uint64, end uint64) error {
	b.exec.PrepareRange(start, end)
	return nil
}

func (b *ExecutionChallengeBackend) getStep(ctx context.Context, position uint64) (validator.MachineStep, error) {
	b.lastStepMutex.Lock()
	lastStep := b.lastStep
	b.lastStepMutex.Unlock()
	if lastStep != nil && lastStep.Ready() {
		lastpos, err := lastStep.Position()
		if err != nil && lastpos == position {

			return lastStep, nil
		}
	}
	step := b.exec.GetStepAt(position)
	err := step.WaitReady(ctx)
	if err != nil {
		b.lastStepMutex.Lock()
		b.lastStep = step
		b.lastStepMutex.Unlock()
	}
	return step, err
}

func (b *ExecutionChallengeBackend) GetHashAtStep(ctx context.Context, position uint64) (common.Hash, error) {
	step, err := b.getStep(ctx, position)
	if err != nil {
		return common.Hash{}, err
	}
	return step.Hash()
}

func (b *ExecutionChallengeBackend) GetProofAt(
	ctx context.Context,
	position uint64,
) ([]byte, error) {
	step, err := b.getStep(ctx, position)
	if err != nil {
		return nil, err
	}
	return step.Proof()
}

func finalStateError(err error) (uint64, validator.GoGlobalState, uint8, error) {
	return 0, validator.GoGlobalState{}, 0, err
}

func (b *ExecutionChallengeBackend) GetFinalState(ctx context.Context) (uint64, validator.GoGlobalState, uint8, error) {
	step := b.exec.GetLastStep()
	err := step.WaitReady(ctx)
	if err != nil {
		return finalStateError(err)
	}
	pos, err := step.Position()
	if err != nil {
		return finalStateError(err)
	}
	status, err := step.Status()
	if err != nil {
		return finalStateError(err)
	}
	globalstate, err := step.GlobalState()
	if err != nil {
		return finalStateError(err)
	}
	return pos, globalstate, uint8(status), nil
}
