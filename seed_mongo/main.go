package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongoConnStr    = "mongodb://root:password@localhost:27017"
	mongoDatabase   = "destdb"
	mongoCollection = "products"
	// Definimos o número total de registros e o tamanho de cada lote
	totalRecords = 1000000
	batchSize    = 5000
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

	// Conexão com MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnStr))
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)
	fmt.Println("Conectado ao MongoDB!")

	collection := mongoClient.Database(mongoDatabase).Collection(mongoCollection)

	fmt.Println("Iniciando a inserção de 1 milhão de registros...")
	startTime := time.Now()

	// Chama a função para popular a collection
	if err := setupMongoSource(ctx, collection); err != nil {
		log.Fatalf("Erro ao popular a origem no MongoDB: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Banco de dados MongoDB populado com sucesso em %s!\n", duration)
}

// setupMongoSource popula a collection de origem no MongoDB usando lotes.
func setupMongoSource(ctx context.Context, collection *mongo.Collection) error {
	fmt.Println("Limpando a collection de origem no MongoDB...")

	if err := collection.Drop(ctx); err != nil {
		log.Printf("Aviso: não foi possível limpar a collection (pode não existir): %v", err)
	}

	// Cria uma slice para o lote atual.
	// A capacidade é definida como batchSize para otimizar a alocação de memória.
	docs := make([]interface{}, 0, batchSize)

	for i := 0; i < totalRecords; i++ {
		// Adiciona um novo produto ao lote atual
		docs = append(docs, Product{
			ID:          i + 1,
			Name:        fmt.Sprintf("Produto Mongo %d", i+1),
			Description: fmt.Sprintf("Descrição do produto vindo do Mongo %d.", i+1),
			Price:       float64(i+1) * 1.25,
			CreatedAt:   time.Now(),
		})

		// Se o lote atingiu o tamanho máximo OU se este é o último registro,
		// então insere o lote no banco de dados.
		if len(docs) == batchSize || i == totalRecords-1 {
			_, err := collection.InsertMany(ctx, docs)
			if err != nil {
				return fmt.Errorf("erro ao inserir o lote de dados no Mongo: %w", err)
			}

			// Imprime o progresso
			fmt.Printf("... %d / %d registros inseridos ...\n", i+1, totalRecords)

			// Limpa a slice para o próximo lote
			docs = make([]interface{}, 0, batchSize)
		}
	}

	fmt.Printf("%d registros inseridos no total.\n", totalRecords)
	return nil
}
