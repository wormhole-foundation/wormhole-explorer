import winston from "winston";
import { JobDefinition } from "../entities";
import { Job } from "../jobs";
import { RunPoolRpcs } from "./RunPoolRpcs";

export class StartJobs {
  private readonly logger = winston.child({ module: "StartJobs" });
  private readonly job: Job;
  private runnables: Map<string, () => Promise<void>> = new Map();

  constructor(job: Job) {
    this.job = job;
  }

  public async run(): Promise<JobDefinition[]> {
    const jobs = await this.job.getJobDefinitions();
    for (const job of jobs) {
      await this.runSingle(job);
    }
    return jobs;
  }

  public async runSingle(job: JobDefinition): Promise<JobDefinition> {
    if (this.runnables.has(job.id)) {
      throw new Error(`Job ${job.id} already exists. Ids must be unique`);
    }

    const handlers = await this.job.getHandlers(job);
    if (handlers.length === 0) {
      this.logger.error(`[runSingle] No handlers for job ${job.id}`);
      throw new Error("No handlers for job");
    }

    const runJob = this.job.getRunPollingJob(job);
    const runPoolRpcs = this.job.getRunPoolRpcs(job);

    this.runnables.set(job.id, () => runJob.run(handlers));
    this.runnables.set("pool-rpcs", () => runPoolRpcs.run());
    this.runnables.get(job.id)!();
    this.runnables.get("pool-rpcs")!();

    return job;
  }
}
