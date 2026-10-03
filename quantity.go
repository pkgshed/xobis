package xobis

import "fmt"

// Phase selects an electricity measurement block in OBIS value group C.
// It is distinct from the medium in group A and the channel in group B.
type Phase uint8

const (
	// Aggregate selects the unshifted base quantity. Its meaning depends on the
	// quantity: total across phases for net active power, any phase for current
	// and voltage. It does not imply that values are summed across phases.
	Aggregate Phase = iota
	L1
	L2
	L3
)

const electricalPhaseOffset Quantity = 20

// ElectricalQuantityFor derives a group C quantity for [MediumElectricity] as
// base + phase*20. It accepts [Aggregate], [L1], [L2] or [L3] and only the base
// measurement family 1..20, including [ElectricityCurrent], [ElectricityVoltage]
// and [ElectricityActivePowerNet]. Already phase-specific quantities and other
// group C families are rejected, even with [Aggregate].
//
// The result describes group C only; use [MediumElectricity] for group A.
// [ElectricityActivePowerNet] retains its signed import-minus-export meaning.
// On error, it returns zero and an error wrapping [ErrRange].
func ElectricalQuantityFor(phase Phase, base Quantity) (Quantity, error) {
	if phase > L3 {
		return 0, fmt.Errorf("xobis: electricity phase %d must be in 0..3: %w", phase, ErrRange)
	}
	if base < 1 || base > 20 {
		return 0, fmt.Errorf("xobis: base electricity quantity %d must be in 1..20: %w", base, ErrRange)
	}
	return base + Quantity(phase)*electricalPhaseOffset, nil
}
