import winston from "winston";
import { Handler, JobDefinition, JobExecution, Runnable } from "../../entities";
import { JobExecutionRepository, JobRepository } from "../../repositories";

export class StartJobs {
  private readonly logger = winston.child({ module: "StartJobs" });
  private readonly jobRepository: JobRepository;
  private readonly jobExecutionRepository: JobExecutionRepository;
  private runnables: Map<string, () => Promise<void>> = new Map();

  constructor(repo: JobRepository, jobExecutionRepository: JobExecutionRepository) {
    this.jobRepository = repo;
    this.jobExecutionRepository = jobExecutionRepository;
  }
  public async run(): Promise<JobExecution[]> {
    const jobs = await this.jobRepository.getJobs(); // TODO: probably should limit by a config number to not fill each pod
    const running: JobExecution[] = [];
    for (const job of jobs) {
      try {
        if (job.paused) {
          if (this.runnables.has(job.id)) {
            await this.runnables.get(job.id)?.();
            this.runnables.delete(job.id);
          }

          this.logger.info(`[run] Job ${job.id} is paused, skipping`);
          continue;
        }
        const maybeJobexecution = await this.tryJobExecution(job);
        running.push(maybeJobexecution);
      } catch (error) {
        this.logger.warn(`[run] Error starting job ${job.id}: ${error}`);
      }
    }

    return running;
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
    this.runnables.set(job.id, innerFn());

    return jobExec;
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
}
