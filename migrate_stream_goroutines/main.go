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

// (As constantes e a struct Product são as mesmas)
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
	// Garante que a tabela de destino no PG esteja limpa e pronta.
	if err := setupPostgresTarget(pgDB); err != nil {
		log.Fatalf("Erro ao configurar o destino no PostgreSQL: %v", err)
	}

	// A FUNÇÃO DE POPULAR O MONGO FOI REMOVIDA DAQUI.

	// ---- 3. INÍCIO DA MIGRAÇÃO ----
	fmt.Println("Iniciando a migração: MongoDB -> PostgreSQL")
	startTime := time.Now()

	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollection)
	productChan := make(chan Product, 100)
	var wg sync.WaitGroup

	// ---- 4. WORKERS (Inserem no PostgreSQL) ----
	for i := 0; i < numWorkers; i++ {
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

	// ---- 5. LEITURA (Stream do MongoDB) ----
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	var count int
	for cursor.Next(ctx) {
		var p Product
		if err := cursor.Decode(&p); err != nil {
			log.Printf("Erro ao decodificar documento do MongoDB: %v", err)
			continue
		}
		productChan <- p
		count++
	}

	close(productChan)
	fmt.Printf("Total de %d registros lidos do MongoDB e enviados para os workers.\n", count)

	// ---- 6. AGUARDA A FINALIZAÇÃO ----
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
