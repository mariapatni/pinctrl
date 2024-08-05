//go:build linux

package pi5

import (
	"context"
	"testing"

	"go.viam.com/test"

	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

func TestEmptyBoard(t *testing.T) {
	b := &pinctrlpi5{
		logger: logging.NewTestLogger(t),
	}

	t.Run("test empty sysfs board", func(t *testing.T) {
		_, err := b.GPIOPinByName("10")
		test.That(t, err, test.ShouldNotBeNil)
	})
}

func TestDeviceTreeParsing(t *testing.T) {
	logger := logging.NewTestLogger(t)
	ctx := context.Background()

	// Create a fake board mapping with two pins for testing.
	// These are needed for making board; not used for pin control testing yet.
	testBoardMappings := make(map[string]GPIOBoardMapping, 2)
	testBoardMappings["1"] = GPIOBoardMapping{
		GPIOChipDev:    "gpiochip0",
		GPIO:           12,
		GPIOName:       "12",
		PWMSysFsDir:    "",
		PWMID:          1,
		HWPWMSupported: true,
	}

	conf := &Config{}
	config := resource.Config{
		Name:                "board1",
		ConvertedAttributes: conf,
	}

	// Test Creations of Boards
	newB, err := NewBoard(ctx, config, ConstPinDefs(testBoardMappings), logger, true)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, newB, test.ShouldNotBeNil)

	// Cast from board.Board to pinctrlpi5 is required to access board's vars
	p5 := newB.(*pinctrlpi5)
	test.That(t, p5.chipSize, test.ShouldEqual, 0x30000)
	test.That(t, p5.physAddr, test.ShouldEqual, 0x1f000d0000)

	// Test pinctrl_utils.go helper functions:
	path := "/axi/pcie@120000/rp1/gpio@d0000"
	alias, err := p5.findPathFromAlias("gpio0")
	test.That(t, err, test.ShouldBeNil)
	test.That(t, cleanFilePath(alias), test.ShouldEqual, cleanFilePath(path))

	byteArray := []byte{0x01, 0x02, 0x03, 0x04}
	val, err := parseCells(0, &byteArray)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, val, test.ShouldEqual, 0x0)

	val, err = parseCells(1, &byteArray)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldEqual, 0x0000000001020304)

	val, err = parseCells(2, &byteArray)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, val, test.ShouldEqual, 0x0)

	byteArray = []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	val, err = parseCells(2, &byteArray)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldEqual, 0x102030405060708)

	// this parseCells() call should omit the first 32 bits from the final outputted value
	byteArray = []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	val, err = parseCells(3, &byteArray)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, val, test.ShouldEqual, 0x102030405060708)

	// this should throw an error since there are no address/size cell files here
	path = "./test-device-tree/proc/device-tree/axi/pcie@120000/rp1/gpio@d0000"
	addr, size, err := getNumAddrSizeCells(path)
	test.That(t, err, test.ShouldNotBeNil)
	test.That(t, addr, test.ShouldEqual, 0)
	test.That(t, size, test.ShouldEqual, 0)

	path = "./test-device-tree/proc/device-tree/axi/pcie@120000/rp1"
	addr, size, err = getNumAddrSizeCells(path)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, addr, test.ShouldEqual, 2)
	test.That(t, size, test.ShouldEqual, 2)

	path = "./test-device-tree/proc/device-tree/axi/pcie@120000"
	addr, size, err = getNumAddrSizeCells(path)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, addr, test.ShouldEqual, 3)
	test.That(t, size, test.ShouldEqual, 2)

	path = "./test-device-tree/proc/device-tree/axi/pcie@120000/rp1/gpio@d0000"
	reg, err := getRegAddr(path, 2)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, reg, test.ShouldEqual, 0xC0400D0000)

	path = "./test-device-tree/proc/device-tree/axi/pcie@120000/rp1"
	ranges, err := getRangesAddr(path, 3, 2, 2)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, ranges[0].childAddr, test.ShouldEqual, 0x4000000002000000)
	test.That(t, ranges[0].parentAddr, test.ShouldEqual, 0x0)
	test.That(t, ranges[0].addrSpaceSize, test.ShouldEqual, 0x400000)

	path = "./test-device-tree/proc/device-tree/axi/pcie@120000"
	ranges, err = getRangesAddr(path, 3, 2, 2)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, ranges[0].childAddr, test.ShouldEqual, 0x0)
	test.That(t, ranges[0].parentAddr, test.ShouldEqual, 0x1F00000000)
	test.That(t, ranges[0].addrSpaceSize, test.ShouldEqual, 0xFFFFFFFC)
	test.That(t, ranges[1].childAddr, test.ShouldEqual, 0x400000000)
	test.That(t, ranges[1].parentAddr, test.ShouldEqual, 0x1C00000000)
	test.That(t, ranges[1].addrSpaceSize, test.ShouldEqual, 0x300000000)

	defer newB.Close(ctx)
}