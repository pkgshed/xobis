package xobis_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pkgshed/xobis"
)

func TestNetActivePowerQuantities(t *testing.T) {
	// Arrange. Power identifiers match the total and phase channels used by
	// smart-meter gateways. Group D selects power or its time integral.
	tests := []struct {
		name                string
		quantity            xobis.Quantity
		powerHex, energyHex string
	}{
		{"total", xobis.ElectricityActivePowerNet, "0100100700ff", "0100100800ff"},
		{"L1", xobis.ElectricityL1ActivePowerNet, "0100240700ff", "0100240800ff"},
		{"L2", xobis.ElectricityL2ActivePowerNet, "0100380700ff", "0100380800ff"},
		{"L3", xobis.ElectricityL3ActivePowerNet, "01004c0700ff", "01004c0800ff"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, mode := range []struct {
				name       string
				processing xobis.Processing
				hex        string
			}{
				{"instantaneous power", xobis.ElectricityInstantaneous, tt.powerHex},
				{"energy integral", xobis.ElectricityTimeIntegral1, tt.energyHex},
			} {
				t.Run(mode.name, func(t *testing.T) {
					// Arrange.
					want, err := xobis.ParseHex(mode.hex)
					require.NoError(t, err)

					// Act.
					code, err := xobis.NewCode(xobis.Groups{
						A: xobis.MediumElectricity,
						B: xobis.ChannelUnspecified,
						C: tt.quantity,
						D: mode.processing,
						E: xobis.ElectricityTariffTotal,
						F: xobis.StorageNotUsed,
					})

					// Assert.
					require.NoError(t, err)
					assert.Equal(t, want, code)
				})
			}
		})
	}
}

func ExampleQuantity_netActivePower() {
	// Arrange.
	groups := xobis.Groups{
		A: xobis.MediumElectricity,
		B: xobis.ChannelUnspecified,
		C: xobis.ElectricityL1ActivePowerNet,
		D: xobis.ElectricityInstantaneous,
		E: xobis.ElectricityTariffTotal,
		F: xobis.StorageNotUsed,
	}

	// Act. Group C stays the same when selecting the energy integral in D.
	power, err := xobis.NewCode(groups)
	if err != nil {
		panic(err)
	}
	groups.D = xobis.ElectricityTimeIntegral1
	energy, err := xobis.NewCode(groups)
	if err != nil {
		panic(err)
	}
	fmt.Println(power)
	fmt.Println(energy)

	// Output:
	// 1-0:36.7.0*255
	// 1-0:36.8.0*255
}
