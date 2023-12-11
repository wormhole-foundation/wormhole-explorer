import pg from "pg";
import winston from "../../../../infrastructure/log";
import { Initializable, MetadataRepository } from "../../../../domain/repositories";

export class PostgresMetadataRepository implements MetadataRepository<any>, Initializable {
  private pool: pg.Pool;
  private schema: string;
  private logger = winston.child({ module: "PostgresMetadataRepository" });

  constructor(pool: pg.Pool, schema: string = "public") {
    this.pool = pool;
    this.schema = schema;
  }

  async init(): Promise<void> {
    this.logger.debug(`Applying migration: ${INIT_SQL(this.schema)}`);
    try {
      const connection = await this.pool.connect();
      await connection.query(
        `BEGIN; SELECT pg_advisory_xact_lock('${PostgresMetadataRepository.lockKey()}');`
      );
      await connection.query(INIT_SQL(this.schema));
      await connection.query(`COMMIT;`);
      connection.release();
    } catch (e) {
      this.logger.error("Error initializing metadata table", e);
    }
  }

  async close(): Promise<void> {
    await this.pool.end();
  }

  async get(id: string): Promise<any> {
    const res = await this.pool.query("SELECT * FROM jobmetadata WHERE job_id = $1", [id]);
    return res.rows[0]?.metadata;
  }

  async save(id: string, metadata: any): Promise<void> {
    await this.pool.query(
      "INSERT INTO jobmetadata(job_id, metadata) VALUES($1, $2) ON CONFLICT (job_id) DO UPDATE SET metadata = $2, updated_at = NOW()",
      [id, metadata]
    );
  }

  static lockKey(): number {
    let hash = 0,
      i = 0,
      len = "metadata-init".length;
    while (i < len) {
      hash = ((hash << 5) - hash + "metadata-init".charCodeAt(i++)) << 0;
    }
    return hash;
  }
}

const INIT_SQL = (schema: string) => `
  CREATE SCHEMA IF NOT EXISTS ${schema};
  CREATE TABLE IF NOT EXISTS jobmetadata(
    job_id VARCHAR(512) PRIMARY KEY,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
`;
