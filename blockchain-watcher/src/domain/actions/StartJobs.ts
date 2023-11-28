import winston from "winston";
import { JobDefinition } from "../entities";
import { JobRepository } from "../repositories";

export class StartJobs {
  private readonly logger = winston.child({ module: "StartJobs" });
  private readonly repo: JobRepository;
  private runnables: Map<string, () => Promise<void>> = new Map();

  constructor(repo: JobRepository) {
    this.repo = repo;
  }

  public async runSingle(job: JobDefinition): Promise<JobDefinition> {
    if (this.runnables.has(job.id)) {
      throw new Error(`Job ${job.id} already exists. Ids must be unique`);
    }

    const handlers = await this.repo.getHandlers(job);
    if (handlers.length === 0) {
      this.logger.error(`No handlers for job ${job.id}`);
      throw new Error("No handlers for job");
    }

    const source = this.repo.getSource(job);

    this.runnables.set(job.id, () => source.run(handlers));
    this.runnables.get(job.id)!();

    return job;
  }

  public async run(): Promise<JobDefinition[]> {
    const jobs = await this.repo.getJobDefinitions();
    for (const job of jobs) {
      await this.runSingle(job);
    }

    return jobs;
  }
}
