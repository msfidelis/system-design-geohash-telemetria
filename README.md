# GeoIP Telemetria — PoC de Rastreamento em Tempo Real

> Implementacão da Prova de Conceito do case de [Telemetria e Geolocalização](https://github.com/msfidelis/linuxtips-curso-descomplicando-o-system-design/blob/main/cases/AVANCADO_TELEMETRIA_LOGISTICA.md) resolvido na aula de System Design da LinuxTips

> Projeto ALTAMENTE Vibecodado kkkk

PoC de um sistema de rastreamento de veículos em tempo real usando geohash para indexação geográfica. O objetivo é demonstrar uma pipeline de ingestão de dados de localização com múltiplos consumidores (hot e cold path), usando tecnologias de streaming e bancos de dados adequados para cada camada.

---

## Arquitetura

![Case Resolvido em Aula](/docs/T2-Logistica-Telemetria.drawio.png)

> Output do case realizado com os alunos 

### Topologia

```
┌─────────────┐   gRPC stream    ┌──────────────────┐   MQTT publish     ┌──────────────────────┐
│  simulacao  │ ───────────────► │  geoip-receiver  │ ─────────────────► │  geoip-mqtt-consumer │
│  (20 cars)  │  LocationPayload │  :50051 (gRPC)   │  geoip/location    │  (enrich + geohash)  │
└─────────────┘                  └──────────────────┘                     └──────────┬───────────┘
                                                                                      │
                                                                           NATS publish geoip.location
                                                                                      │
                                                             ┌────────────────────────┼─────────────────────┐
                                                             │                        │                     │
                                                   ┌─────────▼────────┐   ┌──────────▼──────────┐         │
                                                   │   hot-consumer   │   │   cold-consumer     │         │
                                                   │   (LWW via LWT)  │   │   (cold path log)   │         │
                                                   └─────────┬────────┘   └─────────────────────┘         │
                                                             │                                             │
                                                   ┌─────────▼────────┐                        ┌──────────▼──────┐
                                                   │    Cassandra     │◄───────────────────────│      api        │
                                                   │  (hot storage)   │       read             │   :8080 (REST)  │
                                                   └──────────────────┘                        └─────────────────┘
```

### Fluxo de dados

```
simulacao
  └─► gRPC Bidirecional Stream
        payload: { correlation_id, id, lat, lon, timestamp }
          └─► geoip-receiver
                └─► MQTT topic: geoip/location
                      payload: "correlationId:id:lat:lon:timestamp"
                        └─► geoip-mqtt-consumer
                              └─► Enrich com geohash (precisions 5, 7, 9, 12)
                                    └─► NATS JetStream subject: geoip.location
                                          payload: JSON { correlation_id, id, lat, lon, geohash_*, timestamp }
                                            ├─► hot-consumer  → Cassandra (LWW via LWT)
                                            └─► cold-consumer → log (placeholder cold storage)
```

---

## Componentes

### simulacao
Simula 20 veículos distribuídos em dois clusters na cidade de São Paulo:
- **10 carros** em torno da **Praça da Sé** (`-23.55028, -46.63389`)
- **10 carros** em torno do **MASP** (`-23.587416, -46.657634`)

Cada carro executa um random walk com passos de 30m dentro de um raio máximo de 5km da sua origem. A cada 500ms–2000ms (intervalo aleatório) envia a posição atual via gRPC streaming.

O simulador gera um `correlationId` (UUID v4) por mensagem usando `crypto/rand`, permitindo rastrear cada evento de ponta a ponta na pipeline.

Em caso de falha no stream, reconecta com backoff exponencial (1s → 2s → 4s → ... máximo 30s).

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `GRPC_ADDR` | `geoip-receiver:50051` | Endereço do geoip-receiver |

---

### geoip-receiver
Servidor gRPC que recebe o stream de localização dos clientes e publica cada evento no broker MQTT.

**Protocolo gRPC:** stream bidirecional — o cliente envia `LocationPayload` e recebe `Ack` por mensagem.

```proto
service LocationReceiver {
  rpc Stream(stream LocationPayload) returns (stream Ack);
}

message LocationPayload {
  string correlation_id = 1;
  string id             = 2;
  double lat            = 3;
  double lon            = 4;
  int64  timestamp      = 5;
}
```

**Keepalive configurado:**
- Ping a cada `5s`, timeout `1s`
- Conexão idle máxima: `15s`
- Vida máxima da conexão: `30s` (grace `5s`)
- Aceita pings de clientes sem streams ativos

**Publicação MQTT:** topic `geoip/location`, QoS 0, formato:
```
correlationId:id:lat:lon:timestamp
```

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `GRPC_ADDR` | `:50051` | Endereço de escuta gRPC |
| `MQTT_BROKER` | `tcp://mqtt:1883` | URL do broker MQTT |

---

### geoip-mqtt-consumer
Consome o tópico MQTT `geoip/location`, enriquece o evento com geohash em 4 níveis de precisão e publica o JSON resultante no NATS JetStream.

**Enriquecimento geohash:**
| Campo | Precisão | Exemplo | Área aproximada |
|---|---|---|---|
| `geohash_12` | 12 chars | `6gyf4bftcmty` | ~37cm² |
| `geohash_9` | 9 chars | `6gyf4bftc` | ~4.8m² |
| `geohash_7` | 7 chars | `6gyf4bf` | ~150m × 150m |
| `geohash_5` | 5 chars | `6gyf4` | ~5km × 5km |

O `geohash_5` é usado como chave de partição no Cassandra para queries por região.

**Publicação NATS:** subject `geoip.location`, stream `GEOIP`.

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `MQTT_BROKER` | `tcp://mqtt:1883` | URL do broker MQTT |
| `NATS_URL` | `nats://nats:4222` | URL do servidor NATS |

---

### hot-consumer
Consome o stream NATS JetStream como consumidor durável `hot-storage` e mantém o estado atual de cada veículo no Cassandra usando **Last Write Wins via Lightweight Transactions (LWT)**.

**Estratégia LWW:**
- **Primeiro evento do carro:** `INSERT ... IF NOT EXISTS`
- **Eventos subsequentes:** `UPDATE ... IF timestamp < ?` — garante que um evento antigo nunca sobrescreve um mais novo, mesmo com redelivery ou out-of-order

**Rastreamento de região:** quando o `geohash_5` de um carro muda, o registro antigo em `car_location_by_geohash` é deletado e um novo é inserido com a nova região.

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `NATS_URL` | `nats://nats:4222` | URL do servidor NATS |
| `CASSANDRA_HOST` | `cassandra` | Host do Cassandra |

---

### cold-consumer
Consome o stream NATS JetStream como consumidor durável `cold-storage`. Atualmente loga os eventos — funciona como placeholder para uma integração futura com cold storage (S3, Parquet, Data Lake).

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `NATS_URL` | `nats://nats:4222` | URL do servidor NATS |

---

### api
API REST que lê do Cassandra e expõe dois endpoints de consulta geográfica.

**Variáveis de ambiente:**
| Variável | Padrão | Descrição |
|---|---|---|
| `CASSANDRA_HOST` | `cassandra` | Host do Cassandra |
| `PORT` | `8080` | Porta HTTP |

---

## Infraestrutura

| Serviço | Imagem | Portas | Papel |
|---|---|---|---|
| **MQTT (Mosquitto)** | `eclipse-mosquitto:2` | `1883`, `9001` (WS) | Broker de mensagens para ingestão |
| **NATS** | `nats:2-alpine` | `4222`, `8222` (monitor) | Streaming durável entre consumers |
| **Cassandra** | `cassandra:5` | `9042` | Hot storage: posição atual dos veículos |

### Stream NATS
```
Nome:      GEOIP
Subject:   geoip.location
Retention: LimitsPolicy
Storage:   FileStorage
MaxAge:    24h
```

### Schema Cassandra

```sql
-- Posição atual de cada veículo (1 row por carro)
-- Atualização via LWT: IF timestamp < ? garante Last Write Wins
CREATE TABLE car_location_by_id (
    id         UUID,
    timestamp  BIGINT,
    lat        DOUBLE,
    lon        DOUBLE,
    geohash_12 TEXT,
    geohash_9  TEXT,
    geohash_7  TEXT,
    geohash_5  TEXT,
    PRIMARY KEY (id)
);

-- Veículos agrupados por região geohash_5
-- Carros que saem da região são deletados; a nova região recebe INSERT
CREATE TABLE car_location_by_geohash (
    geohash_5  TEXT,
    id         UUID,
    timestamp  BIGINT,
    lat        DOUBLE,
    lon        DOUBLE,
    geohash_12 TEXT,
    geohash_9  TEXT,
    geohash_7  TEXT,
    PRIMARY KEY (geohash_5, id)
);
```

---

## Decisões de Design

**gRPC bidirecional para ingestão**
O simulador mantém uma conexão persistente keep-alive com o receiver. O stream bidirecional permite que o servidor envie `Ack` por mensagem (controle de fluxo) e que o cliente detecte falhas de conexão imediatamente, acionando o backoff de reconexão.

**MQTT como camada de desacoplamento**
O `geoip-receiver` não conhece os consumidores. Publicar no MQTT desacopla a ingestão do processamento: novos consumidores podem ser adicionados sem alterar o receiver.

**NATS JetStream para fanout durável**
Dois consumidores (hot e cold) consomem o mesmo stream de forma independente e com entrega garantida. O JetStream garante redelivery em caso de falha e mantém os eventos por 24h.

**Geohash em múltiplas precisões**
O encoding é feito uma única vez no `geoip-mqtt-consumer` e os 4 níveis são persistidos. Isso evita recomputação nas queries e permite buscas por regiões de tamanhos diferentes sem alterar o schema.

**LWT no Cassandra para Last Write Wins**
Eventos podem chegar fora de ordem (redelivery, múltiplos producers). O uso de `IF timestamp < ?` garante convergência correta do estado sem coordenação distribuída extra.

**correlationId por mensagem**
UUID v4 gerado no simulador com `crypto/rand` e propagado por toda a pipeline (gRPC → MQTT → NATS → consumers). Permite rastrear um evento individual nos logs de cada serviço sem persistir no modelo de dados.

---

## Rodando o projeto

```bash
docker compose up --force-recreate
```

Na primeira execução, o `cassandra-init` aguarda o Cassandra estar saudável antes de aplicar o schema. O `geoip-receiver` compila o proto automaticamente no startup do container.

---

## API

### Buscar veículos em uma região

Converte `lat/lon` para `geohash_5` e retorna todos os veículos na mesma célula (~5km × 5km).

```bash
curl "0.0.0.0:8080/carros?lat=-23.54951&lon=-46.63425" | jq .
```

```json
{
  "geohash_5": "6gyf4",
  "total": 10,
  "carros": [
    {
      "id": "a1b2c3d4-0001-4e5f-8a9b-000000000001",
      "lat": -23.54955,
      "lon": -46.633468,
      "geohash_12": "6gyf4bftcmty",
      "geohash_9": "6gyf4bftc",
      "geohash_7": "6gyf4bf",
      "geohash_5": "6gyf4",
      "timestamp": 1779317353157
    }
  ]
}
```

### Buscar posição atual de um veículo

```bash
curl "0.0.0.0:8080/carros/a1b2c3d4-0001-4e5f-8a9b-000000000001" | jq .
```

```json
{
  "id": "a1b2c3d4-0001-4e5f-8a9b-000000000001",
  "lat": -23.54976,
  "lon": -46.63382,
  "geohash_12": "6gyf4bfk9nx9",
  "geohash_9": "6gyf4bfk9",
  "geohash_7": "6gyf4bf",
  "geohash_5": "6gyf4",
  "timestamp": 1779317491742
}
```

---

## Operação no Cassandra

### Conectar

```bash
docker compose exec cassandra cqlsh
```

### Queries úteis

```sql
USE geoip;

-- Listar todas as posições atuais
SELECT * FROM car_location_by_id;

-- Posição de um carro específico
SELECT * FROM car_location_by_id WHERE id = a1b2c3d4-0001-4e5f-8a9b-000000000001;

-- Todos os carros em uma região geohash_5
SELECT * FROM car_location_by_geohash WHERE geohash_5 = '6gyf4';

-- Buscar por geohash_5 via índice secundário
SELECT * FROM car_location_by_id WHERE geohash_5 = '6gyf4';
```

### IDs dos veículos simulados

**Cluster Praça da Sé**
```
a1b2c3d4-0001-4e5f-8a9b-000000000001  até  a1b2c3d4-0010-4e5f-8a9b-000000000010
```

**Cluster MASP**
```
b2c3d4e5-0001-4f6a-9b0c-100000000001  até  b2c3d4e5-0010-4f6a-9b0c-100000000010
```
