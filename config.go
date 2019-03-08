package gpool

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

//Config 连接池配置
type Config struct {
	//InitialPoolSize 初始化池中元素数量，取值应在MinPoolSize与MaxPoolSize之间 Default: 5
	InitialPoolSize int
	//MinPoolSize 池中保留的最小元素数量 Default: 2
	MinPoolSize int
	//MaxPoolSize 池中保留的最大连元素数量 Default: 15
	MaxPoolSize int
	//AcquireRetryAttempts 定义在新连接失败后重复尝试的次数 Default: 5
	AcquireRetryAttempts int
	//AcquireIncrement 当池中的元素耗尽时，一次同时创建的元素数 Default: 5
	AcquireIncrement int
	//TestDuration 连接有效性检查间隔，单位毫秒 Default: 1000
	TestDuration int
	//TestOnGetItem 如果设为true那么在取得元素的同时将校验元素的有效性 Default: false
	TestOnGetItem bool
	//Debug 显示调试信息 Default: false
	Debug bool
	//Params 元素初始化参数
	Params map[string]string
}

//String String
func (config *Config) String() string {
	result := fmt.Sprintf("InitialPoolSize : %d \n MinPoolSize : %d \n MaxPoolSize : %d \n AcquireRetryAttempts : %d \n AcquireIncrement : %d \n TestDuration : %d \n TestOnGetItem : %t \n Debug : %t ",
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

//DefaultConfig 创建默认的配置
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

//LoadToml 从toml文件中初始化配置
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
