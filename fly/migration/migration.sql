-- This file contains the SQL queries to create the tables for the wormholescan schema
CREATE SCHEMA IF NOT EXISTS wormholescan;

-- create table wormholescan.wh_observations
CREATE TABLE wormholescan.wh_observations (
    "id" varchar not null,
    "emitter_chain_id" int not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "hash" varchar not null,
    "tx_hash" varchar not null,
    "guardian_address" varchar not null,
    "signature" bytea not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz null,
    PRIMARY KEY (id)
);
CREATE INDEX "wh_observations_hash_idx"
    ON wormholescan.wh_observations ("hash");
CREATE INDEX "wh_observations_tx_hash_idx"
    ON wormholescan.wh_observations ("tx_hash");
CREATE INDEX "wh_observations_emitter_chain_id_emitter_address_sequence_idx" 
    on wormholescan.wh_observations ("emitter_chain_id", "emitter_address", "sequence");
CREATE INDEX "wh_observations_created_at_idx"
    ON wormholescan.wh_observations ("created_at" desc);

-- create table wormholescan.wh_attestation_vaas
CREATE TABLE wormholescan.wh_attestation_vaas (
    "id" varchar not null,
    "vaa_id" varchar not null,
    "version" smallint not null,
    "emitter_chain_id" int not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "guardian_set_index" bigint not null,
    "raw" bytea not null,
    "timestamp" timestamptz not null,
    "active" boolean not null,
    "is_duplicated" boolean not null,
    "consistency_level" smallint null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz null,
   PRIMARY KEY (id)
);
CREATE INDEX "wh_attestation_vaas_vaa_id_idx" 
    ON wormholescan.wh_attestation_vaas ("vaa_id");
CREATE INDEX "wh_attestation_vaas_emitter_chain_id_idx" 
    ON wormholescan.wh_attestation_vaas ("emitter_chain_id");
CREATE INDEX "wh_attestation_vaas_emitter_chain_id_emitter_address_idx" 
    ON wormholescan.wh_attestation_vaas ("emitter_chain_id","emitter_address");
CREATE INDEX "wh_attestation_vaas_timestamp_idx" 
    ON wormholescan.wh_attestation_vaas ("timestamp" desc);

-- create table wormholescan.wh_operation_transactions
CREATE TABLE wormholescan.wh_operation_transactions (
    "chain_id" int not null,
    "tx_hash" varchar not null,
    "type" varchar not null,
    "created_at" timestamp not null,
    "updated_at" timestamp not null,
    "attestation_vaas_id" varchar not null,
    "message_id" varchar not null,
    "status" varchar null,
    "from_address" varchar null,
    "to_address" varchar null,
    "block_number" varchar null,
    "blockchain_method" varchar null,
    "fee_detail" jsonb null,
    "timestamp" timestamptz not null,
    "rpc_response" json null,
    PRIMARY KEY (message_id, tx_hash)
);
CREATE INDEX "wh_operation_transactions_message_id_idx"
    ON wormholescan.wh_operation_transactions ("message_id");
CREATE INDEX "wh_operation_transactions_tx_hash_idx"
    ON wormholescan.wh_operation_transactions ("tx_hash");
CREATE INDEX "wh_operation_transactions_from_address_idx"
    ON wormholescan.wh_operation_transactions ("from_address");
CREATE INDEX "wh_operation_transactions_to_address_idx"
    ON wormholescan.wh_operation_transactions ("to_address");
CREATE INDEX "wh_operation_transactions_chain_id_type_idx"
    ON wormholescan.wh_operation_transactions ("chain_id", "type");
CREATE INDEX "wh_operation_transactions_attestation_vaas_id_idx"
    ON wormholescan.wh_operation_transactions ("attestation_vaas_id");
CREATE INDEX "wh_operation_transactions_timestamp_idx" 
    ON wormholescan.wh_operation_transactions ("timestamp" desc);

-- create table wormholescan.wh_operation_transactions_processed
CREATE TABLE wormholescan.wh_operation_transactions_processed (
    "message_id" varchar not null,
    "tx_hash" varchar not null,
    "attestation_vaas_id" varchar not null,
    "type" varchar not null,
    "processed" bool not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY ("message_id", "tx_hash")
);
CREATE INDEX "wh_operation_transactions_processed_attestation_vaas_id_idx"
    ON wormholescan.wh_operation_transactions_processed ("attestation_vaas_id");

-- create table wormholescan.wh_governor_status
CREATE TABLE wormholescan.wh_governor_status (
	id varchar NOT NULL,
	guardian_name varchar NOT NULL,
	message jsonb NOT NULL,
    "timestamp" timestamptz not null,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	CONSTRAINT wh_governor_status_pkey PRIMARY KEY (id)
);

-- create table wormholescan.wh_governor_config
CREATE TABLE wormholescan.wh_governor_config (
    "id" varchar not null,
    "guardian_name" varchar not null,
    "counter" bigint not null,
    "timestamp" timestamptz not null,
    "tokens" jsonb not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (id)
);

-- create table wormholescan.wh_heartbeats
CREATE TABLE wormholescan.wh_heartbeats(
    "id" varchar not null,
    "guardian_name" varchar not null,
    "boot_timestamp" timestamptz not null,
    "timestamp" timestamptz not null,
    "version" varchar not null,
    "networks" jsonb not null,
    "feature" text[],
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (id)
);

