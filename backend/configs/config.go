package configs

import (
	"fmt"

	"github.com/spf13/viper"
)

// type LogConfig struct {
// 	Level   string `yaml:"level"`
// 	Pattern string `yaml:"pattern"`
// 	OutPut  string `yaml:"output"`
// }

func CreateConfig(file string) *viper.Viper {
	config := viper.New()
	configPath := "config/"
	config.AddConfigPath(configPath) // 文件所在目录
	config.SetConfigName(file)       // 文件名
	config.SetConfigType("yaml")     // 文件类型
	configFile := configPath + file + ".yaml"

	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			panic(fmt.Errorf("找不到配置文件:%s", configFile)) //系统初始化阶段发生任何错误，直接结束进程
		} else {
			panic(fmt.Errorf("解析配置文件%s出错:%s", configFile, err))
		}
	}

	return config
}
