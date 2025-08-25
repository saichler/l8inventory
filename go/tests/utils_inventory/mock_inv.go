package utils_inventory

import (
	"fmt"

	"github.com/saichler/probler/go/types"
)

// CreateMockNetworkDevices creates a slice of mock NetworkDevice elements
func CreateMockNetworkDevices(count int) []*types.NetworkDevice {
	devices := make([]*types.NetworkDevice, count)

	for i := 0; i < count; i++ {
		devices[i] = createMockNetworkDevice(i)
	}

	return devices
}

// createMockNetworkDevice creates a single mock NetworkDevice with new structure
func createMockNetworkDevice(index int) *types.NetworkDevice {
	deviceID := fmt.Sprintf("device-%d", index)

	return &types.NetworkDevice{
		Id: deviceID,
		Info: &types.EquipmentInfo{
			Vendor:          "Cisco",
			Model:           "Catalyst 9300",
			Series:          "9300",
			Family:          "Catalyst",
			Software:        "16.12.04",
			Hardware:        "V01",
			Version:         "16.12.04",
			SysName:         fmt.Sprintf("Switch-%d", index),
			SysOid:          "1.3.6.1.4.1.9.1.2438",
			SerialNumber:    fmt.Sprintf("SN%08d", index),
			FirmwareVersion: "16.12.04",
			IpAddress:       fmt.Sprintf("192.168.1.%d", index+1),
			Location:        fmt.Sprintf("DataCenter-%d", index/10),
			Latitude:        37.7749 + float64(index)*0.001,
			Longitude:       -122.4194 + float64(index)*0.001,
			Uptime:          fmt.Sprintf("%d", 86400*(index+1)),
		},
		Physicals: map[string]*types.Physical{
			"main": createMockPhysical(deviceID, index),
		},
		Logicals: map[string]*types.Logical{
			"main": createMockLogical(deviceID, index),
		},
	}
}

// createMockPhysical creates a mock Physical structure
func createMockPhysical(deviceID string, index int) *types.Physical {
	chassisID := fmt.Sprintf("%s-chassis-0", deviceID)

	return &types.Physical{
		Id: fmt.Sprintf("%s-physical", deviceID),
		Chassis: []*types.Chassis{
			{
				Id:           chassisID,
				SerialNumber: fmt.Sprintf("CHASSIS%08d", index),
				Slots:        createMockSlots(chassisID, 2),
			},
		},
		Ports: createMockPorts(deviceID, 24),
		PowerSupplies: []*types.PowerSupply{
			{
				Id:           fmt.Sprintf("%s-psu-0", chassisID),
				Name:         "Power Supply 0",
				SerialNumber: fmt.Sprintf("PSU%08d", index),
				Wattage:      715,
				Current:      350.5,
			},
		},
		Fans: []*types.Fan{
			{
				Id:       fmt.Sprintf("%s-fan-0", chassisID),
				Name:     "Fan 0",
				SpeedRpm: 2500,
			},
		},
	}
}

// createMockSlots creates mock slots for a chassis
func createMockSlots(chassisID string, count int) []*types.Slot {
	slots := make([]*types.Slot, count)

	for i := 0; i < count; i++ {
		slots[i] = &types.Slot{
			Id: fmt.Sprintf("%s-slot-%d", chassisID, i),
		}
	}

	return slots
}

// createMockPorts creates mock ports
func createMockPorts(deviceID string, count int) []*types.Port {
	ports := make([]*types.Port, count)

	for i := 0; i < count; i++ {
		portID := fmt.Sprintf("%s-port-%d", deviceID, i)
		ports[i] = &types.Port{
			Id: portID,
		}
	}

	return ports
}

// createMockLogical creates a mock Logical structure
func createMockLogical(deviceID string, index int) *types.Logical {
	return &types.Logical{
		Id:        fmt.Sprintf("%s-logical", deviceID),
		Intefaces: createMockInterfaces(deviceID, 24),
	}
}

// createMockInterfaces creates mock interfaces
func createMockInterfaces(deviceID string, count int) []*types.Interface {
	interfaces := make([]*types.Interface, count)

	for i := 0; i < count; i++ {
		interfaceID := fmt.Sprintf("%s-interface-%d", deviceID, i)
		interfaces[i] = &types.Interface{
			Id:          interfaceID,
			Name:        fmt.Sprintf("GigabitEthernet1/0/%d", i+1),
			Status:      "up",
			Description: fmt.Sprintf("Port %d", i+1),
			Speed:       uint64(1000000000), // 1Gbps
			MacAddress:  fmt.Sprintf("00:1a:2b:3c:%02x:%02x", i/256, i%256),
			IpAddress:   fmt.Sprintf("10.1.1.%d", i+10),
			Mtu:         1500,
			AdminStatus: true,
			Statistics: &types.InterfaceStatistics{
				RxBytes:   uint64(1000000 * (i + 1)),
				TxBytes:   uint64(800000 * (i + 1)),
				RxPackets: uint64(10000 * (i + 1)),
				TxPackets: uint64(8000 * (i + 1)),
			},
		}
	}

	return interfaces
}
