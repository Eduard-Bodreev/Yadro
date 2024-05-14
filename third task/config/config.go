package config

import (
	"log"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	SourceURL string `yaml:"source_url"`
	DBFile    string `yaml:"db_file"`
	Parallel  int    `yaml:"parallel"`
}

func InitConfig() Config {
	configFilePath, ok := os.LookupEnv("CONFIG_FILE")
	if !ok {
		panic("can't get config path from env")
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configFilePath)
	viper.SetDefault("source_url", "https://xkcd.com")
	viper.SetDefault("db_file", "database.json")
	viper.SetDefault("parallel", runtime.NumCPU())

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	parallel := viper.GetInt("parallel")
	if parallel <= 0 {
		parallel = runtime.NumCPU()
	}

	return Config{
		SourceURL: viper.GetString("source_url"),
		DBFile:    viper.GetString("db_file"),
		Parallel:  parallel,
	}
}
