package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MapConf struct {
	MbusServerAddress  string `yaml:"mbus_server_address"`
	MbusRegAddress     uint16 `yaml:"mbus_register_address"`
	MbusRegDescription string `yaml:"mbus_register_description"`
	SnmpBaseOID        string `yaml:"snmp_base_oid"`
}

type Config struct {
	SnmpSrvAddr   string    `yaml:"snmp_server_address"`
	SnmpCommunity string    `yaml:"snmp_community"`
	Maps          []MapConf `yaml:"maps"`
}

func LoadConfig(configFile string) Config {
	f, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	var conf Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
	return conf
}
