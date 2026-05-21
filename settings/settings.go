package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var Config Configuration

type Configuration struct {
	MQTTURL        string `mapstructure:"MQTT_URL"`
	DBURI          string `mapstructure:"DB_URI"`
	DB_NAME        string `mapstructure:"DB_NAME"`
	DB_TIME        int    `mapstructure:"DB_TIME"`
	AppPort        string `mapstructure:"APP_PORT"`
	AllowedDomains string `mapstructure:"ALLOWED_DOMAINS"`
}

func InitConfig() (Configuration, error) {
	var configDir, envDir string

	currentDir, err := os.Getwd()
	if err != nil {
		return Config, err
	}

	configDir = filepath.Join(currentDir, "")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		configDir = "./"
	}

	envDir = filepath.Join(currentDir, "")
	if _, err := os.Stat(envDir); os.IsNotExist(err) {
		envDir = "./"
	}

	// Load `.env`
	viper.AddConfigPath(envDir)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	if err := viper.MergeInConfig(); err != nil {
		return Config, err
	}

	// Bind sensitive env variables
	envVars := []string{
		"DB_URI", "DB_TIME", "APP_PORT", "DB_NAME", "JWT_SECRET", "ALLOWED_DOMAINS", "MQTT_URL",
	}

	for _, envVar := range envVars {
		if err := viper.BindEnv(envVar); err != nil {
			return Config, err
		}
	}

	// Unmarshal the combined configuration
	if err := viper.Unmarshal(&Config); err != nil {
		return Config, err
	}

	fmt.Println("Configuration loaded successfully")
	return Config, nil
}
