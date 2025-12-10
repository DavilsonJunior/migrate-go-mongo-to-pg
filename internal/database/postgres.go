package database

import (
	"context"
	"database/sql"
	"fmt"

	"migration-go/internal/config"

	_ "github.com/lib/pq"
)

// PostgresManager gerencia conexões com PostgreSQL
type PostgresManager struct {
	db     *sql.DB
	config *config.PostgresConfig
}

// NewPostgresManager cria uma nova instância do gerenciador PostgreSQL
func NewPostgresManager(cfg *config.PostgresConfig) *PostgresManager {
	return &PostgresManager{
		config: cfg,
	}
}

// Connect estabelece conexão com o PostgreSQL
func (pm *PostgresManager) Connect() error {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pm.config.User,
		pm.config.Password,
		pm.config.Host,
		pm.config.Port,
		pm.config.Database,
		pm.config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("erro ao conectar ao PostgreSQL: %w", err)
	}

	// Testa a conexão
	if err := db.Ping(); err != nil {
		return fmt.Errorf("erro ao testar conexão PostgreSQL: %w", err)
	}

	pm.db = db
	return nil
}

// GetDB retorna a instância do banco de dados
func (pm *PostgresManager) GetDB() *sql.DB {
	return pm.db
}

// Close fecha a conexão com o banco de dados
func (pm *PostgresManager) Close() error {
	if pm.db != nil {
		return pm.db.Close()
	}
	return nil
}

// QueryProducts executa a query para buscar produtos
func (pm *PostgresManager) QueryProducts(ctx context.Context) (*sql.Rows, error) {
	query := "SELECT product_id, name, description, price, created_at FROM products ORDER BY product_id"
	return pm.db.QueryContext(ctx, query)
}
