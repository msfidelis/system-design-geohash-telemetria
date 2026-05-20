CREATE KEYSPACE IF NOT EXISTS geoip
    WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
    AND durable_writes = true;

USE geoip;

-- Posição atual de um carro (chave-valor, uma linha por carro)
-- Update via LWT: IF timestamp < ? garante Last Write Wins
-- Query: SELECT * FROM car_location_by_id WHERE id = ?
DROP TABLE IF EXISTS car_location_by_id;
CREATE TABLE IF NOT EXISTS car_location_by_id (
    id         UUID,
    timestamp  BIGINT,
    lat        DOUBLE,
    lon        DOUBLE,
    geohash_12 TEXT,
    geohash_9  TEXT,
    geohash_7  TEXT,
    geohash_5  TEXT,
    PRIMARY KEY (id)
) WITH comment = 'Hot storage: posição atual do carro (LWW via LWT)';

-- Estado atual dos carros por região geográfica (uma linha por carro por região)
-- Carros que saem da região são deletados; nova região recebe INSERT
-- Query: SELECT * FROM car_location_by_geohash WHERE geohash_5 = ?
DROP TABLE IF EXISTS car_location_by_geohash;
CREATE TABLE IF NOT EXISTS car_location_by_geohash (
    geohash_5  TEXT,
    id         UUID,
    timestamp  BIGINT,
    lat        DOUBLE,
    lon        DOUBLE,
    geohash_12 TEXT,
    geohash_9  TEXT,
    geohash_7  TEXT,
    PRIMARY KEY (geohash_5, id)
) WITH comment = 'Hot storage: carros atualmente em cada região geohash_5';
CREATE INDEX IF NOT EXISTS ON car_location_by_id (geohash_5);
