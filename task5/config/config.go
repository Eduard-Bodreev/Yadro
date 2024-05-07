package config

import (
	"flag"
	"log"
	"os"
	"runtime"

	"github.com/spf13/viper"
)

type Config struct {
	SourceURL string `mapstructure:"source_url"`
	DBFile    string `mapstructure:"db_file"`
	IndexFile string `mapstructure:"index_file"`
	Parallel  int    `mapstructure:"parallel"`
	Port      string `mapstructure:"port"`
}

var portFlag string

func init() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config/config.yaml", "path to config file")
	flag.StringVar(&portFlag, "p", "", "port to run the server on")
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", configPath)
	}

	viper.SetConfigFile(configPath)
	viper.SetDefault("source_url", "https://xkcd.com")
	viper.SetDefault("db_file", "database.json")
	viper.SetDefault("index_file", "index.json")
	viper.SetDefault("parallel", runtime.NumCPU())

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	if portFlag != "" {
		viper.Set("port", portFlag)
	} else {
		viper.SetDefault("port", "8080")
	}
}

func InitConfig() Config {
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
	}
}
