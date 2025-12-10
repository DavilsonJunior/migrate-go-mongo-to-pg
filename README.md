# Go Migration PostgreSQL ↔ MongoDB

Este projeto demonstra diferentes estratégias de migração entre bancos PostgreSQL e MongoDB usando Go, agora com estrutura organizada e configuração via variáveis de ambiente.

## 🚀 Estrutura do Projeto

```
├── internal/
│   ├── config/          # Gerenciamento de configurações
│   ├── database/        # Gerenciadores de conexão
│   └── models/          # Modelos de dados compartilhados
├── migrate_simple/      # Migração simples (tudo em memória)
├── migrate_goroutines_only/  # Usando goroutines para concorrência
├── migrate_stream_only/     # Usando streaming sem concorrência
├── migrate_stream_goroutines/ # Streaming + goroutines (otimizada)
├── break_memory/        # Teste de limite de memória
├── seed_mongo/          # Popular MongoDB com dados de teste
├── .env.example         # Exemplo de variáveis de ambiente
└── docker-compose.yml   # Containers PostgreSQL e MongoDB
```

## 🔧 Configuração

### 1. Variáveis de Ambiente

Copie o arquivo de exemplo e configure suas variáveis:

```bash
cp .env.example .env
```

Edite o arquivo `.env` com suas configurações:

```env
# PostgreSQL Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5440
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DATABASE=sourcedb
POSTGRES_SSLMODE=disable

# MongoDB Configuration
MONGO_HOST=localhost
MONGO_PORT=27017
MONGO_USER=root
MONGO_PASSWORD=password
MONGO_DATABASE=destdb
MONGO_COLLECTION=products

# Application Configuration
NUM_WORKERS=10
BATCH_SIZE=1000
```

### 2. Instalação de Dependências

```bash
go mod tidy
```

### 3. Subir os Bancos de Dados

```bash
docker-compose up -d
```

## 📋 Executáveis Disponíveis

### 1. Seed MongoDB (Preparação)
Popula o MongoDB com 1 milhão de registros de teste:

```bash
go run ./seed_mongo
# ou
go build -o seed_mongo_bin ./seed_mongo && ./seed_mongo_bin
```

### 2. Migração Simples
Carrega todos os dados na memória e depois insere no PostgreSQL:

```bash
go run ./migrate_simple
# ou
go build -o migrate_simple_bin ./migrate_simple && ./migrate_simple_bin
```

**Características:**
- ✅ Simples de implementar
- ❌ Alto consumo de memória
- ❌ Não escalável para grandes volumes

### 3. Migração com Goroutines
Usa goroutines para acelerar a inserção no PostgreSQL:

```bash
go run ./migrate_goroutines_only
# ou
go build -o migrate_goroutines_bin ./migrate_goroutines_only && ./migrate_goroutines_bin
```

**Características:**
- ✅ Inserção paralela (mais rápida)
- ❌ Ainda carrega tudo na memória
- ⚡ Configura workers via `NUM_WORKERS`

### 4. Migração com Stream
Processa os dados em streaming (um por vez):

```bash
go run ./migrate_stream_only
# ou
go build -o migrate_stream_bin ./migrate_stream_only && ./migrate_stream_bin
```

**Características:**
- ✅ Baixo consumo de memória
- ❌ Inserção sequencial (mais lenta)
- ✅ Escalável para qualquer volume

### 5. Migração Otimizada (Stream + Goroutines)
**Recomendada**: Combina streaming com processamento paralelo:

```bash
go run ./migrate_stream_goroutines
# ou
go build -o migrate_stream_goroutines_bin ./migrate_stream_goroutines && ./migrate_stream_goroutines_bin
```

**Características:**
- ✅ Baixo consumo de memória
- ✅ Inserção paralela (rápida)
- ✅ Escalável e performática
- ⚡ Configura workers via `NUM_WORKERS`

### 6. Teste de Limite de Memória
Demonstra problemas de memória com grandes volumes:

```bash
go run ./break_memory
# ou
go build -o break_memory_bin ./break_memory && ./break_memory_bin
```

## 📊 Comparação de Performance

| Estratégia | Memória | Velocidade | Escalabilidade | Complexidade |
|------------|---------|------------|----------------|--------------|
| **Simple** | Alta | Média | Baixa | Baixa |
| **Goroutines Only** | Alta | Alta | Baixa | Média |
| **Stream Only** | Baixa | Baixa | Alta | Média |
| **Stream + Goroutines** | Baixa | Alta | Alta | Alta |

## 🎯 Funcionalidades Implementadas

### ✅ Melhorias Realizadas

1. **Configuração Centralizada**
   - Suporte a variáveis de ambiente via `.env`
   - Configuração padrão com fallbacks
   - Estrutura tipada para todas as configs

2. **Separação de Responsabilidades**
   - `internal/config`: Gerenciamento de configurações
   - `internal/database`: Gerenciadores de conexão PostgreSQL e MongoDB
   - `internal/models`: Modelos de dados compartilhados

3. **Eliminação de Duplicação**
   - Código de conexão centralizado
   - Modelos compartilhados entre todos os executáveis
   - Reutilização de funções comuns

4. **Gerenciamento de Dependências**
   - Adicionado `github.com/joho/godotenv` para variáveis de ambiente
   - Organização limpa das importações

### 📦 Pacotes Internos

#### Config (`internal/config`)
- Carregamento de variáveis de ambiente
- Configurações tipadas para PostgreSQL, MongoDB e aplicação
- Geração automática de strings de conexão

#### Database (`internal/database`)
- **PostgresManager**: Gerencia conexões PostgreSQL
- **MongoManager**: Gerencia conexões MongoDB
- Métodos utilitários para operações comuns

#### Models (`internal/models`)
- **Product**: Modelo padrão de produto
- **LargeProduct**: Modelo para testes de memória
- Estruturas BSON configuradas

## 🔍 Monitoramento

Durante a execução, monitore:

- **Memória**: `htop` ou `top` (Linux), Activity Monitor (macOS), Task Manager (Windows)
- **Logs**: Progresso detalhado no terminal
- **Bancos**: Conecte nas instâncias para verificar os dados

## 🚀 Próximos Passos

Para uso em produção, considere:

1. **Pool de Conexões**: Implementar connection pooling
2. **Retry Logic**: Adicionar retry automático em falhas
3. **Métricas**: Integrar com Prometheus/Grafana
4. **Logs Estruturados**: Usar logrus ou zap
5. **Testes**: Adicionar testes unitários e de integração
6. **CI/CD**: Pipeline de build e deploy
7. **Observabilidade**: Tracing distribuído com Jaeger/OpenTelemetry

## ⚡ Dicas de Performance

1. **Ajuste `NUM_WORKERS`** conforme sua máquina
2. **Configure `BATCH_SIZE`** para otimizar inserções em lote
3. **Use conexões persistentes** ao invés de criar/fechar a cada operação
4. **Monitore métricas** de ambos os bancos durante a migração