-- This script will create the folowing DBs: prover_db_1, state_db_1 pool_db_1, event_db_1, prover_db_2 state_db_2, pool_db_2, event_db_2
-- It will also create the schema for prover_db_* DBs

CREATE DATABASE state_db_1;
CREATE DATABASE prover_db_1;
CREATE DATABASE pool_db_1;
CREATE DATABASE event_db_1;

CREATE DATABASE state_db_2;
CREATE DATABASE prover_db_2;
CREATE DATABASE pool_db_2;
CREATE DATABASE event_db_2;

\connect prover_db_1;

CREATE SCHEMA state;
CREATE TABLE state.nodes (hash BYTEA PRIMARY KEY, data BYTEA NOT NULL);
CREATE TABLE state.program (hash BYTEA PRIMARY KEY, data BYTEA NOT NULL);

CREATE USER prover_user_1 with password 'prover_pass';
ALTER DATABASE prover_db_1 OWNER TO prover_user_1;
ALTER SCHEMA state OWNER TO prover_user_1;
ALTER SCHEMA public OWNER TO prover_user_1;
ALTER TABLE state.nodes OWNER TO prover_user_1;
ALTER TABLE state.program OWNER TO prover_user_1;
ALTER USER prover_user_1 SET SEARCH_PATH=state;

\connect prover_db_2;

CREATE SCHEMA state;
CREATE TABLE state.nodes (hash BYTEA PRIMARY KEY, data BYTEA NOT NULL);
CREATE TABLE state.program (hash BYTEA PRIMARY KEY, data BYTEA NOT NULL);

CREATE USER prover_user_2 with password 'prover_pass';
ALTER DATABASE prover_db_2 OWNER TO prover_user_2;
ALTER SCHEMA state OWNER TO prover_user_2;
ALTER SCHEMA public OWNER TO prover_user_2;
ALTER TABLE state.nodes OWNER TO prover_user_2;
ALTER TABLE state.program OWNER TO prover_user_2;
ALTER USER prover_user_2 SET SEARCH_PATH=state;

\connect event_db_1;

CREATE TYPE level_t AS ENUM ('emerg', 'alert', 'crit', 'err', 'warning', 'notice', 'info', 'debug');
CREATE TABLE public.event (
   id BIGSERIAL PRIMARY KEY,
   received_at timestamp WITH TIME ZONE default CURRENT_TIMESTAMP,
   ip_address inet,
   source varchar(32) not null,
   component varchar(32),
   level level_t not null,
   event_id varchar(32) not null,
   description text,
   data bytea,
   json jsonb
);

\connect event_db_2;

CREATE TYPE level_t AS ENUM ('emerg', 'alert', 'crit', 'err', 'warning', 'notice', 'info', 'debug');
CREATE TABLE public.event (
   id BIGSERIAL PRIMARY KEY,
   received_at timestamp WITH TIME ZONE default CURRENT_TIMESTAMP,
   ip_address inet,
   source varchar(32) not null,
   component varchar(32),
   level level_t not null,
   event_id varchar(32) not null,
   description text,
   data bytea,
   json jsonb
);
