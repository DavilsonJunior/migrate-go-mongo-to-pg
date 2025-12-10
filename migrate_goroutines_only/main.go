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
	setupPostgresTarget(pgManager.GetDB()) // Limpa a tabela de destino

	fmt.Println("Iniciando a migração (Apenas Goroutines, sem Stream)...")
	startTime := time.Now()
	collection := mongoManager.GetCollection()

	// ---- 4. LEITURA (Tudo para a Memória - SEM STREAM) ----
	fmt.Println("Lendo TODOS os documentos do MongoDB para a memória...")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}

	var productsInMemory []models.Product
	if err = cursor.All(ctx, &productsInMemory); err != nil { // todos documentos cursor
		log.Fatalf("Erro ao decodificar todos os documentos: %v", err)
	}
	fmt.Printf("%d documentos carregados na memória.\n", len(productsInMemory))

	// ---- 5. ESCRITA (Concorrente com Goroutines) ----
	productChan := make(chan models.Product, 100) // Pra que serve o canal?
	var wg sync.WaitGroup
	pgDB := pgManager.GetDB()

	// Inicia os workers que irão inserir no PG
	for i := 0; i < cfg.App.NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for product := range productChan {
				_, err := pgDB.Exec(
					`INSERT INTO products (id, name, description, price, created_at) VALUES ($1, $2, $3, $4, $5)`,
					product.ID, product.Name, product.Description, product.Price, product.CreatedAt,
				)
				if err != nil {
					log.Printf("Erro ao inserir produto ID %d: %v", product.ID, err)
				}
			}
		}()
	}

	// Alimenta o canal com os produtos da slice em memória
	for _, product := range productsInMemory {
		productChan <- product
	}
	close(productChan)

	// Aguarda todos os workers terminarem
	fmt.Printf("work teste")
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("Migração concluída em %s!\n", duration)
}

func setupPostgresTarget(db *sql.DB) { /* ... (mesma função de antes para limpar a tabela) ... */
	db.Exec(`TRUNCATE TABLE products;`)
}
