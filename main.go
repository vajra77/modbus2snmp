package main

import (
	"fmt"
	"log"

	modbus "github.com/adibhanna/modbus-go"
	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
)

func main() {
	// Connect to MODBUS TCP server
	client := modbus.NewTCPClient("10.231.0.52:502")
	defer func(client *modbus.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)
	err := client.Connect()
	if err != nil {
		log.Fatal(err)
	}
	register := modbus.Address(10031)
	oids := []*GoSNMPServer.PDUValueControlItem{}
	oids = append(oids, &GoSNMPServer.PDUValueControlItem{
		OID:  fmt.Sprintf("1.3.6.1.2.1.2.2.1.1.%d", register),
		Type: gosnmp.Integer,
		OnGet: func() (value interface{}, err error) {
			values, err := client.ReadHoldingRegisters(register, 1)
			if err != nil {
				panic(err)
			}
			return GoSNMPServer.Asn1IntegerWrap(int(values[0])), nil
		},
		Document: "IfIndex",
	})

	master := GoSNMPServer.MasterAgent{
		Logger: GoSNMPServer.NewDefaultLogger(),
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			Users:                    []gosnmp.UsmSecurityParameters{},
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{"public"},
				OIDs:         oids,
			},
		},
	}
	server := GoSNMPServer.NewSNMPServer(master)
	err = server.ListenUDP("udp", "127.0.0.1:1161")
	if err != nil {
		panic(err)
	}
	err = server.ServeForever()
	if err != nil {
		return
	}
}
