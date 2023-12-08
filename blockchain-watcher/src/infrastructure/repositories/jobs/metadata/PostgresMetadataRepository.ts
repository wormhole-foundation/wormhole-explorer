import pg from "pg";
import winston from "../../../../infrastructure/log";
import { Initializable, MetadataRepository } from "../../../../domain/repositories";

export class PostgresMetadataRepository implements MetadataRepository<any>, Initializable {
  private client: pg.Pool;
  private logger = winston.child({ module: "PostgresMetadataRepository" });

  constructor(client: pg.Pool) {
    this.client = client;
  }

  async init(): Promise<void> {
    this.logger.info(INIT_SQL);
    await this.client.query(`SELECT pg_advisory_lock('${PostgresMetadataRepository.lockHash()}');`);
    await this.client.query(INIT_SQL);
    await this.client.query(
      `SELECT pg_advisory_unlock('${PostgresMetadataRepository.lockHash()}');`
    );
  }

  async close(): Promise<void> {
    await this.client.end();
  }

  async get(id: string): Promise<any> {
    const pool = await this.client.connect();
    const res = await pool.query("SELECT * FROM jobmetadata WHERE job_id = $1", [id]);
    pool.release();
    return res.rows[0]?.metadata;
  }

  async save(id: string, metadata: any): Promise<void> {
    const pool = await this.client.connect();
    await pool.query(
      "INSERT INTO jobmetadata(job_id, metadata) VALUES($1, $2) ON CONFLICT (job_id) DO UPDATE SET metadata = $2, updated_at = NOW()",
      [id, metadata]
    );
    pool.release();
  }

  static lockHash(): number {
    var hash = 0,
      i = 0,
      len = "metadata-init".length;
    while (i < len) {
      hash = ((hash << 5) - hash + "metadata-init".charCodeAt(i++)) << 0;
    }
    return hash;
  }
}

const INIT_SQL = `
  CREATE TABLE IF NOT EXISTS jobmetadata(
    job_id VARCHAR(512) PRIMARY KEY,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
  );
`;
