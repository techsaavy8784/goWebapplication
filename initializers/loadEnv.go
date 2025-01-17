package initializers

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	IMGStorePath string `mapstructure:"IMG_STORE_PATH"`

	DBHost           string `mapstructure:"POSTGRES_HOST"`
	DBUserName       string `mapstructure:"POSTGRES_USER"`
	DBUserPassword   string `mapstructure:"POSTGRES_PASSWORD"`
	DBName           string `mapstructure:"POSTGRES_DB"`
	DBPort           string `mapstructure:"POSTGRES_PORT"`
	ServerPort       string `mapstructure:"PORT"`
	TELEGRAM_TOKEN   string `mapsstructure:"TELEGRAM_TOKEN"`
	TELEGRAM_CHANNEL int    `mapsstructure:"TELEGRAM_CHANNEL"`
	SERVER_URL       string `mapsstructure:"SERVER_URL"`

	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"`
	RedisUri     string `mapstructure:"REDIS_URL"`
	Amqpurl      string `mapstructure:"AMQP_URL"`
	RabbitMQUri  string `mapstructure:"RABBITMQ_URL"`
	RethinkDBUri string `mapstructure:"RETHINK_URL"`

	AccessTokenPrivateKey  string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
	AccessTokenPublicKey   string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY"`
	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY"`
	AccessTokenExpiresIn   time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN"`
	RefreshTokenExpiresIn  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN"`
	AccessTokenMaxAge      int           `mapstructure:"ACCESS_TOKEN_MAXAGE"`
	RefreshTokenMaxAge     int           `mapstructure:"REFRESH_TOKEN_MAXAGE"`

	EmailFrom string `mapstructure:"EMAIL_FROM"`
	SMTPHost  string `mapstructure:"SMTP_HOST"`
	SMTPPass  string `mapstructure:"SMTP_PASS"`
	SMTPPort  int    `mapstructure:"SMTP_PORT"`
	SMTPUser  string `mapstructure:"SMTP_USER"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("app")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
