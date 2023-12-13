import winston from "winston";
import { Handler, JobDefinition, JobExecution, Runnable } from "../../entities";
import { JobExecutionRepository, JobRepository } from "../../repositories";

export class StartJobs {
  private readonly logger = winston.child({ module: "StartJobs" });
  private readonly jobRepository: JobRepository;
  private readonly jobExecutionRepository: JobExecutionRepository;
  private readonly options: { maxConcurrentJobs: number };
  private runnables: Map<string, { running: () => Promise<void>; exec: JobExecution }> = new Map();

  constructor(
    repo: JobRepository,
    jobExecutionRepository: JobExecutionRepository,
    options: { maxConcurrentJobs: number } = { maxConcurrentJobs: 10 }
  ) {
    this.jobRepository = repo;
    this.jobExecutionRepository = jobExecutionRepository;
    this.options = options;
  }
  public async run(): Promise<JobExecution[]> {
    if (!this.hasCapacity()) {
      return this.getCurrentExecutions();
    }

    const jobs = await this.jobRepository.getJobs();

    for (const job of jobs) {
      if (!this.hasCapacity()) {
        break;
      }

      try {
        if (job.paused) {
          if (this.runnables.has(job.id)) {
            await this.runnables.get(job.id)?.running();
            this.runnables.delete(job.id);
          }

          this.logger.info(`[run] Job ${job.id} is paused, skipping`);
          continue;
        }

        this.runnables.get(job.id)?.exec ?? (await this.tryJobExecution(job));
      } catch (error) {
        this.logger.warn(`[run] Error starting job ${job.id}: ${error}`);
      }
    }

    this.logger.info(`[run] Ended looking for jobs. Running:  ${this.runnables.size}`);

    return this.getCurrentExecutions();
  }

  private getCurrentExecutions(): JobExecution[] {
    return Array.from(this.runnables.values()).map((runner) => runner.exec);
  }

  private hasCapacity(): boolean {
    const available = this.runnables.size < this.options.maxConcurrentJobs;
    if (!available) {
      this.logger.info(
        `[run] Max concurrent jobs reached (${this.options.maxConcurrentJobs}), stopping`
      );
    }
    return available;
  }

  private async tryJobExecution(job: JobDefinition): Promise<JobExecution> {
    const handlers = await this.jobRepository.getHandlers(job);
    if (handlers.length === 0) {
      this.logger.error(`[run] No handlers for job ${job.id}`);
      throw new Error("No handlers for job");
    }

    const runnable = this.jobRepository.getRunnableJob(job);

    return this.trackExecution(job, handlers, runnable);
  }

  private async trackExecution(
    job: JobDefinition,
    handlers: Handler[],
    runnable: Runnable
  ): Promise<JobExecution> {
    const jobExec = await this.jobExecutionRepository.start(job);
    const innerFn = () => {
      runnable
        .run(handlers)
        .then(() => this.jobExecutionRepository.stop(jobExec))
        .catch(async (error) => {
          this.logger.error(`[trackExecution] Error running job ${jobExec.job.id}: ${error}`);
          if (!(error instanceof Error)) {
            error = new Error(error);
          }
          await this.jobExecutionRepository.stop(jobExec, error);
        });

      return runnable.stop;
    };
    this.runnables.set(job.id, { running: innerFn(), exec: jobExec });

    return jobExec;
  }
}
