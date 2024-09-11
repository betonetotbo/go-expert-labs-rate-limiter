# go-expert-labs-rate-limiter

Aplicação HTTP com suporte a rating-limit por request.

As requisições HTTP são classificadas em 2 tipos:
* Por IP
* Por header `API_KEY`

## Configurações

Com base no arquivo `.env` na raiz do projeto é possível definir as seguintes configurações:

```bash
# porta do servidor HTTP da aplicação
PORT=3000
# endereço e porta do servidor Redis 
REDIS_HOST=localhost
REDIS_PORT=6379
# quantidade máxima de request por segundo
RPS=10
# tempo de negação dos requests (HTTP 429) ao exceder o RPS
INTERVAL=30s
# RPS por token (formato TOKEN_RPS.<TOKEN>=<RPS>)
TOKEN_RPS.abc123=100
```

## Redis

Existe uma definição de Redis para executar junto ao projeto. Para exeuctá-la basta comandar:

```bash
make redis
```

## Executando a aplicação

Primeiro faça o build dela com:

```bash
make build
```

Para executá-la:

```bash
./server
```

## Teste de carga

Para executar um teste de carga comande:

```bash
# instala o https://github.com/fortio/fortio globalmente
make instloadtest

# teste de carga por IP
make loadtestip

# teste de carga por TOKEN
make loadtesttoken
```
