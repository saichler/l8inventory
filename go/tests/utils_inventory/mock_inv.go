package utils_inventory

import (
	"fmt"
	"github.com/saichler/l8inventory/go/types"
)

// CreateMockNetworkDevices creates a slice of mock NetworkDevice elements
func CreateMockNetworkDevices(count int) []*types.NetworkDevice {
	devices := make([]*types.NetworkDevice, count)

	for i := 0; i < count; i++ {
		devices[i] = createMockNetworkDevice(i)
	}

	return devices
}

// createMockNetworkDevice creates a single mock NetworkDevice with all layers populated
func createMockNetworkDevice(index int) *types.NetworkDevice {
	deviceID := fmt.Sprintf("device-%d", index)
	deviceName := fmt.Sprintf("Switch-%d", index)

	return &types.NetworkDevice{
		Id:   deviceID,
		Name: deviceName,
		EquipmentInfo: &types.EquipmentInfo{
			Vendor:          "Cisco",
			Model:           "Catalyst 9300",
			Series:          "9300",
			Family:          "Catalyst",
			SoftwareVersion: "16.12.04",
			HardwareVersion: "V01",
			FirmwareVersion: "16.12.04",
			SystemName:      deviceName,
			SystemOid:       "1.3.6.1.4.1.9.1.2438",
			SerialNumber:    fmt.Sprintf("SN%08d", index),
			MacAddress:      fmt.Sprintf("00:1a:2b:3c:4d:%02x", index%256),
			IpAddresses:     []string{fmt.Sprintf("192.168.1.%d", index+1)},
			Location: &types.Location{
				Name:      fmt.Sprintf("DataCenter-%d", index/10),
				Latitude:  37.7749 + float64(index)*0.001,
				Longitude: -122.4194 + float64(index)*0.001,
			},
			Uptime: int64(86400 * (index + 1)), // Days in seconds
		},
		PhysicalLayer:   createMockPhysicalLayer(deviceID, index),
		LogicalLayer:    createMockLogicalLayer(deviceID, index),
		TechnologyLayer: createMockTechnologyLayer(deviceID, index),
		DeviceTree:      createMockDeviceTree(deviceID, deviceName),
		Relationships:   createMockRelationships(deviceID, index),
	}
}

// createMockPhysicalLayer creates a mock physical layer
func createMockPhysicalLayer(deviceID string, index int) *types.PhysicalLayer {
	chassisID := fmt.Sprintf("%s-chassis-0", deviceID)
	chassis := &types.Chassis{
		Id:           chassisID,
		Name:         "Main Chassis",
		SerialNumber: fmt.Sprintf("CHASSIS%08d", index),
		PartNumber:   "C9300-24T-A",
		ChassisType:  types.ChassisType_MAIN_CHASSIS,
		Status:       types.TreeNodeStatus_ACTIVE,
		Slots:        createMockSlots(chassisID, 2),
		PowerSupplies: []*types.PowerSupply{
			{
				Id:            fmt.Sprintf("%s-psu-0", chassisID),
				Name:          "Power Supply 0",
				SerialNumber:  fmt.Sprintf("PSU%08d", index),
				CapacityWatts: 715.0,
				CurrentDraw:   350.5,
				Status:        types.TreeNodeStatus_ACTIVE,
			},
		},
		Fans: []*types.Fan{
			{
				Id:       fmt.Sprintf("%s-fan-0", chassisID),
				Name:     "Fan 0",
				SpeedRpm: 2500,
				Status:   types.TreeNodeStatus_ACTIVE,
			},
		},
		Sensors: []*types.Sensor{
			{
				Id:                fmt.Sprintf("%s-temp-0", chassisID),
				Name:              "CPU Temperature",
				SensorType:        types.SensorType_TEMPERATURE,
				CurrentValue:      45.5,
				ThresholdWarning:  70.0,
				ThresholdCritical: 85.0,
				Units:             "Celsius",
				Status:            types.TreeNodeStatus_ACTIVE,
			},
		},
	}

	ports := createMockPorts(deviceID, 24)

	return &types.PhysicalLayer{
		Root: &types.TreeNode{
			Id:     fmt.Sprintf("%s-physical", deviceID),
			Name:   "Physical Layer",
			Type:   "PhysicalLayer",
			Status: types.TreeNodeStatus_ACTIVE,
			Attributes: map[string]string{
				"layer": "physical",
			},
		},
		ChassisList:     []*types.Chassis{chassis},
		PortList:        ports,
		CardList:        []*types.Card{},
		TransceiverList: []*types.Transceiver{},
	}
}

