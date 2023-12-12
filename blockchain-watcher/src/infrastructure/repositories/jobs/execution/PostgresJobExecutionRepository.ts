import pg from "pg";
import winston from "../../../../infrastructure/log";
import { JobDefinition, JobExecution } from "../../../../domain/entities";
import { JobExecutionRepository } from "../../../../domain/repositories";

export class PostgresJobExecutionRepository implements JobExecutionRepository {
  private client: pg.Client;
  private logger = winston.child({ module: "PostgresJobExecutionRepository" });

  constructor(client: pg.Client) {
    this.client = client;
  }

  async init(): Promise<void> {
    await this.client.connect();
  }

  async close(): Promise<void> {
    await this.client.end();
  }

  async start(job: JobDefinition): Promise<JobExecution> {
    const result = await this.client.query("SELECT pg_try_advisory_lock($1)", [
      PostgresJobExecutionRepository.lockKey(job.id),
    ]);
    if (result.rows[0].pg_try_advisory_lock === true) {
      this.logger.info(`Job ${job.id} locked`);
    } else {
      throw new Error(`Job ${job.id} is already running`);
    }

    return { id: job.id, job, status: "running", startedAt: new Date() };
  }

  async stop(jobExec: JobExecution, error?: Error): Promise<JobExecution> {
    await this.client.query("SELECT pg_advisory_unlock($1)", [
      PostgresJobExecutionRepository.lockKey(jobExec.job.id),
    ]);

    const execution = jobExec;
    execution.status = "stopped";
    execution.error = error;
    execution.finishedAt = new Date();

    return execution;
  }

  static lockKey(id: string) {
    let hash = 0,
      i = 0,
      len = id.length;
    while (i < len) {
      hash = ((hash << 5) - hash + id.charCodeAt(i++)) << 0;
    }
    return hash;
  }
}
