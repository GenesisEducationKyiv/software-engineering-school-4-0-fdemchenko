CREATE ROLE exchanger WITH LOGIN PASSWORD 'pa55word';

CREATE DATABASE exchanger;
ALTER DATABASE exchanger OWNER TO exchanger;
GRANT ALL ON database exchanger to exchanger;
