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
	setupPostgresTarget(pgManager.GetDB())

	fmt.Println("Iniciando a migração (Apenas Stream, sem Goroutines)...")
	startTime := time.Now()
	collection := mongoManager.GetCollection()

	// ---- 4. LEITURA E ESCRITA (Streaming Sequencial) ----
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	fmt.Println("Iniciando leitura via stream e inserção sequencial...")
	count := 0
	pgDB := pgManager.GetDB()

	for cursor.Next(ctx) {
		var p models.Product
		if err := cursor.Decode(&p); err != nil {
			log.Printf("Erro ao decodificar documento: %v", err)
			continue
		}

		// Inserção direta e sequencial dentro do mesmo loop de leitura
		_, err := pgDB.Exec(
			`INSERT INTO products (id, name, description, price, created_at) VALUES ($1, $2, $3, $4, $5)`,
			p.ID, p.Name, p.Description, p.Price, p.CreatedAt,
		)
		if err != nil {
			log.Printf("Erro ao inserir produto ID %d: %v", p.ID, err)
		}

		count++
		if count%5000 == 0 {
			fmt.Printf("... %d registros migrados ...\n", count)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("Migração concluída em %s!\n", duration)
}

func setupPostgresTarget(db *sql.DB) { /* ... (mesma função de antes para limpar a tabela) ... */
	db.Exec(`TRUNCATE TABLE products;`)
}
