package flags

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config структура конфигурации
type Config struct {
	ServerAddress  string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

// var flags = pflag.NewFlagSet("flags", pflag.ExitOnError)

func init() {
	// Define the flags and bind them to viper
	pflag.StringP("ServerAddress", "a", "localhost:8080", "HTTP server network address")
	pflag.IntP("ReportInterval", "r", 10, "Interval between fetching reportable metrics in seconds")
	pflag.IntP("PollInterval", "p", 2, "Interval between polling metrics in seconds")

	// Parse the command-line flags
    pflag.Parse()

    // Check for unknown flags
    for _, arg := range pflag.Args() {
        if !strings.HasPrefix(arg, "-") {
            log.Fatalf("Unknown flag: %v", arg)
        }
    }

	// Bind the flags to viper
	bindFlagToViper("ServerAddress")
	bindFlagToViper("ReportInterval")
	bindFlagToViper("PollInterval")

	// Set the environment variable names
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvToViper("ServerAddress", "ADDRESS")
	bindEnvToViper("ReportInterval", "REPORT_INTERVAL")
	bindEnvToViper("PollInterval", "POLL_INTERVAL")

	// Read the environment variables
	viper.AutomaticEnv()
}

func bindFlagToViper(flagName string) {
	if err := viper.BindPFlag(flagName, pflag.Lookup(flagName)); err != nil {
		log.Println(err)
	}
}

func bindEnvToViper(viperKey, envKey string) {
	if err := viper.BindEnv(viperKey, envKey); err != nil {
		log.Println(err)
	}
}

// NewConfig создает новую конфигурацию
func NewConfig() *Config {
	return &Config{
		ServerAddress:  GetServerAddress(),
		ReportInterval: GetReportInterval(),
		PollInterval:   GetPollInterval(),
	}
}

// GetServerAddress возвращает адрес сервера
func GetServerAddress() string {
	return viper.GetString("ServerAddress")
}

// GetReportInterval возвращает интервал для отправки метрик
func GetReportInterval() time.Duration {
	return time.Duration(viper.GetInt("ReportInterval")) * time.Second
}

// GetPollInterval возвращает интервал получения метрик
func GetPollInterval() time.Duration {
	return time.Duration(viper.GetInt("PollInterval")) * time.Second
}