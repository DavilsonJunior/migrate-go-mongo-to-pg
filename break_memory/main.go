package main

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

// A struct continua a mesma
type LargeProduct struct {
	ID          int
	Name        string
	Description string
	Price       float64
	CreatedAt   time.Time
	SomeData    [128]byte
}

// Aumentamos o número para garantir o estouro em 32GB de RAM
const recordsToCreate = 130_000_000 // 130 milhões de registros

func main() {
	log.Println("Iniciando teste de quebra de memória para 32GB de RAM...")
	log.Println("Abra o Monitor de Atividade (macOS), Gerenciador de Tarefas (Windows) ou htop (Linux) para observar o consumo de RAM.")

	// Slice que vai crescer até quebrar a memória
	var memoryHog []LargeProduct

	// Goroutine para imprimir estatísticas de memória
	go func() {
		for {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf("Memória Alocada: %d MB", m.Alloc/1024/1024)
			time.Sleep(1 * time.Second) // Diminuí o tempo para vermos a alocação mais rápido
		}
	}()

	log.Printf("Tentando alocar %d registros na memória...", recordsToCreate)

	// Loop para encher a memória
	for i := 0; i < recordsToCreate; i++ {
		product := LargeProduct{
			ID:          i + 1,
			Name:        fmt.Sprintf("Produto Super Pesado %d", i+1),
			Description: "Esta é uma descrição longa para garantir que a string ocupe um espaço considerável na memória RAM do sistema.",
		}
		memoryHog = append(memoryHog, product)
	}

	log.Println("Se você está vendo esta mensagem, seu computador é um monstro! :)")
}