-- create table wormholescan.wh_attestation_vaas_pythnet
CREATE TABLE wormholescan.wh_attestation_vaas_pythnet (
    "id" varchar not null,
    "vaa_id" varchar not null,
    "version" smallint not null,
    "emitter_chain_id" int not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "guardian_set_index" bigint not null,
    "raw" bytea not null,
    "timestamp" timestamptz not null,
    "active" boolean not null,
    "is_duplicated" boolean not null,
    "consistency_level" smallint null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz null,
   PRIMARY KEY (id)
);
CREATE INDEX "wh_attestation_vaas_pythnet_vaa_id_idx" 
    ON wormholescan.wh_attestation_vaas_pythnet ("vaa_id");
CREATE INDEX "wh_attestation_vaas_pythnet_emitter_chain_id_idx" 
    ON wormholescan.wh_attestation_vaas_pythnet ("emitter_chain_id");
CREATE INDEX "wh_attestation_vaas_pythnet_emitter_chain_id_emitter_address_idx" 
    ON wormholescan.wh_attestation_vaas_pythnet ("emitter_chain_id","emitter_address");
CREATE INDEX "wh_attestation_vaas_pythnet_timestamp_idx" 
    ON wormholescan.wh_attestation_vaas_pythnet ("timestamp" desc);

-- create table wormholescan.wh_guardian_sets
CREATE TABLE wormholescan.wh_guardian_sets (
    "id" bigint not null,
    "expiration_time" timestamptz null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (id)
);

-- create table wormholescan.wh_guardian_set_addresses
CREATE TABLE wormholescan.wh_guardian_set_addresses (
    "guardian_set_id" bigint not null,
    "index" bigint not null,
    "address" bytea not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (guardian_set_id, "index")
);

-- create table wormholescan.governor_config_chains
CREATE TABLE wormholescan.wh_governor_config_chains (
    "governor_config_id" varchar not null,
    "chain_id" int not null,
    "notional_limit" decimal(20,0) not null,
    "big_transaction_size" decimal(20,0) not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY key (governor_config_id, chain_id)
);

-- create table wormholescan.wh_guardian_governor_vaas
CREATE TABLE wormholescan.wh_guardian_governor_vaas (
    "guardian_address" varchar not null,
    "vaa_id" varchar not null,
    "guardian_name" varchar not null,    
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY  (guardian_address, vaa_id)
);

-- create table wormholescan.wh_governor_vaas
CREATE TABLE  wormholescan.wh_governor_vaas (
    "id" varchar not null,
    "chain_id" int not null,
    "emitter_address" varchar not null,
    "sequence" decimal(20,0) not null,
    "tx_hash" varchar not null,
    "release_time" timestamptz not null,
    "notional_value" decimal(20,0) not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY  (id)
);

-- create table wormholescan.wh_operation_prices
CREATE TABLE wormholescan.wh_operation_prices (
    "id" varchar not null,
    "message_id" varchar not null,
    "token_chain_id" int not null,
    "token_address" varchar not null,
    "coingecko_id" varchar not null,
    "symbol" varchar not null,
    "token_usd_price" decimal(20,8) not null,
    "total_token" decimal(30,8) not null,
    "total_usd" decimal(20,8) not null,
    "timestamp" timestamptz not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY (id)
);

-- create table wormholescan.wh_attestation_vaa_properties
CREATE TABLE wormholescan.wh_attestation_vaa_properties (
    "id" varchar not null,
    "message_id" varchar not null,
    "app_id" text[] null,
    "payload" json null,
    "raw_standard_fields" json null,
    "from_chain_id" int null,
    "from_address" varchar null,
    "to_chain_id" int null,
    "to_address" varchar null,
    "token_chain_id" int null,
    "token_address" varchar null,
    "amount" decimal(30,0) null,
    "fee_chain_id" int null,
    "fee_address" varchar null,
    "fee" decimal(30,0) null,
    "timestamp" timestamptz not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
     PRIMARY KEY (id)
);
CREATE INDEX "wh_attestation_vaa_properties_message_id_idx"
    ON wormholescan.wh_attestation_vaa_properties ("message_id");
CREATE INDEX "wh_attestation_vaa_properties_app_id_idx" 
    ON wormholescan.wh_attestation_vaa_properties USING gin("app_id"); 
CREATE INDEX "wh_attestation_vaa_properties_from_address_idx" 
    ON wormholescan.wh_attestation_vaa_properties ("from_address");
CREATE INDEX "wh_attestation_vaa_properties_to_address_idx" 
    ON wormholescan.wh_attestation_vaa_properties ("to_address");
CREATE INDEX "wh_attestation_vaa_properties_timestamp_idx" 
    ON wormholescan.wh_attestation_vaa_properties ("timestamp" desc);

-- create table wormholescan.wh_relays
CREATE TABLE wormholescan.wh_relays (
    "vaa_id" varchar not null,
    "relayer" varchar not null,
    "event" varchar not null,
    "status" varchar null,
    "received_at" timestamptz null,
    "completed_at" timestamptz null,
    "failed_at" timestamptz null,
    "from_tx_hash" varchar null,
    "to_tx_hash" varchar null,
    "message" json not null,
    "created_at" timestamptz not null,
    "updated_at" timestamptz not null,
    PRIMARY KEY ("vaa_id")
);
CREATE INDEX "wh_relays_transactions_from_tx_hash_idx"
    ON wormholescan.wh_relays ("from_tx_hash");
CREATE INDEX "wh_relays_transactions_to_tx_hash_idx"
    ON wormholescan.wh_relays ("to_tx_hash");

