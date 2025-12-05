package main

import (
	"flag"

	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
	"github.com/slayercat/GoSNMPServer/mibImps"
)

func main() {

	confFile := flag.String("config", "mbus2snmp.yaml", "Configuration file")
	flag.Parse()

	config := LoadConfig(*confFile)

	var oids []*GoSNMPServer.PDUValueControlItem

	oids = append(oids, &GoSNMPServer.PDUValueControlItem{
		OID:  "1.3.6.1.2.1.1.1.0",
		Type: gosnmp.OctetString,
		OnGet: func() (value interface{}, err error) {
			return GoSNMPServer.Asn1OctetStringWrap(config.SNMPSysDescr), nil
		},
		Document: "ifIndex",
	})

	oids = append(oids, &GoSNMPServer.PDUValueControlItem{
		OID:  "1.3.6.1.2.1.1.2.0",
		Type: gosnmp.ObjectIdentifier,
		OnGet: func() (value interface{}, err error) {
			return GoSNMPServer.Asn1ObjectIdentifierWrap(config.SNMPObjectID), nil
		},
		Document: "ifIndex",
	})

	for _, mib := range mibImps.All() {
		oids = append(oids, mib)
	}

	for _, m := range config.Maps {
		newMap := NewRegMap(m.ModbusServerAddress,
			m.ModbusRegAddress,
			m.ModbusRegDescription,
			m.SNMPBaseOID)
		oids = append(oids, newMap.OID())
	}

	master := GoSNMPServer.MasterAgent{
		Logger: GoSNMPServer.NewDefaultLogger(),
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			Users:                    []gosnmp.UsmSecurityParameters{
				/*	{
					UserName:                 "modbus2snmp",
					AuthenticationProtocol:   gosnmp.MD5,
					PrivacyProtocol:          gosnmp.DES,
					AuthenticationPassphrase: "modbus2snmp",
					PrivacyPassphrase:        "modbus2snmp",
				},*/
			},
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{config.SNMPCommunity},
				OIDs:         oids,
			},
		},
	}
	server := GoSNMPServer.NewSNMPServer(master)
	err := server.ListenUDP("udp", config.SNMPSrvAddr)
	if err != nil {
		panic(err)
	}
	err = server.ServeForever()
	if err != nil {
		return
	}
}
