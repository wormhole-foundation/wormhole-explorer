-- This file contains the SQL queries to create the tables for the wormhole schema
-- create table wormhole.wh_observations
CREATE TABLE wormhole.wh_observations (
    "id" varchar not null,
    "emitter_chain_id" smallint not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "hash" varchar not null,
    "tx_hash" varchar not null,
    "guardian_address" varchar not null,
    "signature" bytea not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (id)
);
CREATE INDEX "wh_observations_hash_idx"
    ON wh_observations ("hash");
CREATE INDEX "wh_observations_tx_hash_idx"
    ON wh_observations ("tx_hash");
CREATE INDEX "wh_observations_emitter_chain_id_emitter_address_sequence_idx" 
    on wh_observations ("emitter_chain_id", "emitter_address", "sequence");

-- create table wormhole.wh_attestation_vaas
CREATE TABLE wormhole.wh_attestation_vaas (
    "id" varchar not null,
    "vaa_id" varchar not null,
    "version" smallint not null,
    "emitter_chain_id" smallint not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "guardian_set_index" bigint not null,
    "raw" bytea not null,
    "timestamp" timestamptz not null,
    "active" boolean not null,
    "is_duplicated" boolean not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
   PRIMARY KEY (id)
);

CREATE INDEX "wh_attestation_vaas_vaa_id_idx" 
    ON wormhole.wh_attestation_vaas ("vaa_id");
CREATE INDEX "wh_attestation_vaas_emitter_chain_id_idx" 
    ON wormhole.wh_attestation_vaas ("emitter_chain_id");
CREATE INDEX "wh_attestation_vaas_emitter_chain_id_emitter_address_idx" 
    ON wormhole.wh_attestation_vaas ("emitter_chain_id","emitter_address");
CREATE INDEX "wh_attestation_vaas_timestamp_idx" 
    ON wormhole.wh_attestation_vaas ("timestamp" desc);

-- create table wormhole.wh_governor_status
CREATE TABLE wormhole.wh_governor_status (
	id varchar NOT NULL,
	guardian_name varchar NOT NULL,
	message jsonb NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	CONSTRAINT wh_governor_status_pkey PRIMARY KEY (id)
);