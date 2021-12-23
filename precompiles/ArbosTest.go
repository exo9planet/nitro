//
// Copyright 2021, Offchain Labs, Inc. All rights reserved.
//

package precompiles

import (
	"errors"
)

type ArbosTest struct {
	Address addr
}

func (con ArbosTest) BurnArbGas(c ctx, evm mech, gasAmount huge) error {
	if !gasAmount.IsUint64() {
		return errors.New("Not a uint64")
	}
	//nolint:errcheck
	c.burn(gasAmount.Uint64()) // burn the amount, even if it's more than the user has
	return nil
}

func (con ArbosTest) GetAccountInfo(c ctx, evm mech, addr addr) error {
	return errors.New("unimplemented")
}

func (con ArbosTest) GetMarshalledStorage(c ctx, evm mech, addr addr) error {
	return errors.New("unimplemented")
}

func (con ArbosTest) InstallAccount(
	c ctx,
	evm mech,
	addr addr,
	isEOA bool,
	balance huge,
	nonce huge,
	code []byte,
	initStorage []byte,
) error {
	return errors.New("unimplemented")
}

func (con ArbosTest) SetNonce(c ctx, evm mech, addr addr, nonce huge) error {
	return errors.New("unimplemented")
}