
# Arquitetura 

## Topologia

## Componentes


# Instruções do Laboratório

## Rodando o projeto localmente

```bash
docker compose up --force-recreate
```


## Requests via API 

* Buscando carros numa região latitude e longitude

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
    },
    {
      "id": "a1b2c3d4-0002-4e5f-8a9b-000000000002",
      "lat": -23.551612,
      "lon": -46.634527,
      "geohash_12": "6gyf4b9cbv26",
      "geohash_9": "6gyf4b9cb",
      "geohash_7": "6gyf4b9",
      "geohash_5": "6gyf4",
      "timestamp": 1779317335311
    },
    {
      "id": "a1b2c3d4-0003-4e5f-8a9b-000000000003",
      "lat": -23.550021,
      "lon": -46.635599,
      "geohash_12": "6gyf4bbgpvbu",
      "geohash_9": "6gyf4bbgp",
      "geohash_7": "6gyf4bb",
      "geohash_5": "6gyf4",
      "timestamp": 1779317341034
    },
    {
      "id": "a1b2c3d4-0004-4e5f-8a9b-000000000004",
      "lat": -23.550784,
      "lon": -46.633617,
      "geohash_12": "6gyf4bdqtzwv",
      "geohash_9": "6gyf4bdqt",
      "geohash_7": "6gyf4bd",
      "geohash_5": "6gyf4",
      "timestamp": 1779317351556
    },
    {
      "id": "a1b2c3d4-0005-4e5f-8a9b-000000000005",
      "lat": -23.549904,
      "lon": -46.633758,
      "geohash_12": "6gyf4bf7fetw",
      "geohash_9": "6gyf4bf7f",
      "geohash_7": "6gyf4bf",
      "geohash_5": "6gyf4",
      "timestamp": 1779317355338
    },
    {
      "id": "a1b2c3d4-0006-4e5f-8a9b-000000000006",
      "lat": -23.551822,
      "lon": -46.633637,
      "geohash_12": "6gyf4bd2tqpv",
      "geohash_9": "6gyf4bd2t",
      "geohash_7": "6gyf4bd",
      "geohash_5": "6gyf4",
      "timestamp": 1779317333660
    },
    {
      "id": "a1b2c3d4-0007-4e5f-8a9b-000000000007",
      "lat": -23.551046,
      "lon": -46.632935,
      "geohash_12": "6gyf4bdvjyg6",
      "geohash_9": "6gyf4bdvj",
      "geohash_7": "6gyf4bd",
      "geohash_5": "6gyf4",
      "timestamp": 1779317352133
    },
    {
      "id": "a1b2c3d4-0008-4e5f-8a9b-000000000008",
      "lat": -23.549609,
      "lon": -46.634046,
      "geohash_12": "6gyf4bfjefwy",
      "geohash_9": "6gyf4bfje",
      "geohash_7": "6gyf4bf",
      "geohash_5": "6gyf4",
      "timestamp": 1779317354275
    },
    {
      "id": "a1b2c3d4-0009-4e5f-8a9b-000000000009",
      "lat": -23.549433,
      "lon": -46.633651,
      "geohash_12": "6gyf4bfqt5t8",
      "geohash_9": "6gyf4bfqt",
      "geohash_7": "6gyf4bf",
      "geohash_5": "6gyf4",
      "timestamp": 1779317344779
    },
    {
      "id": "a1b2c3d4-0010-4e5f-8a9b-000000000010",
      "lat": -23.549342,
      "lon": -46.633296,
      "geohash_12": "6gyf4bfxjkqp",
      "geohash_9": "6gyf4bfxj",
      "geohash_7": "6gyf4bf",
      "geohash_5": "6gyf4",
      "timestamp": 1779317348315
    }
  ]
}
```

* Buscando localizacão do carro via ID

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



### Operação no Cassandra

#### Conectando ao Cassandra

```
docker compose exec cassandra cqlsh
```

#### Queries Uteis 

```sql
DESCRIBE keyspaces;
```

```sql
USE geoip;
```

```sql
DESCRIBE TABLES;
```

* Buscando posicão dos carros 
  
```sql
SELECT * FROM car_location_by_id ; 
```

```sql
SELECT * FROM car_location_by_id WHERE id = a1b2c3d4-0003-4e5f-8a9b-000000000003;
```

* Buscando posição dos carros por região 

```sql
SELECT * FROM car_location_by_id WHERE geohash_5 = '6gyf4';
```