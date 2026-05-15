package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	MongoDB       MongoDBConfig       `mapstructure:"mongodb"`
	Redis         RedisConfig         `mapstructure:"redis"`
	RateLimit     RateLimitConfig     `mapstructure:"rate_limit"`
	Kafka         KafkaConfig         `mapstructure:"kafka"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	MinIO         MinIOConfig         `mapstructure:"minio"`
	Logging       LoggingConfig       `mapstructure:"logging"`
}

type ObservabilityConfig struct {
	PrometheusPort string `mapstructure:"prometheus_port"`
	JaegerEndpoint string `mapstructure:"jaeger_endpoint"`
	ServiceName    string `mapstructure:"service_name"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
}

type JWTConfig struct {
	PublicKeyURL string `mapstructure:"public_key_url"`
	CacheTTL     int    `mapstructure:"cache_ttl"`
}

type RateLimitConfig struct {
	Requests      int `mapstructure:"requests"`
	Window        int `mapstructure:"window"`
	AdminRequests int `mapstructure:"admin_requests"`
	AdminWindow   int `mapstructure:"admin_window"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}
type MongoDBConfig struct {
	URI      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	CacheTTL int    `mapstructure:"cashet_ttl"`
}
type KafkaConfig struct {
	Brokers        string   `mapstructure:"brokers"`
	ConsumerGroup  string   `mapstructure:"consumer_group"`
	ConsumerTopics []string `mapstructure:"consumer_topics"`
	ProducerTopics []string `mapstructure:"producer_topics"`
}
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
