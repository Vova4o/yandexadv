package flags

import (
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config структура конфигурации
type Config struct {
	ServerAddress   string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	ServerLogFile   string
	DBDSN           string
	SecretKey       string
}

// GetFlags устанавливает и получает флаги
func GetFlags() {
	// Set the environment variable names
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvToViper("DatabaseDSN", "DATABASE_DSN")
	bindEnvToViper("ServerAddress", "ADDRESS")
	bindEnvToViper("StoreInterval", "STORE_INTERVAL")
	bindEnvToViper("FileStoragePath", "FILE_STORAGE_PATH")
	bindEnvToViper("Restore", "RESTORE")
	bindEnvToViper("ServerLoggerFile", "SERVER_LOGGER_FILE")
	bindEnvToViper("Key", "KEY")

	// Read the environment variables
	viper.AutomaticEnv()

	// postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable
	// metrics-db.json

	// Define the flags and bind them to viper
	pflag.StringP("DatabaseDSN", "d", "", "Database DSN")
	pflag.StringP("ServerAddress", "a", "localhost:8080", "HTTP server network address")
	pflag.IntP("StoreInterval", "i", 300, "Interval in seconds to store the current server readings to disk")
	pflag.StringP("FileStoragePath", "f", "", "Full filename where current values are saved")
	pflag.BoolP("Restore", "r", true, "Whether to load previously saved values from the specified file at server startup")
	pflag.StringP("ServerLoggerFile", "l", "serverlog.log", "Full filename where server logs are saved")
	pflag.StringP("Key", "k", "", "Key for the server")

	//d=postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable

	// Parse the command-line flags
	pflag.Parse()

	// Check for unknown flags
	for _, arg := range pflag.Args() {
		if !strings.HasPrefix(arg, "-") {
			log.Fatalf("Unknown flag: %v", arg)
		}
	}

	// Bind the flags to viper
	bindFlagToViper("DatabaseDSN")
	bindFlagToViper("ServerAddress")
	bindFlagToViper("StoreInterval")
	bindFlagToViper("FileStoragePath")
	bindFlagToViper("Restore")
	bindFlagToViper("ServerLoggerFile")
	bindFlagToViper("Key")
}

func bindFlagToViper(flagName string) {
	// Проверяем, установлена ли переменная окружения
	if viper.IsSet(flagName) {
		log.Printf("Skipping binding flag %s because environment variable is set", flagName)
		return
	}

	// Связываем флаг с viper, если переменная окружения не установлена
	if err := viper.BindPFlag(flagName, pflag.Lookup(flagName)); err != nil {
		log.Println(err)
	}
}

func bindEnvToViper(viperKey, envKey string) {
	if err := viper.BindEnv(viperKey, envKey); err != nil {
		log.Println(err)
	}
}

// NewConfig создает новый экземпляр конфигурации
func NewConfig() *Config {
	GetFlags()
	return &Config{
		ServerAddress:   Address(),
		StoreInterval:   Interval(),
		FileStoragePath: FileStoragePath(),
		Restore:         Restore(),
		ServerLogFile:   ServerLogFile(),
		DBDSN:           DBDSN(),
		SecretKey:       Key(),
	}
}

// Key возвращает ключ
func Key() string {
	return viper.GetString("Key")
}

// DBDSN возвращает строку подключения к базе данных
func DBDSN() string {
	return viper.GetString("DatabaseDSN")
}

// ServerLogFile возвращает путь к файлу логирования сервера
func ServerLogFile() string {
	return viper.GetString("ServerLoggerFile")
}

// Address возвращает адрес сервера
func Address() string {
	return viper.GetString("ServerAddress")
}

// Interval возвращает интервал сохранения текущих значений сервера на диск
func Interval() int {
	return viper.GetInt("StoreInterval")
}

// FileStoragePath возвращает путь к файлу хранения
func FileStoragePath() string {
	path := viper.GetString("FileStoragePath")
	if path == "=" {
		return ""
	}
	return path
}

// Restore возвращает флаг восстановления
func Restore() bool {
	return viper.GetBool("Restore")
}
