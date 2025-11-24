package main

import (
	"flag"

	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
)

func main() {

	confFile := flag.String("config", "mbus2snmp.yaml", "Configuration file")
	flag.Parse()

	config := LoadConfig(*confFile)

	var oids []*GoSNMPServer.PDUValueControlItem

	for _, m := range config.Maps {
		newMap := NewRegMap(m.MbusServerAddress,
			m.MbusRegAddress,
			m.MbusConversion,
			m.MbusUnit,
			m.SnmpBaseOID)
		oids = append(oids, newMap.OID())
	}
	master := GoSNMPServer.MasterAgent{
		//Logger: GoSNMPServer.NewDefaultLogger(),
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			Users:                    []gosnmp.UsmSecurityParameters{},
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
