package database

import (
	"context"
	"fmt"

	"migration-go/internal/config"
	"migration-go/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoManager gerencia conexões com MongoDB
type MongoManager struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
	config     *config.MongoConfig
}

// NewMongoManager cria uma nova instância do gerenciador MongoDB
func NewMongoManager(cfg *config.MongoConfig) *MongoManager {
	return &MongoManager{
		config: cfg,
	}
}

// Connect estabelece conexão com o MongoDB
func (mm *MongoManager) Connect(ctx context.Context) error {
	connStr := fmt.Sprintf("mongodb://%s:%s@%s:%s",
		mm.config.User,
		mm.config.Password,
		mm.config.Host,
		mm.config.Port,
	)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
	if err != nil {
		return fmt.Errorf("erro ao conectar ao MongoDB: %w", err)
	}

	// Testa a conexão
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("erro ao testar conexão MongoDB: %w", err)
	}

	mm.client = client
	mm.database = client.Database(mm.config.Database)
	mm.collection = mm.database.Collection(mm.config.Collection)

	return nil
}

// GetClient retorna o cliente MongoDB
func (mm *MongoManager) GetClient() *mongo.Client {
	return mm.client
}

// GetDatabase retorna a instância do banco de dados
func (mm *MongoManager) GetDatabase() *mongo.Database {
	return mm.database
}

// GetCollection retorna a instância da coleção
func (mm *MongoManager) GetCollection() *mongo.Collection {
	return mm.collection
}

// Disconnect desconecta do MongoDB
func (mm *MongoManager) Disconnect(ctx context.Context) error {
	if mm.client != nil {
		return mm.client.Disconnect(ctx)
	}
	return nil
}

// InsertOne insere um documento na coleção
func (mm *MongoManager) InsertOne(ctx context.Context, product models.Product) error {
	_, err := mm.collection.InsertOne(ctx, product)
	if err != nil {
		return fmt.Errorf("erro ao inserir produto no MongoDB: %w", err)
	}
	return nil
}

// InsertMany insere múltiplos documentos na coleção
func (mm *MongoManager) InsertMany(ctx context.Context, products []interface{}) error {
	if len(products) == 0 {
		return nil
	}

	_, err := mm.collection.InsertMany(ctx, products)
	if err != nil {
		return fmt.Errorf("erro ao inserir produtos no MongoDB: %w", err)
	}
	return nil
}

// DropCollection remove todos os documentos da coleção
func (mm *MongoManager) DropCollection(ctx context.Context) error {
	return mm.collection.Drop(ctx)
}
