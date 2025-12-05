package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MapConf struct {
	ModbusServerAddress  string `yaml:"mbus_server_address"`
	ModbusRegAddress     uint16 `yaml:"mbus_register_address"`
	ModbusRegDescription string `yaml:"mbus_register_description"`
	SNMPBaseOID          string `yaml:"snmp_base_oid"`
}

type Config struct {
	SNMPSrvAddr   string    `yaml:"snmp_server_address"`
	SNMPCommunity string    `yaml:"snmp_community"`
	SNMPObjectID  string    `yaml:"snmp_object_id"`
	SNMPSysDescr  string    `yaml:"snmp_sys_description"`
	Maps          []MapConf `yaml:"maps"`
}

func NewConfig(configFile string) (*Config, error) {

	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("error while opening file: %w", err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	var conf Config

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&conf); err != nil {
		return nil, fmt.Errorf("error while reading config: $w", err)
	}

	return &conf, nil
}
