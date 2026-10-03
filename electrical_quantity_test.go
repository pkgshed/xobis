package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestElectricalQuantityFor(t *testing.T) {
	// Arrange. Expected values are the OBIS group C assignments, independent of
	// the derivation. Existing exported constants must retain those assignments.
	tests := []struct {
		name    string
		base    xobis.Quantity
		want    [4]xobis.Quantity
		aliases [4]xobis.Quantity
	}{
		{"current", xobis.ElectricityCurrent, [4]xobis.Quantity{11, 31, 51, 71},
			[4]xobis.Quantity{xobis.ElectricityCurrent, xobis.ElectricityL1Current, xobis.ElectricityL2Current, xobis.ElectricityL3Current}},
		{"voltage", xobis.ElectricityVoltage, [4]xobis.Quantity{12, 32, 52, 72},
			[4]xobis.Quantity{xobis.ElectricityVoltage, xobis.ElectricityL1Voltage, xobis.ElectricityL2Voltage, xobis.ElectricityL3Voltage}},
		{"net active power", xobis.ElectricityActivePowerNet, [4]xobis.Quantity{16, 36, 56, 76},
			[4]xobis.Quantity{xobis.ElectricityActivePowerNet, xobis.ElectricityL1ActivePowerNet, xobis.ElectricityL2ActivePowerNet, xobis.ElectricityL3ActivePowerNet}},
	}
	phases := []struct {
		name  string
		phase xobis.Phase
	}{
		{"aggregate", xobis.Aggregate}, {"L1", xobis.L1}, {"L2", xobis.L2}, {"L3", xobis.L3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, phase := range phases {
				t.Run(phase.name, func(t *testing.T) {
					// Act.
					quantity, err := xobis.ElectricalQuantityFor(phase.phase, tt.base)

					// Assert.
					require.NoError(t, err)
					assert.Equal(t, tt.want[i], quantity)
					assert.Equal(t, tt.want[i], tt.aliases[i])
				})
			}
		})
	}
}

func TestElectricalQuantityForMeasurementFamily(t *testing.T) {
	// Arrange. The standard also defines phase blocks for these base quantities,
	// including both boundaries of the supported 1..20 family.
	tests := []struct {
		name string
		base xobis.Quantity
		want [4]xobis.Quantity
	}{
		{"active power import", xobis.ElectricityActivePowerImport, [4]xobis.Quantity{1, 21, 41, 61}},
		{"active power export", xobis.ElectricityActivePowerExport, [4]xobis.Quantity{2, 22, 42, 62}},
		{"reactive power import", xobis.ElectricityReactivePowerImport, [4]xobis.Quantity{3, 23, 43, 63}},
		{"reactive power export", xobis.ElectricityReactivePowerExport, [4]xobis.Quantity{4, 24, 44, 64}},
		{"power factor", xobis.ElectricityPowerFactor, [4]xobis.Quantity{13, 33, 53, 73}},
		{"frequency", xobis.ElectricityFrequency, [4]xobis.Quantity{14, 34, 54, 74}},
		{"absolute active power", xobis.Quantity(15), [4]xobis.Quantity{15, 35, 55, 75}},
		{"active power quadrant IV", xobis.Quantity(20), [4]xobis.Quantity{20, 40, 60, 80}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, phase := range []xobis.Phase{xobis.Aggregate, xobis.L1, xobis.L2, xobis.L3} {
				// Act.
				quantity, err := xobis.ElectricalQuantityFor(phase, tt.base)

				// Assert.
				require.NoError(t, err)
				assert.Equal(t, tt.want[i], quantity, "phase %d", phase)
			}
		})
	}
}

func TestElectricalQuantityForInvalidInputs(t *testing.T) {
	// Arrange. These inputs include unrelated families, quantities that would
	// be shifted twice, and phases whose offset arithmetic would overflow uint8.
	tests := []struct {
		name  string
		phase xobis.Phase
		base  xobis.Quantity
	}{
		{"phase beyond L3", xobis.Phase(4), xobis.ElectricityCurrent},
		{"phase offset overflow", xobis.Phase(13), xobis.ElectricityActivePowerNet},
		{"maximum phase", xobis.Phase(255), xobis.ElectricityVoltage},
		{"zero base", xobis.L1, xobis.Quantity(0)},
		{"already shifted import", xobis.L1, xobis.Quantity(21)},
		{"already shifted current", xobis.L1, xobis.ElectricityL1Current},
		{"already shifted aggregate", xobis.Aggregate, xobis.ElectricityL2Voltage},
		{"angles", xobis.L1, xobis.Quantity(81)},
		{"neutral current", xobis.L3, xobis.Quantity(91)},
		{"service object", xobis.L2, xobis.Quantity(96)},
		{"manufacturer quantity", xobis.L1, xobis.Quantity(128)},
		{"quantity overflow", xobis.L3, xobis.Quantity(255)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act.
			quantity, err := xobis.ElectricalQuantityFor(tt.phase, tt.base)

			// Assert.
			assert.ErrorIs(t, err, xobis.ErrRange)
			assert.Zero(t, quantity)
		})
	}
}

func ExampleElectricalQuantityFor() {
	// Arrange.
	quantity, err := xobis.ElectricalQuantityFor(xobis.L3, xobis.ElectricityActivePowerNet)
	if err != nil {
		panic(err)
	}

	// Act. Medium, phase and processing occupy separate concepts and groups.
	code, err := xobis.NewCode(xobis.Groups{
		A: xobis.MediumElectricity,
		B: xobis.ChannelUnspecified,
		C: quantity,
		D: xobis.ElectricityInstantaneous,
		E: xobis.ElectricityTariffTotal,
		F: xobis.StorageNotUsed,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(quantity)
	fmt.Println(code)
	fmt.Println(xobis.ToHex(code))

	// Output:
	// 76
	// 1-0:76.7.0*255
	// 01004c0700ff
}