// createMockSlots creates mock slots for a chassis
func createMockSlots(chassisID string, count int) []*types.Slot {
	slots := make([]*types.Slot, count)

	for i := 0; i < count; i++ {
		slots[i] = &types.Slot{
			Id:         fmt.Sprintf("%s-slot-%d", chassisID, i),
			Name:       fmt.Sprintf("Slot %d", i),
			SlotNumber: int32(i),
			SlotType:   types.SlotType_LINE_CARD,
			Status:     types.TreeNodeStatus_ACTIVE,
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
			Id:         portID,
			Name:       fmt.Sprintf("GigabitEthernet1/0/%d", i+1),
			PortNumber: int32(i + 1),
			PortType:   types.PortType_ETHERNET,
			Speed:      types.PortSpeed_SPEED_1G,
			Duplex:     types.PortDuplex_FULL_DUPLEX,
			Status:     types.TreeNodeStatus_ACTIVE,
			Statistics: &types.PortStatistics{
				BytesIn:    int64(1000000 * (i + 1)),
				BytesOut:   int64(800000 * (i + 1)),
				PacketsIn:  int64(10000 * (i + 1)),
				PacketsOut: int64(8000 * (i + 1)),
			},
		}
	}

	return ports
}

// createMockLogicalLayer creates a mock logical layer
func createMockLogicalLayer(deviceID string, index int) *types.LogicalLayer {
	interfaces := createMockLogicalInterfaces(deviceID, 24)
	vlans := createMockVLANs(deviceID, 5)

	return &types.LogicalLayer{
		Root: &types.TreeNode{
			Id:     fmt.Sprintf("%s-logical", deviceID),
			Name:   "Logical Layer",
			Type:   "LogicalLayer",
			Status: types.TreeNodeStatus_ACTIVE,
			Attributes: map[string]string{
				"layer": "logical",
			},
		},
		InterfaceList: interfaces,
		VlanList:      vlans,
		VrfList:       []*types.VRF{},
		BridgeList:    []*types.Bridge{},
		TunnelList:    []*types.Tunnel{},
	}
}

// createMockLogicalInterfaces creates mock logical interfaces
func createMockLogicalInterfaces(deviceID string, count int) []*types.LogicalInterface {
	interfaces := make([]*types.LogicalInterface, count)

	for i := 0; i < count; i++ {
		interfaceID := fmt.Sprintf("%s-interface-%d", deviceID, i)
		interfaces[i] = &types.LogicalInterface{
			Id:            interfaceID,
			Name:          fmt.Sprintf("GigabitEthernet1/0/%d", i+1),
			InterfaceType: types.InterfaceType_PHYSICAL_INTERFACE,
			Description:   fmt.Sprintf("Port %d", i+1),
			IpAddresses:   []string{fmt.Sprintf("10.1.1.%d", i+10)},
			MacAddress:    fmt.Sprintf("00:1a:2b:3c:%02x:%02x", i/256, i%256),
			Mtu:           1500,
			Status:        types.TreeNodeStatus_ACTIVE,
			Statistics: &types.InterfaceStatistics{
				BytesIn:    int64(1000000 * (i + 1)),
				BytesOut:   int64(800000 * (i + 1)),
				PacketsIn:  int64(10000 * (i + 1)),
				PacketsOut: int64(8000 * (i + 1)),
			},
		}
	}

	return interfaces
}

// createMockVLANs creates mock VLANs
func createMockVLANs(deviceID string, count int) []*types.VLAN {
	vlans := make([]*types.VLAN, count)

	for i := 0; i < count; i++ {
		vlanID := fmt.Sprintf("%s-vlan-%d", deviceID, i)
		vlans[i] = &types.VLAN{
			Id:          vlanID,
			Name:        fmt.Sprintf("VLAN_%d", i+10),
			VlanId:      int32(i + 10),
			Description: fmt.Sprintf("VLAN %d", i+10),
			Status:      types.TreeNodeStatus_ACTIVE,
		}
	}

	return vlans
}

