package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gosnmp/gosnmp"
	"github.com/slayercat/GoSNMPServer"
	"github.com/slayercat/GoSNMPServer/mibImps"
)

const (
	// sysDescr - System description
	oidSysDescr = "1.3.6.1.2.1.1.1.0"
	// sysObjectID - System object identifier
	oidSysObjectID = "1.3.6.1.2.1.1.2.0"
)

func run() error {

	confFile := flag.String("config", "mbus2snmp.yaml", "Configuration file")
	flag.Parse()

	log.Printf("loading configuration from %s", *confFile)
	config, err := NewConfig(*confFile)
	if err != nil {
		log.Fatalf("failed to load config from %s: %v", *confFile, err)
	}

	oids, cleanup, err := buildOIDsWithCleanup(config)
	if err != nil {
		return fmt.Errorf("build OIDs: %w", err)
	}
	defer cleanup()

	master := buildMasterAgent(config, oids)
	server := GoSNMPServer.NewSNMPServer(master)

	if err := server.ListenUDP("udp", config.SNMPSrvAddr); err != nil {
		return fmt.Errorf("listen faild: %w", err)
	}

	return serveWithGracefulShutdown(server, config.SNMPSrvAddr)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func buildMasterAgent(config *Config, oids []*GoSNMPServer.PDUValueControlItem) GoSNMPServer.MasterAgent {

	master := GoSNMPServer.MasterAgent{
		Logger: GoSNMPServer.NewDefaultLogger(),
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
			Users:                    []gosnmp.UsmSecurityParameters{},
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{config.SNMPCommunity},
				OIDs:         oids,
			},
		},
	}

	return master
}

func buildOIDsWithCleanup(config *Config) ([]*GoSNMPServer.PDUValueControlItem, func(), error) {
	estimatedSize := 2 + len(mibImps.All()) + len(config.Maps)
	oids := make([]*GoSNMPServer.PDUValueControlItem, 0, estimatedSize)
	regMaps := make([]*RegMap, 0, len(config.Maps))

	cleanup := func() {
		for _, rm := range regMaps {
			if err := rm.Close(); err != nil {
				log.Printf("error closing RegMap: %v", err)
			}
		}
	}

	// System OIDs
	oids = append(oids, &GoSNMPServer.PDUValueControlItem{
		OID:  oidSysDescr,
		Type: gosnmp.OctetString,
		OnGet: func() (interface{}, error) {
			return GoSNMPServer.Asn1OctetStringWrap(config.SNMPSysDescr), nil
		},
		Document: "System Description",
	})

	oids = append(oids, &GoSNMPServer.PDUValueControlItem{
		OID:  oidSysObjectID,
		Type: gosnmp.ObjectIdentifier,
		OnGet: func() (interface{}, error) {
			return GoSNMPServer.Asn1ObjectIdentifierWrap(config.SNMPObjectID), nil
		},
		Document: "System Object Identifier",
	})

	// MIB imports
	oids = append(oids, mibImps.All()...)

	// Modbus register mappings
	for _, m := range config.Maps {
		regMap := NewRegMap(
			m.ModbusServerAddress,
			m.ModbusRegAddress,
			m.ModbusRegDescription,
			m.SNMPBaseOID,
		)
		regMaps = append(regMaps, regMap)
		oids = append(oids, regMap.OID())
	}

	return oids, cleanup, nil
}

func serveWithGracefulShutdown(server *GoSNMPServer.SNMPServer, addr string) error {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		log.Printf("SNMP server listening on %s", addr)
		errChan <- server.ServeForever()
	}()

	select {
	case <-sigChan:
		log.Println("received shutdown signal")
		cancel()
		// Give server time to cleanup
		return nil
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("server stopped: %w", err)
		}
		return nil
	}
}
