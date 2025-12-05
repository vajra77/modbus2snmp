package main

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/adibhanna/modbus-go"
	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
)

type RegMap struct {
	mu                   sync.RWMutex
	client               *modbus.Client
	value                uint16
	ModbusSrvAddress     string
	ModbusRegAddress     modbus.Address
	ModbusRegDescription string
	SNMPBaseOID          string
}

func NewRegMap(srvAddr string, regAddr uint16, descr string, base string) *RegMap {
	return &RegMap{
		mu:                   sync.RWMutex{},
		client:               modbus.NewTCPClient(srvAddr),
		value:                0,
		ModbusSrvAddress:     srvAddr,
		ModbusRegAddress:     modbus.Address(regAddr),
		ModbusRegDescription: descr,
		SNMPBaseOID:          base,
	}
}

func (reg *RegMap) Value() uint32 {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	return uint32(reg.value)
}

func (reg *RegMap) Read() error {
	if reg.client == nil {
		return errors.New("client not initialized")
	}
	if err := reg.client.Connect(); err != nil {
		return fmt.Errorf("could not connect to modbus: %w", err)
	}
	defer func() {
		err := reg.client.Close()
		if err != nil {
			log.Printf("failed to close modbus client: %v", err)
		}
	}()

	values, err := reg.client.ReadHoldingRegisters(reg.ModbusRegAddress, 1)
	if err != nil {
		reg.value = 0
		return fmt.Errorf("reading register %d: %w", reg.ModbusRegAddress, err)
	}

	reg.mu.Lock()
	reg.value = values[0]
	reg.mu.Unlock()

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
			return GoSNMPServer.Asn1Gauge32Wrap(uint(reg.Value())), nil
		},
		Document: reg.ModbusRegDescription,
	}
}
