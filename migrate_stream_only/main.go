package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	setupPostgresTarget(pgDB)

	fmt.Println("Iniciando a migração (Apenas Stream, sem Goroutines)...")
	startTime := time.Now()
	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	// ---- 3. LEITURA E ESCRITA (Streaming Sequencial) ----
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Erro ao buscar documentos no MongoDB: %v", err)
	}
	defer cursor.Close(ctx)

	fmt.Println("Iniciando leitura via stream e inserção sequencial...")
	count := 0
	for cursor.Next(ctx) {
		var p Product
		if err := cursor.Decode(&p); err != nil {
			log.Printf("Erro ao decodificar documento: %v", err)
			continue
		}

		// Inserção direta e sequencial dentro do mesmo loop de leitura
		_, err := pgDB.Exec(
			// A CORREÇÃO ESTÁ AQUI: a$4 foi trocado por $4
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
