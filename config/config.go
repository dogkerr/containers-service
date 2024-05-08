package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		App
		HTTP
		// Redis
		Postgres
		LogConfig

		GRPC
		RabbitMQ

		Docker
		Dkron
		Auth
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"    env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	// Redis struct {
	// 	Address  string `env-required:"true"  env:"REDIS_ADDRESS"`
	// 	Password string `env-required:"true" yaml:"password" env:"REDIS_PASSWORD"`
	// }

	Postgres struct {
		Username string `env-required:"true" yaml:"username"  env:"USERNAME_POSTGRES"`
		PGURL    string `json:"pg_url" yaml:"pg_url" env:"PG_URL"`
		Password string `env-required:"true" yaml:"password" env:"PASSWORD_POSTGRES"`
		PGScheme string `env-required:"true" json:"pg_scheme" yaml:"pg_scheme" env:"PG_SCHEME"`
		PGDB     string `env-required:"true" json:"pg_db" yaml:"pg_db" env:"PG_DB"`
	}

	LogConfig struct {
		Level string `json:"level" yaml:"level" env:"LOG_LEVEL"`
		// Filename   string `json:"filename" yaml:"filename"`
		// MaxSize    int    `json:"maxsize" yaml:"maxsize"`
		MaxAge     int `json:"max_age" yaml:"max_age" env:"LOG_MAXAGE"`
		MaxBackups int `json:"max_backups" yaml:"max_backups" env:"LOG_MAXBACKUP"`
	}

	GRPC struct {
		URLGrpc    string `json:"urlGRPC" yaml:"urlGRPC" env:"URL_GRPC"`
		MonitorURL string `json:"monitor_client" env:"GRPC_CLIENT"`
	}

	RabbitMQ struct {
		RMQAddress string `json:"rabbitmqAddress" yaml:"rmqAddress" env:"RABBITMQ_ADDRESS"`
	}

	Docker struct {
		DockerHost string `json:"docker_host" env:"DOCKER_HOST"`
	}
	Dkron struct {
		DkronURL     string `json:"dkron_url" env:"DKRON_URL"`
		MyServiceURL string `json:"ctr_svc_url" env:"CTR_URL"`
	}
	Auth struct {
		PublicKeyAuth string `json:"pubkey_auth" env:"PUBLIC_KEY_AUTH"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	err = cleanenv.ReadConfig(path+".env", cfg) // buat di doker , ../.env kalo debug (.env kalo docker)
	// err = cleanenv.ReadConfig(path+"/local.env", cfg) // local run

	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
