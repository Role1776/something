package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port               string        `yaml:"port"`
		Timeout            time.Duration `yaml:"timeout"`
		ReadTimeout        time.Duration `yaml:"read_timeout"`
		WriteTimeout       time.Duration `yaml:"write_timeout"`
		MaxHeaderMegabytes int           `yaml:"max_header_megabytes"`
	}

	Database struct {
		ConnString      string        `env:"DATABASE_URL"`
		Timeout         time.Duration `yaml:"timeout"`
		ConnTimeout     time.Duration `yaml:"conn_timeout"`
		ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
		MaxOpenConns    int           `yaml:"max_open_conns"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	}

	Auth struct {
		JWTAccessSecret  string `env:"JWT_ACCESS_SECRET"`
		JWTRefreshSecret string `env:"JWT_REFRESH_SECRET"`
	}

	Session struct {
		ExpiresAt time.Duration `yaml:"expires_at"`
	}

	Logging struct {
		Level string `yaml:"level"`
	}

	AI struct {
		Key   string `env:"AI_KEY"`
		Model string `yaml:"model"`
	}

	Summarizer struct {
		ChunkSize    int     `yaml:"chunk_size"`
		MaxLen       int     `yaml:"max_len"`
		CompressTemp float32 `yaml:"compress_temp"`
		FormatTemp   float32 `yaml:"format_temp"`
	}

	Instructions struct {
		Book                    string `yaml:"book_instruction"`
		Jurisprudence           string `yaml:"jurisprudence_instruction"`
		Article                 string `yaml:"article_instruction"`
		Doc                     string `yaml:"doc_instruction"`
		Def                     string `yaml:"default_instruction"`
		Formatter_book          string `yaml:"formatter_book_instruction"`
		Formatter_jurisprudence string `yaml:"formatter_jurisprudence_instruction"`
		Formatter_article       string `yaml:"formatter_article_instruction"`
		Formatter_doc           string `yaml:"formatter_doc_instruction"`
		Formatter_def           string `yaml:"formatter_article_instruction"`
	}

	Email struct {
		SMTPHost        string        `env:"SMTP_HOST"`
		SMTPPassword    string        `env:"SMTP_PASSWORD"`
		SMTPPort        int           `env:"SMTP_PORT"`
		SMTPFromAddress string        `env:"SMTP_FROM_ADDRESS"`
		ExpiresAt       time.Duration `yaml:"EXPIRES_AT"`
	}
}

func Load() (*Config, error) {
	cfgPath, envPath := fetchPath()
	if cfgPath == "" || envPath == "" {
		return nil, fmt.Errorf("config path is empty")
	}

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", cfgPath)
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", envPath)
	}

	cfg := &Config{}

	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("load .env: %w", err)
	}
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return cfg, nil
}

func fetchPath() (cfg string, env string) {
	flag.StringVar(&cfg, "config", "", "path to config.yaml file")
	flag.StringVar(&env, "env", "", "path to .env file")
	flag.Parse()
	return cfg, env
}
