import { StatRepository } from "../repositories";
import winston from "winston";

export abstract class RunPoolRpcs {
  private statRepo?: StatRepository;

  protected abstract logger: winston.Logger;
  protected abstract set(): Promise<void>;
  protected abstract report(): void;

  constructor(statsRepo: StatRepository) {
    this.statRepo = statsRepo;
  }

  public async run(): Promise<void> {
    try {
      setInterval(async () => {
        const poolStartTime = performance.now();

        await this.set();
        this.report();

        const poolEndTime = performance.now();
        const poolExecutionTime = Number(((poolEndTime - poolStartTime) / 1000).toFixed(2));

        this.statRepo?.measure("pool_execution_time", poolExecutionTime, { job: "pool-rpcs" });
      }, 10 * 60 * 1000); // 10 minutes
    } catch (e: Error | any) {
      this.logger.error("[run] Error processing pool providers", e);
      this.statRepo?.count("pool_runs_total", { id: "pool-rpcs", status: "error" });
    }
  }
}
