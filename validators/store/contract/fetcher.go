package contract

import (
	"fmt"

	"github.com/Gabulhas/polygon-external-consensus/contracts/staking"
	"github.com/Gabulhas/polygon-external-consensus/crypto"
	"github.com/Gabulhas/polygon-external-consensus/state"
	"github.com/Gabulhas/polygon-external-consensus/types"
	"github.com/Gabulhas/polygon-external-consensus/validators"
)

// FetchValidators fetches validators from a contract switched by validator type
func FetchValidators(
	validatorType validators.ValidatorType,
	transition *state.Transition,
	from types.Address,
) (validators.Validators, error) {
	switch validatorType {
	case validators.ECDSAValidatorType:
		return FetchECDSAValidators(transition, from)
	case validators.BLSValidatorType:
		return FetchBLSValidators(transition, from)
	}

	return nil, fmt.Errorf("unsupported validator type: %s", validatorType)
}

// FetchECDSAValidators queries a contract for validator addresses and returns ECDSAValidators
func FetchECDSAValidators(
	transition *state.Transition,
	from types.Address,
) (validators.Validators, error) {
	valAddrs, err := staking.QueryValidators(transition, from)
	if err != nil {
		return nil, err
	}

	ecdsaValidators := validators.NewECDSAValidatorSet()
	for _, addr := range valAddrs {
		if err := ecdsaValidators.Add(validators.NewECDSAValidator(addr)); err != nil {
			return nil, err
		}
	}

	return ecdsaValidators, nil
}

// FetchBLSValidators queries a contract for validator addresses & BLS Public Keys and returns ECDSAValidators
func FetchBLSValidators(
	transition *state.Transition,
	from types.Address,
) (validators.Validators, error) {
	valAddrs, err := staking.QueryValidators(transition, from)
	if err != nil {
		return nil, err
	}

	blsPublicKeys, err := staking.QueryBLSPublicKeys(transition, from)
	if err != nil {
		return nil, err
	}

	blsValidators := validators.NewBLSValidatorSet()

	for idx := range valAddrs {
		// ignore the validator whose BLS Key is not set
		// because BLS validator needs to have both Address and BLS Public Key set
		// in the contract
		if _, err := crypto.UnmarshalBLSPublicKey(blsPublicKeys[idx]); err != nil {
			continue
		}

		if err := blsValidators.Add(validators.NewBLSValidator(
			valAddrs[idx],
			blsPublicKeys[idx],
		)); err != nil {
			return nil, err
		}
	}

	return blsValidators, nil
}
