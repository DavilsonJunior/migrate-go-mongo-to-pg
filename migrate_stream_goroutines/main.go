package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
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
	// Garante que a tabela de destino no PG esteja limpa e pronta.
	if err := setupPostgresTarget(pgManager.GetDB()); err != nil {
		log.Fatalf("Erro ao configurar o destino no PostgreSQL: %v", err)
	}

	// ---- 4. INÍCIO DA MIGRAÇÃO ----
	fmt.Println("Iniciando a migração: MongoDB -> PostgreSQL")
	startTime := time.Now()

	collection := mongoManager.GetCollection()
	productChan := make(chan models.Product, 100)
	var wg sync.WaitGroup
	pgDB := pgManager.GetDB()

	// ---- 5. WORKERS (Inserem no PostgreSQL) ----
	for i := 0; i < cfg.App.NumWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for product := range productChan {
				_, err := pgDB.Exec(
					`INSERT INTO products (id, name, description, price, created_at) VALUES ($1, $2, $3, $4, $5)`,
					product.ID, product.Name, product.Description, product.Price, product.CreatedAt,
				)
				if err != nil {
					log.Printf("Worker %d: Erro ao inserir produto ID %d no PG: %v", workerID, product.ID, err)
				}
			}
		}(i)
	}

	// ---- 6. LEITURA (Stream do MongoDB) ----
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	var count int
	for cursor.Next(ctx) {
		var p models.Product
		if err := cursor.Decode(&p); err != nil {
			log.Printf("Erro ao decodificar documento do MongoDB: %v", err)
			continue
		}
		productChan <- p
		count++
	}

	close(productChan)
	fmt.Printf("Total de %d registros lidos do MongoDB e enviados para os workers.\n", count)

	// ---- 7. AGUARDA A FINALIZAÇÃO ----
	wg.Wait()
	duration := time.Since(startTime)
	fmt.Printf("Migração concluída com sucesso em %s!\n", duration)
}

// setupPostgresTarget garante que a tabela de destino exista e esteja vazia.
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
