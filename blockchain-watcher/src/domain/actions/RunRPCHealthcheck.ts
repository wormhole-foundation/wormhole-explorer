import { StatRepository } from "../repositories";
import winston from "winston";

export abstract class RunRPCHealthcheck {
  private statRepo?: StatRepository;
  private interval: number;

  protected abstract logger: winston.Logger;
  protected abstract set(): Promise<void>;
  protected abstract report(): void;

  constructor(statsRepo: StatRepository, interval: number) {
    this.statRepo = statsRepo;
    this.interval = interval;
  }

  public async run(): Promise<void> {
    try {
      this.startInterval();
    } catch (e: Error | any) {
      this.logger.error("[run] Error starting interval for pool providers", e);
      this.statRepo?.count("pool_runs_total", { id: "rpc-healthcheck", status: "error" });
    }
  }

  private startInterval(): void {
    setInterval(async () => {
      await this.executePoolTask();
    }, this.interval);
  }

  private async executePoolTask(): Promise<void> {
    try {
      const poolStartTime = performance.now();

      await this.set();
      this.report();

      const poolEndTime = performance.now();
      const poolExecutionTime = Number(((poolEndTime - poolStartTime) / 1000).toFixed(2));

      this.statRepo?.measure("pool_execution_time", poolExecutionTime, { job: "rpc-healthcheck" });
    } catch (e: Error | any) {
      this.logger.error("[executePoolTask] Error processing pool providers", e);
      this.statRepo?.count("pool_runs_total", { id: "rpc-healthcheck", status: "error" });
    }
  }
}
