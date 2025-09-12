package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Environment string       `mapstructure:"environment"`
	LogLevel    string       `mapstructure:"log_level"`
	Server      ServerConfig `mapstructure:"server"`
	Database    DBConfig     `mapstructure:"database"`
	Redis       RedisConfig  `mapstructure:"redis"`
	RabbitMQ    RabbitConfig `mapstructure:"rabbitmq"`
	Elasticsearch ElasticConfig `mapstructure:"elasticsearch"`
	RateLimit   RateLimitConfig `mapstructure:"rate_limit"`
	JWT         JWTConfig    `mapstructure:"jwt"`
	Sources     []SourceConfig `mapstructure:"sources"`
	Collector   CollectorConfig `mapstructure:"collector"`
	Metrics     MetricsConfig `mapstructure:"metrics"`
}

type ServerConfig struct {
	Address      string `mapstructure:"address"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type DBConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxConns     int    `mapstructure:"max_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type RabbitConfig struct {
	URL          string `mapstructure:"url"`
	Exchange     string `mapstructure:"exchange"`
	QueuePrefix  string `mapstructure:"queue_prefix"`
	PrefetchCount int   `mapstructure:"prefetch_count"`
}

type ElasticConfig struct {
	Addresses []string `mapstructure:"addresses"`
	Username  string   `mapstructure:"username"`
	Password  string   `mapstructure:"password"`
	Index     string   `mapstructure:"index"`
}

type RateLimitConfig struct {
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	BurstSize         int           `mapstructure:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval"`
}

type JWTConfig struct {
	SecretKey      string        `mapstructure:"secret_key"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
	Issuer         string        `mapstructure:"issuer"`
}

type SourceConfig struct {
	Name        string            `mapstructure:"name"`
	Type        string            `mapstructure:"type"` // rss, api, scraper
	URL         string            `mapstructure:"url"`
	Schedule    string            `mapstructure:"schedule"`
	RateLimit   int              `mapstructure:"rate_limit"`
	Headers     map[string]string `mapstructure:"headers"`
	Enabled     bool             `mapstructure:"enabled"`
}

type CollectorConfig struct {
	WorkerCount     int           `mapstructure:"worker_count"`
	QueueSize       int           `mapstructure:"queue_size"`
	JobTimeout      time.Duration `mapstructure:"job_timeout"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryDelay      time.Duration `mapstructure:"retry_delay"`
	MetricsEnabled  bool          `mapstructure:"metrics_enabled"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/news-aggregator/")

	// Set defaults
	setDefaults()

	// Enable environment variable override
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("server.address", ":8080")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.idle_timeout", 120)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.database", "news_aggregator")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.max_lifetime", 300)

	// Redis defaults
	viper.SetDefault("redis.address", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	// RabbitMQ defaults
	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("rabbitmq.exchange", "news_exchange")
	viper.SetDefault("rabbitmq.queue_prefix", "news")
	viper.SetDefault("rabbitmq.prefetch_count", 10)

	// Elasticsearch defaults
	viper.SetDefault("elasticsearch.addresses", []string{"http://localhost:9200"})
	viper.SetDefault("elasticsearch.index", "news_articles")

	// Rate limiting defaults
	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.burst_size", 10)
	viper.SetDefault("rate_limit.cleanup_interval", "1m")

	// JWT defaults
	viper.SetDefault("jwt.secret_key", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiration_time", "24h")
	viper.SetDefault("jwt.issuer", "news-aggregator")

	// Collector defaults
	viper.SetDefault("collector.worker_count", 10)
	viper.SetDefault("collector.queue_size", 1000)
	viper.SetDefault("collector.job_timeout", "30s")
	viper.SetDefault("collector.retry_attempts", 3)
	viper.SetDefault("collector.retry_delay", "5s")
	viper.SetDefault("collector.metrics_enabled", true)

	// Metrics defaults
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", ":9090")
	viper.SetDefault("metrics.path", "/metrics")
}
