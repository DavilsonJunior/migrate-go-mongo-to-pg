package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Postgres PostgresConfig
	MongoDB  MongoConfig
	App      AppConfig
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type MongoConfig struct {
	Host       string
	Port       string
	User       string
	Password   string
	Database   string
	Collection string
}

type AppConfig struct {
	NumWorkers int
	BatchSize  int
}

// LoadConfig carrega a configuração das variáveis de ambiente
func LoadConfig() (*Config, error) {
	// Tenta carregar o arquivo .env (se existir)
	if err := godotenv.Load(); err != nil {
		// Se não encontrar .env, continua com as variáveis de ambiente do sistema
		fmt.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	config := &Config{
		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "localhost"),
			Port:     getEnv("POSTGRES_PORT", "5440"),
			User:     getEnv("POSTGRES_USER", "user"),
			Password: getEnv("POSTGRES_PASSWORD", "password"),
			Database: getEnv("POSTGRES_DATABASE", "sourcedb"),
			SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		},
		MongoDB: MongoConfig{
			Host:       getEnv("MONGO_HOST", "localhost"),
			Port:       getEnv("MONGO_PORT", "27017"),
			User:       getEnv("MONGO_USER", "root"),
			Password:   getEnv("MONGO_PASSWORD", "password"),
			Database:   getEnv("MONGO_DATABASE", "destdb"),
			Collection: getEnv("MONGO_COLLECTION", "products"),
		},
		App: AppConfig{
			NumWorkers: getEnvAsInt("NUM_WORKERS", 10),
			BatchSize:  getEnvAsInt("BATCH_SIZE", 1000),
		},
	}

	return config, nil
}

// GetPostgresConnectionString retorna a string de conexão do PostgreSQL
func (c *Config) GetPostgresConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.SSLMode,
	)
}

// GetMongoConnectionString retorna a string de conexão do MongoDB
func (c *Config) GetMongoConnectionString() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		c.MongoDB.User,
		c.MongoDB.Password,
		c.MongoDB.Host,
		c.MongoDB.Port,
	)
}

// getEnv retorna o valor da variável de ambiente ou o valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retorna o valor da variável de ambiente como inteiro ou o valor padrão
func getEnvAsInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
