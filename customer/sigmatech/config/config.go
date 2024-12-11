package config

import (

	// if using go modules

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type JwtConfig struct {
	JWT_MAGIC_SECRET   string `env:"JWT_MAGIC_SECRET"`
	JWT_ACCESS_SECRET  string `env:"JWT_ACCESS_SECRET"`
	JWT_REFRESH_SECRET string `env:"JWT_REFRESH_SECRET"`
	JWT_ACCESS_EXP     int    `env:"JWT_ACCESS_EXP"`
	JWT_REFRESH_EXP    int    `env:"JWT_REFRESH_EXP"`
}

type DatabaseConfig struct {
	DB_HOST                    string `env:"DB_HOST"`
	DB_PORT                    string `env:"DB_PORT"`
	DB_USER                    string `env:"DB_USER"`
	DB_NAME                    string `env:"DB_NAME"`
	DB_PASSWORD                string `env:"DB_PASSWORD"`
	DB_MAX_OPEN_CONNECTION     int    `env:"DB_MAX_OPEN_CONNECTION"`
	DB_MAX_IDLE_CONNECTION     int    `env:"DB_MAX_IDLE_CONNECTION"`
	DB_CONNECTION_MAX_LIFETIME int    `env:"DB_CONNECTION_MAX_LIFETIME"`
	DB_LOG_MODE                bool   `env:"DB_LOGMODE"`
	DB_SCHEMA                  string `env:"DB_SCHEMA"`
}

type RedisConfig struct {
	REDIS_HOST                    string `env:"REDIS_HOST"`
	REDIS_PORT                    string `env:"REDIS_PORT"`
	REDIS_PASSWORD                string `env:"REDIS_PASSWORD"`
	REDIS_DB                      int    `env:"REDIS_DB"`
	REDIS_MAX_RETRIES             int    `env:"REDIS_MAX_RETRIES"`
	REDIS_DIAl_TIMEOUT            int    `env:"REDIS_DIAl_TIMEOUT"`
	REDIS_MAX_OPEN_CONNECTION     int    `env:"REDIS_MAX_OPEN_CONNECTION"`
	REDIS_MAX_IDLE_CONNECTION     int    `env:"REDIS_MAX_IDLE_CONNECTION"`
	REDIS_CONNECTION_MAX_LIFETIME int    `env:"REDIS_CONNECTION_MAX_LIFETIME"`
	REDIS_LOG_MODE                bool   `env:"REDIS_LOGMODE"`
}

type HTTPServerConfig struct {
	HTTPSERVER_URL                         string `env:"HTTPSERVER_URL"`
	HTTPSERVER_LISTEN                      string `env:"HTTPSERVER_LISTEN"`
	HTTPSERVER_PORT                        string `env:"HTTPSERVER_PORT"`
	HTTPSERVER_READ_TIMEOUT                int    `env:"HTTPSERVER_READ_TIMEOUT"`
	HTTPSERVER_WRITE_TIMEOUT               int    `env:"HTTPSERVER_WRITE_TIMEOUT"`
	HTTPSERVER_MAX_CONNECTIONS_PER_IP      int    `env:"HTTPSERVER_MAX_CONNECTIONS_PER_IP"`
	HTTPSERVER_MAX_REQUESTS_PER_CONNECTION int    `env:"HTTPSERVER_MAX_REQUESTS_PER_CONNECTION"`
	HTTPSERVER_MAX_KEEP_ALIVE_DURATION     int    `env:"HTTPSERVER_MAX_KEEP_ALIVE_DURATION"`
}

type LogConfig struct {
	LOG_FILE_PATH      string `env:"LOG_FILE_PATH"`
	LOG_FILE_NAME      string `env:"LOG_FILE_NAME"`
	LOG_FILE_MAXSIZE   int    `env:"LOG_FILE_MAXSIZE"`
	LOG_FILE_MAXBACKUP int    `env:"LOG_FILE_MAXBACKUP"`
	LOG_FILE_MAXAGE    int    `env:"LOG_FILE_MAXAGE"`

	ACCESS_LOG_FILE_PATH      string `env:"ACCESS_LOG_FILE_PATH"`
	ACCESS_LOG_FILE_NAME      string `env:"ACCESS_LOG_FILE_NAME"`
	ACCESS_LOG_FILE_MAXSIZE   int    `env:"ACCESS_LOG_FILE_MAXSIZE"`
	ACCESS_LOG_FILE_MAXBACKUP int    `env:"ACCESS_LOG_FILE_MAXBACKUP"`
	ACCESS_LOG_FILE_MAXAGE    int    `env:"ACCESS_LOG_FILE_MAXAGE"`
}

type ServiceConfig struct {
	ProjectVersion      string `env:"VERSION"`
	JwtConfig           JwtConfig
	DatabaseConfig      DatabaseConfig
	RedisConfig         RedisConfig
	HTTPServerConfig    HTTPServerConfig
	IntegrationConfig   IntegrationConfig
	LogConfig           LogConfig
	EncryptionConfig    EncryptionConfig
	AWSConfig           AWSConfig
	Environment         string `env:"ENVIRONMENT"`
	IPGeoLocationConfig IPGeoLocationConfig
}

type IntegrationConfig struct {
	Shopee      ShopeeConfig
	Omnichannel OmnichannelConfig
}

type IPGeoLocationConfig struct {
	IPGEOLOCATION_API_KEY string `env:"IPGEOLOCATION_API_KEY"`
}

type ShopeeConfig struct {
	ShopeeBaseURL    string `env:"SHOPEE_BASE_URL"`
	ShopeePartnerID  int    `env:"SHOPEE_PARTNER_ID"`
	ShopeePartnerKey string `env:"SHOPEE_PARTNER_KEY"`
	DumpHTTP         bool   `env:"SHOPEE_DUMP_HTTP" envDefault:"false"`
}

type AWSConfig struct {
	AccessKeyId     string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	Region          string `env:"AWS_REGION"`
	S3Endpoint      string `env:"AWS_S3_ENDPOINT"`
	S3BucketName    string `env:"AWS_S3_BUCKET_NAME"`
}

type EncryptionConfig struct {
	AESSecret string `env:"AES_SECRET"`
}

type OmnichannelConfig struct {
	OmnichannelBaseURL string `env:"OMNICHANNEL_BASE_URL" envDefault:"http://localhost:8000"`
}

var Config *ServiceConfig

func LoadConfig() (*ServiceConfig, error) {
	err := godotenv.Load(".env") // Load environment variables from .env file
	if err != nil {
		panic("Error loading .env file " + err.Error()) // Panic if .env file cannot be loaded
	}

	config := ServiceConfig{} // Create a variable to hold the configuration
	if err := env.Parse(&config); err != nil {
		panic("unable to load env config " + err.Error()) // Panic if environment variables cannot be parsed
	}

	return &config, nil // Return the loaded configuration
}
