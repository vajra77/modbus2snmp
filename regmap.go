package main

import (
	"fmt"
	"log"

	"github.com/adibhanna/modbus-go"
	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
)

type RegMap struct {
	value          uint16
	MbusSrvAddress string
	MbusRegAddress modbus.Address
	MbusConversion uint16
	MbusUnit       string
	SnmpBaseOID    string
}

func NewRegMap(srvAddr string, regAddr uint16, conv uint16, unit string, base string) *RegMap {
	return &RegMap{
		value:          0,
		MbusSrvAddress: srvAddr,
		MbusRegAddress: modbus.Address(regAddr),
		MbusConversion: conv,
		MbusUnit:       unit,
		SnmpBaseOID:    base,
	}
}

func (reg *RegMap) Value() uint16 {
	return reg.value
}

func (reg *RegMap) Float() float32 {
	return float32(reg.value / reg.MbusConversion)
}

func (reg *RegMap) String() string {
	return fmt.Sprintf("%.2f %s", reg.Float(), reg.MbusUnit)
}

func (reg *RegMap) Read() error {
	client := modbus.NewTCPClient(reg.MbusSrvAddress)
	defer func(client *modbus.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)

	err := client.Connect()
	if err != nil {
		return err
	}

	values, err := client.ReadHoldingRegisters(reg.MbusRegAddress, 1)
	if err != nil {
		reg.value = 0
		return err
	}

	reg.value = values[0]
	return nil
}

func (reg *RegMap) OID() *GoSNMPServer.PDUValueControlItem {
	return &GoSNMPServer.PDUValueControlItem{
		OID:  fmt.Sprintf("%s.%d", reg.SnmpBaseOID, reg.MbusRegAddress),
		Type: gosnmp.OpaqueFloat,
		OnGet: func() (value interface{}, err error) {
			err = reg.Read()
			if err != nil {
				return 0, err
			}
			return GoSNMPServer.Asn1OpaqueFloatWrap(reg.Float()), nil
		},
		Document: "IfIndex",
	}
}
