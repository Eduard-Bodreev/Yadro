package config

import (
	"flag"
	"log"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

var configPath string

type Config struct {
	SourceURL string `mapstructure:"source_url"`
	DBFile    string `mapstructure:"db_file"`
	IndexFile string `mapstructure:"index_file"`
	Parallel  int    `mapstructure:"parallel"`
	Port      string `mapstructure:"port"`
	DSN       string `mapstructure:"dsn"`
}

func init() {
	flag.StringVar(&configPath, "config", "./config/config.yaml", "path to config file")
	viper.SetDefault("source_url", "https://xkcd.com")
	viper.SetDefault("db_file", "database.json")
	viper.SetDefault("index_file", "index.json")
	viper.SetDefault("parallel", runtime.NumCPU())
	viper.SetDefault("port", "8080")
}

func InitConfig() Config {
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", configPath)
	}

	viper.SetConfigFile(configPath)
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
		IndexFile: viper.GetString("index_file"),
		Parallel:  parallel,
		Port:      viper.GetString("port"),
		DSN:       viper.GetString("dsn"),
	}
}
