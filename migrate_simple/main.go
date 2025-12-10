package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"migration-go/internal/config"
	"migration-go/internal/database"
	"migration-go/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	ctx := context.Background()

	// ---- 1. CARREGAR CONFIGURAÇÃO ----
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// ---- 2. CONEXÕES ----
	// PostgreSQL
	pgManager := database.NewPostgresManager(&cfg.Postgres)
	if err := pgManager.Connect(); err != nil {
		log.Fatalf("Erro ao conectar ao PostgreSQL: %v", err)
	}
	defer pgManager.Close()
	fmt.Println("Conectado ao PostgreSQL!")

	// MongoDB
	mongoManager := database.NewMongoManager(&cfg.MongoDB)
	if err := mongoManager.Connect(ctx); err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer mongoManager.Disconnect(ctx)
	fmt.Println("Conectado ao MongoDB!")

	// ---- 3. PREPARAÇÃO DO DESTINO ----
	if err := setupPostgresTarget(pgManager.GetDB()); err != nil {
		log.Fatalf("Erro ao configurar o destino no PostgreSQL: %v", err)
	}

	fmt.Println("Iniciando a migração SIMPLES (Tudo em Memória)...")
	startTime := time.Now()

	collection := mongoManager.GetCollection()

	// ---- 4. LEITURA (Tudo para a Memória) ----
	fmt.Println("Lendo TODOS os documentos do MongoDB para a memória...")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}

	// Aqui está a grande diferença: carregamos tudo em uma slice de uma vez.
	var productsInMemory []models.Product
	if err = cursor.All(ctx, &productsInMemory); err != nil {
		log.Fatalf("Erro ao decodificar todos os documentos para a memória: %v", err)
	}
	fmt.Printf("%d documentos carregados na memória com sucesso.\n", len(productsInMemory))

	// ---- 5. ESCRITA (Loop Sequencial) ----
	fmt.Println("Iniciando inserção sequencial no PostgreSQL...")
	pgDB := pgManager.GetDB()

	// Loop simples, um por um. Sem concorrência.
	for i, product := range productsInMemory {
		_, err := pgDB.Exec(
			`INSERT INTO products (id, name, description, price, created_at) VALUES ($1, $2, $3, $4, $5)`,
			product.ID, product.Name, product.Description, product.Price, product.CreatedAt,
		)
		if err != nil {
			// Em caso de erro, apenas logamos e continuamos
			log.Printf("Erro ao inserir produto ID %d no PG: %v", product.ID, err)
		}

		// Log de progresso para não parecer que travou
		if (i+1)%5000 == 0 {
			fmt.Printf("... %d registros inseridos ...\n", i+1)
		}
	}

	// ---- 6. FINALIZAÇÃO ----
	duration := time.Since(startTime)
	fmt.Printf("Migração SIMPLES concluída com sucesso em %s!\n", duration)
}

// Função para preparar a tabela de destino (mesma de antes)
func setupPostgresTarget(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price NUMERIC(10, 2) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE
	);
	TRUNCATE TABLE products;`
	fmt.Println("Preparando a tabela de destino 'products' no PostgreSQL...")
	_, err := db.Exec(query)
	return err
}
