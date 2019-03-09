package gpool

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

//Config pool config
type Config struct {
	//InitialPoolSize initial pool size. Default: 5
	InitialPoolSize int
	//MinPoolSize min item in pool. Default: 2
	MinPoolSize int
	//MaxPoolSize  max item in pool. Default: 15
	MaxPoolSize int
	//AcquireRetryAttempts retry times when get item Failed. Default: 5
	AcquireRetryAttempts int
	//AcquireIncrement  create count item when pool is empty. Default: 5
	AcquireIncrement int
	//TestDuration interval time between check item avaiable. Default: 1000
	TestDuration int
	//TestOnGetItem  test avaiable when get item. Default: false
	TestOnGetItem bool
	//Debug show debug information. Default: false
	Debug bool
	//Params item initial params
	Params map[string]string
}

//String String
func (config *Config) String() string {
	result := fmt.Sprintf("InitialPoolSize : %d \n MinPoolSize : %d \n MaxPoolSize : %d \n AcquireRetryAttempts : %d \n AcquireIncrement : %d \n TestDuration : %d \n TestOnGetItem : %t \n Debug : %t \n",
		config.InitialPoolSize,
		config.MinPoolSize,
		config.MaxPoolSize,
		config.AcquireRetryAttempts,
		config.AcquireIncrement,
		config.TestDuration,
		config.TestOnGetItem,
		config.Debug,
	)
	result = result + "Params:\n"
	for key, value := range config.Params {
		result = result + fmt.Sprintf("\t%s : %s \n", key, value)
	}
	return result
}

//DefaultConfig create default config
func DefaultConfig() *Config {
	return &Config{
		InitialPoolSize:      5,
		MinPoolSize:          2,
		MaxPoolSize:          15,
		AcquireRetryAttempts: 5,
		AcquireIncrement:     5,
		TestDuration:         1000,
		TestOnGetItem:        false,
		Debug:                false,
		Params:               make(map[string]string),
	}
}

//LoadToml load config from toml file
func (config *Config) LoadToml(tomlLocal string) error {
	inf, err := os.Stat(tomlLocal)
	if err != nil {
		log.Println("load toml config ERROR - FILE NOT EXIST")
		return err
	}
	if !strings.HasSuffix(inf.Name(), ".toml") {
		log.Println("load toml config ERROR - FILE TYPE ERROR")
		return errors.New("FILE TYPE ERROR")
	}
	_, err = toml.DecodeFile(tomlLocal, config)
	return err
}
