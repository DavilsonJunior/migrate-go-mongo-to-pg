package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Constantes e Struct são as mesmas
const (
	postgresConnStr = "postgres://user:password@localhost:5440/sourcedb?sslmode=disable"
	mongoConnStr    = "mongodb://root:password@localhost:27017"
	mongoDatabase   = "destdb"
	mongoCollection = "products"
	numWorkers      = 10
)

type Product struct {
	ID          int       `bson:"product_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Price       float64   `bson:"price"`
	CreatedAt   time.Time `bson:"created_at"`
}

func main() {
	ctx := context.Background()

	// ---- 1. CONEXÕES ----
	pgDB, err := sql.Open("postgres", postgresConnStr)
	if err != nil {
		log.Fatalf("Erro ao conectar ao PostgreSQL: %v", err)
	}
	defer pgDB.Close()
	fmt.Println("Conectado ao PostgreSQL!")

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnStr))
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)
	fmt.Println("Conectado ao MongoDB!")

	// ---- 2. PREPARAÇÃO DO DESTINO ----
	setupPostgresTarget(pgDB) // Limpa a tabela de destino

	fmt.Println("Iniciando a migração (Apenas Goroutines, sem Stream)...")
	startTime := time.Now()
	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	// ---- 3. LEITURA (Tudo para a Memória - SEM STREAM) ----
	fmt.Println("Lendo TODOS os documentos do MongoDB para a memória...")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}

	var productsInMemory []Product
	if err = cursor.All(ctx, &productsInMemory); err != nil {
		log.Fatalf("Erro ao decodificar todos os documentos: %v", err)
	}
	fmt.Printf("%d documentos carregados na memória.\n", len(productsInMemory))

	// ---- 4. ESCRITA (Concorrente com Goroutines) ----
	productChan := make(chan Product, 100)
	var wg sync.WaitGroup

	// Inicia os workers que irão inserir no PG
	for i := 0; i < numWorkers; i++ {
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
	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("Migração concluída em %s!\n", duration)
}

func setupPostgresTarget(db *sql.DB) { /* ... (mesma função de antes para limpar a tabela) ... */
	db.Exec(`TRUNCATE TABLE products;`)
}
