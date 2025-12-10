package models

import "time"

// Product representa a estrutura de um produto
type Product struct {
	ID          int       `bson:"product_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Price       float64   `bson:"price"`
	CreatedAt   time.Time `bson:"created_at"`
}

// LargeProduct representa a estrutura de um produto com campos grandes (para testes de memória)
type LargeProduct struct {
	ID             int       `bson:"product_id"`
	Name           string    `bson:"name"`
	Description    string    `bson:"description"`
	Price          float64   `bson:"price"`
	CreatedAt      time.Time `bson:"created_at"`
	LargeData      []byte    `bson:"large_data"`
	AdditionalInfo string    `bson:"additional_info"`
}
