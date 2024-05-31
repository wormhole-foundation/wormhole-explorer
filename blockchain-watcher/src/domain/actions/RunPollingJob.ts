import { StatRepository } from "../repositories";
import { performance } from "perf_hooks";
import { setTimeout } from "timers/promises";
import { Handler } from "../entities";
import winston from "winston";

const DEFAULT_INTERVAL = 1_000;

export abstract class RunPollingJob {
  private interval: number;
  private id: string;
  private statRepo?: StatRepository;
  private running: boolean = false;
  protected abstract logger: winston.Logger;
  protected abstract preHook(): Promise<void>;
  protected abstract hasNext(): Promise<boolean>;
  protected abstract report(): void;
  protected abstract get(): Promise<any[]>;
  protected abstract persist(): Promise<void>;

  constructor(id: string, statRepo?: StatRepository, interval: number = DEFAULT_INTERVAL) {
    this.interval = interval;
    this.statRepo = statRepo;
    this.running = true;
    this.id = id;
  }

  public async run(handlers: Handler[]): Promise<void> {
    this.logger.info("[run] Starting polling job");
    await this.preHook();

    while (this.running) {
      if (!(await this.hasNext())) {
        this.logger.info("[run] Finished processing");
        await this.stop();
        break;
      }

      let items: any[];

      try {
        this.report();

        const jobStartTime = performance.now();

        items = await this.get();
        await Promise.all(handlers.map((handler) => handler(items)));

        const jobEndTime = performance.now();
        const jobExecutionTime = BigInt((jobEndTime - jobStartTime).toFixed(0));

        this.statRepo?.measure("job_execution_time", jobExecutionTime, { job: this.id });
        this.statRepo?.count("job_items_total", { id: this.id }, items.length);
      } catch (e: Error | any) {
        this.logger.error("[run] Error processing items", e);
        this.statRepo?.count("job_runs_total", { id: this.id, status: "error" });
        await setTimeout(this.interval);
        continue;
      }

      await this.persist();
      this.statRepo?.count("job_runs_total", { id: this.id, status: "success" });
      await setTimeout(this.interval);
    }
  }

  public async stop(): Promise<void> {
    this.running = false;
    this.statRepo?.count("job_runs_stopped", { id: this.id });
  }
}
