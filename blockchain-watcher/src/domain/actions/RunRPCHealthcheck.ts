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
      this.logger.error("[run] Error starting interval for rpcHealthCheck providers", e);
      this.statRepo?.count("rpc_healthcheck_runs_total", {
        id: "rpc-healthcheck",
        status: "error",
      });
    }
  }

  private startInterval(): void {
    setInterval(async () => {
      await this.executeRpcHealthcheckTask();
    }, this.interval);
  }

  private async executeRpcHealthcheckTask(): Promise<void> {
    try {
      const rpcHealthCheckStartTime = performance.now();

      await this.set();
      this.report();

      const rpcHealthCheckEndTime = performance.now();
      const rpcHealthCheckExecutionTime = Number(
        ((rpcHealthCheckEndTime - rpcHealthCheckStartTime) / 1000).toFixed(2)
      );

      this.statRepo?.measure("rpc_healthcheck_execution_time", rpcHealthCheckExecutionTime, {
        job: "rpc-healthcheck",
      });
    } catch (e: Error | any) {
      this.logger.error("[executeRpcHealthcheckTask] Error processing rpcHealthCheck providers", e);
      this.statRepo?.count("rpc_healthcheck_runs_total", {
        id: "rpc-healthcheck",
        status: "error",
      });
    }
  }
}
