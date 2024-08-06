CREATE ROLE customers_service WITH LOGIN PASSWORD 'pa55word';

CREATE DATABASE customers_service;
ALTER DATABASE customers_service OWNER TO customers_service;
GRANT ALL ON DATABASE customers_service to customers_service;

