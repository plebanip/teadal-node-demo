#!/bin/bash
set -e

postgres_pwd=`echo "$( cat /etc/secret-volume/postgres.password )"`
keycloak_pwd=`echo "$( cat /etc/secret-volume/keycloak.password )"`

psql -v ON_ERROR_STOP=1 --username postgres <<-EOSQL
  CREATE USER keycloak WITH PASSWORD '$keycloak_pwd'; 
  CREATE DATABASE keycloak;
  GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak; 
  ALTER DATABASE keycloak OWNER TO keycloak;
EOSQL

