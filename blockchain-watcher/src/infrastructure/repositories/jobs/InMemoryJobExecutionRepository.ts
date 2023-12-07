import { JobDefinition, JobExecution } from "../../../domain/entities";
import { JobExecutionRepository } from "../../../domain/repositories";

export class InMemoryJobExecutionRepository implements JobExecutionRepository {
  private executions: Map<string, JobExecution> = new Map();

  async start(job: JobDefinition): Promise<JobExecution> {
    if (this.executions.has(job.id)) {
      throw new Error(`Job ${job.id} already running`);
    }

    const execution = { id: job.id, job, status: "running", startedAt: new Date() };
    this.executions.set(job.id, execution);
    return execution;
  }

  async stop(jobExec: JobExecution, error?: Error): Promise<JobExecution> {
    const execution = this.executions.get(jobExec.job.id);
    if (!execution) {
      throw new Error(`No execution for job ${jobExec.job.id}`);
    }

    execution.status = "stopped";
    execution.error = error;
    execution.finishedAt = new Date();

    return execution;
  }
}
