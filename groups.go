package xobis

import "fmt"

// Groups holds mutable, typed values for OBIS groups A through F.
// It is construction input, not a validated identifier. [NewCode] and [NewPattern]
// validate and copy it; subsequent edits do not affect the constructed value.
type Groups struct {
	A Medium
	B Channel
	C Quantity
	D Processing
	E Classification
	F Storage
}

// Validate checks the numeric ranges of all six groups: A must be in 0..15,
// and B through F are constrained to 0..255 by their types. It does not check
// whether a combination is assigned by a standard or a manufacturer.
// [NewPattern] validates only selected groups and ignores omitted values.
func (g Groups) Validate() error {
	if g.A > 15 {
		return fmt.Errorf("xobis: group A must be in 0..15: %w", ErrRange)
	}
	return nil
}

// Medium identifies the energy type or abstract data in group A.
// Values in 0..15 are in range, including reserved values without a named constant.
//
//go:generate go tool -modfile=tools/tools.mod github.com/0x5a17ed/stringer/v2 -output medium_string.go -enums Medium
type Medium uint8

// Media identifiers. Cooling and heat retain their conventional names;
// IEC 62056-6-1:2023 groups both under thermal energy.
const (
	MediumAbstract          Medium = 0
	MediumElectricity       Medium = 1
	MediumHeatCostAllocator Medium = 4
	MediumCooling           Medium = 5
	MediumHeat              Medium = 6
	MediumGas               Medium = 7
	MediumColdWater         Medium = 8
	MediumHotWater          Medium = 9
	MediumOther             Medium = 15
)

// Channel identifies the measurement or communication channel in group B.
// Values 1..64 are channel numbers, 65..127 are utility specific, 128..199
// are manufacturer specific, and 200..255 are reserved.
type Channel uint8

// ChannelUnspecified indicates that no channel has been specified.
const ChannelUnspecified Channel = 0

// Quantity identifies the physical or abstract data item in group C.
// Its interpretation depends on [Medium]. Electricity-prefixed constants apply
// to electricity objects only; values for other media remain representable.
type Quantity uint8

// Common electricity quantities. Power quantities become energy quantities
// when combined with an appropriate time integral in group D. These constants
// are a convenience subset, not an exhaustive list of assigned quantities.
const (
	ElectricityActivePowerImport   Quantity = 1
	ElectricityActivePowerExport   Quantity = 2
	ElectricityReactivePowerImport Quantity = 3
	ElectricityReactivePowerExport Quantity = 4
	ElectricityReactivePowerQI     Quantity = 5
	ElectricityReactivePowerQII    Quantity = 6
	ElectricityReactivePowerQIII   Quantity = 7
	ElectricityReactivePowerQIV    Quantity = 8
	ElectricityApparentPowerImport Quantity = 9
	ElectricityApparentPowerExport Quantity = 10
	ElectricityCurrent             Quantity = 11
	ElectricityVoltage             Quantity = 12
	ElectricityPowerFactor         Quantity = 13
	ElectricityFrequency           Quantity = 14
	ElectricityL1Current           Quantity = 31
	ElectricityL1Voltage           Quantity = 32
	ElectricityL2Current           Quantity = 51
	ElectricityL2Voltage           Quantity = 52
	ElectricityL3Current           Quantity = 71
	ElectricityL3Voltage           Quantity = 72
)

// Processing identifies the processing or classification in group D.
// Its interpretation depends on [Medium] and [Quantity].
type Processing uint8

// Common processing modes for electricity measurement quantities.
// These names do not apply to service, error, list or profile objects.
const (
	ElectricityBillingPeriodAverage Processing = 0
	ElectricityCumulativeMinimum1   Processing = 1
	ElectricityCumulativeMaximum1   Processing = 2
	ElectricityMinimum1             Processing = 3
	ElectricityCurrentAverage1      Processing = 4
	ElectricityLastAverage1         Processing = 5
	ElectricityMaximum1             Processing = 6
	ElectricityInstantaneous        Processing = 7
	ElectricityTimeIntegral1        Processing = 8
	ElectricityTimeIntegral2        Processing = 9
)

// Classification identifies further processing or classification in group E.
// Depending on groups A through D, this can be a tariff, harmonic, or another
// classification. It is deliberately not named Tariff.
type Classification uint8

// Common tariff selectors for electricity energy and demand registers.
// These constants do not describe the harmonic or other interpretations of E.
const (
	ElectricityTariffTotal Classification = 0
	ElectricityTariff1     Classification = 1
	ElectricityTariff2     Classification = 2
	ElectricityTariff3     Classification = 3
	ElectricityTariff4     Classification = 4
)

// Storage identifies historical data or further classification in group F.
// It is not a chronological index for all object kinds.
type Storage uint8

// StorageNotUsed is used when group F has no further classification.
// For billing-period measurement registers, 255 selects the current period.
// It remains an explicit value when matching patterns, not a wildcard.
const StorageNotUsed Storage = 255
