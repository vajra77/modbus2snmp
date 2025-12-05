package main

import (
	"fmt"
	"log"

	"github.com/adibhanna/modbus-go"
	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
)

type RegMap struct {
	value                uint16
	ModbusSrvAddress     string
	ModbusRegAddress     modbus.Address
	ModbusRegDescription string
	SNMPBaseOID          string
}

func NewRegMap(srvAddr string, regAddr uint16, descr string, base string) *RegMap {
	return &RegMap{
		value:                0,
		ModbusSrvAddress:     srvAddr,
		ModbusRegAddress:     modbus.Address(regAddr),
		ModbusRegDescription: descr,
		SNMPBaseOID:          base,
	}
}

func (reg *RegMap) Value() uint {
	return uint(reg.value)
}

func (reg *RegMap) Read() error {
	client := modbus.NewTCPClient(reg.ModbusSrvAddress)
	defer func() {
		err := client.Close()
		if err != nil {
			log.Printf("failed to close modbus client: %v", err)
		}
	}()

	err := client.Connect()
	if err != nil {
		return err
	}

	values, err := client.ReadHoldingRegisters(reg.ModbusRegAddress, 1)
	if err != nil {
		reg.value = 0
		return err
	}

	reg.value = values[0]
	return nil
}

func (reg *RegMap) OID() *GoSNMPServer.PDUValueControlItem {
	return &GoSNMPServer.PDUValueControlItem{
		OID:  fmt.Sprintf("%s.%d", reg.SNMPBaseOID, reg.ModbusRegAddress),
		Type: gosnmp.Gauge32,
		OnGet: func() (value interface{}, err error) {
			err = reg.Read()
			if err != nil {
				return 0, err
			}
			return GoSNMPServer.Asn1Gauge32Wrap(reg.Value()), nil
		},
		Document: "ifIndex",
	}
}
