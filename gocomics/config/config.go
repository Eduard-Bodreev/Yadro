package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	SourceURL string `yaml:"source_url"`
	DBFile    string `yaml:"db_file"`
}

func InitConfig() {
	configFilePath, ok := os.LookupEnv("CONFIG_FILE")
	if !ok {
		panic("can't get config path from env")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)
	viper.SetDefault("source_url", "https://xkcd.com")
	viper.SetDefault("db_file", "database.json")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file, %s", err)
	}
}
