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

	for _, m := range config.Maps {
		newMap := NewRegMap(m.MbusServerAddress,
			m.MbusRegAddress,
			m.MbusRegDescription,
			m.SnmpBaseOID)
		oids = append(oids, newMap.OID())
	}

	for _, mib := range mibImps.All() {
		oids = append(oids, mib)
	}

	master := GoSNMPServer.MasterAgent{
		Logger: GoSNMPServer.NewDefaultLogger(),
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			Users: []gosnmp.UsmSecurityParameters{
				{
					UserName:                 "modbus2snmp",
					AuthenticationProtocol:   gosnmp.MD5,
					PrivacyProtocol:          gosnmp.DES,
					AuthenticationPassphrase: "modbus2snmp",
					PrivacyPassphrase:        "modbus2snmp",
				},
			},
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{config.SnmpCommunity},
				OIDs:         oids,
			},
		},
	}
	server := GoSNMPServer.NewSNMPServer(master)
	err := server.ListenUDP("udp", config.SnmpSrvAddr)
	if err != nil {
		panic(err)
	}
	err = server.ServeForever()
	if err != nil {
		return
	}
}