// createMockTechnologyLayer creates a mock technology layer
func createMockTechnologyLayer(deviceID string, index int) *types.TechnologyLayer {
	routingProtocols := []*types.RoutingProtocol{
		{
			Id:           fmt.Sprintf("%s-ospf", deviceID),
			Name:         "OSPF",
			ProtocolType: types.RoutingProtocolType_OSPF,
			Networks:     []string{"10.0.0.0/8", "192.168.0.0/16"},
			Configuration: map[string]string{
				"area":      "0",
				"router-id": fmt.Sprintf("1.1.1.%d", index+1),
			},
			Status: types.TreeNodeStatus_ACTIVE,
		},
	}

	switchingProtocols := []*types.SwitchingProtocol{
		{
			Id:           fmt.Sprintf("%s-stp", deviceID),
			Name:         "STP",
			ProtocolType: types.SwitchingProtocolType_STP,
			Configuration: map[string]string{
				"priority": "32768",
				"mode":     "pvst+",
			},
			Status: types.TreeNodeStatus_ACTIVE,
		},
	}

	managementProtocols := []*types.ManagementProtocol{
		{
			Id:           fmt.Sprintf("%s-snmp", deviceID),
			Name:         "SNMP",
			ProtocolType: types.ManagementProtocolType_SNMP,
			Configuration: map[string]string{
				"version":   "2c",
				"community": "public",
			},
			Status: types.TreeNodeStatus_ACTIVE,
		},
	}

	return &types.TechnologyLayer{
		Root: &types.TreeNode{
			Id:     fmt.Sprintf("%s-technology", deviceID),
			Name:   "Technology Layer",
			Type:   "TechnologyLayer",
			Status: types.TreeNodeStatus_ACTIVE,
			Attributes: map[string]string{
				"layer": "technology",
			},
		},
		RoutingProtocols:    routingProtocols,
		SwitchingProtocols:  switchingProtocols,
		SecurityProtocols:   []*types.SecurityProtocol{},
		QosProtocols:        []*types.QoSProtocol{},
		ManagementProtocols: managementProtocols,
	}
}

// createMockDeviceTree creates a mock device tree
func createMockDeviceTree(deviceID, deviceName string) *types.TreeNode {
	return &types.TreeNode{
		Id:     deviceID,
		Name:   deviceName,
		Type:   "NetworkDevice",
		Status: types.TreeNodeStatus_ACTIVE,
		Attributes: map[string]string{
			"device_type": "switch",
			"vendor":      "cisco",
		},
		Children: []*types.TreeNode{
			{
				Id:     fmt.Sprintf("%s-physical", deviceID),
				Name:   "Physical Layer",
				Type:   "PhysicalLayer",
				Status: types.TreeNodeStatus_ACTIVE,
				Depth:  1,
				Path:   fmt.Sprintf("%s/physical", deviceID),
			},
			{
				Id:     fmt.Sprintf("%s-logical", deviceID),
				Name:   "Logical Layer",
				Type:   "LogicalLayer",
				Status: types.TreeNodeStatus_ACTIVE,
				Depth:  1,
				Path:   fmt.Sprintf("%s/logical", deviceID),
			},
			{
				Id:     fmt.Sprintf("%s-technology", deviceID),
				Name:   "Technology Layer",
				Type:   "TechnologyLayer",
				Status: types.TreeNodeStatus_ACTIVE,
				Depth:  1,
				Path:   fmt.Sprintf("%s/technology", deviceID),
			},
		},
		Depth: 0,
		Path:  deviceID,
	}
}

// createMockRelationships creates mock relationships between components
func createMockRelationships(deviceID string, index int) []*types.Relationship {
	relationships := []*types.Relationship{
		{
			Id:               fmt.Sprintf("%s-rel-1", deviceID),
			SourceId:         deviceID,
			TargetId:         fmt.Sprintf("%s-chassis-0", deviceID),
			RelationshipType: types.RelationshipType_CONTAINS,
			Description:      "Device contains chassis",
			Attributes: map[string]string{
				"type": "containment",
			},
		},
		{
			Id:               fmt.Sprintf("%s-rel-2", deviceID),
			SourceId:         fmt.Sprintf("%s-port-0", deviceID),
			TargetId:         fmt.Sprintf("%s-interface-0", deviceID),
			RelationshipType: types.RelationshipType_CONNECTED_TO,
			Description:      "Port connected to interface",
			Attributes: map[string]string{
				"type": "connectivity",
			},
		},
	}

	// Add cross-device relationship if not the first device
	if index > 0 {
		relationships = append(relationships, &types.Relationship{
			Id:               fmt.Sprintf("%s-rel-peer", deviceID),
			SourceId:         deviceID,
			TargetId:         fmt.Sprintf("device-%d", index-1),
			RelationshipType: types.RelationshipType_CONNECTED_TO,
			Description:      "Connected to peer device",
			Attributes: map[string]string{
				"type": "peer-link",
			},
		})
	}

	return relationships
}
