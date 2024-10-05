import { RunRPCHealthcheck } from "./RunRPCHealthcheck";
import { JobDefinition } from "../entities";
import winston from "winston";
import { Job } from "../jobs";

const RPC_HEALTHCHECK = "rpc-healthcheck";

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
      await this.runJob(job);
    }
    await this.runRPCHealthcheck(jobs);
    return jobs;
  }

  public async runJob(job: JobDefinition): Promise<JobDefinition> {
    if (this.runnables.has(job.id)) {
      throw new Error(`Job ${job.id} already exists. Ids must be unique`);
    }

    const handlers = await this.job.getHandlers(job);
    if (handlers.length === 0) {
      this.logger.error(`[runSingle] No handlers for job ${job.id}`);
      throw new Error("No handlers for job");
    }

    const runJob = this.job.getPollingJob(job);

    this.runnables.set(job.id, () => runJob.run(handlers));
    this.runnables.get(job.id)!();
    return job;
  }

  public async runRPCHealthcheck(jobs: JobDefinition[]): Promise<RunRPCHealthcheck> {
    const runRPCHealthcheck = this.job.getRPCHealthcheck(jobs);
    this.runnables.set(RPC_HEALTHCHECK, () => runRPCHealthcheck.run());
    this.runnables.get(RPC_HEALTHCHECK)!();
    return runRPCHealthcheck;
  }
}
