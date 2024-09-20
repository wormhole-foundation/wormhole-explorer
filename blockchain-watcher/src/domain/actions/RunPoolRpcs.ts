import { StatRepository } from "../repositories";
import { JobDefinition } from "../entities";
import winston from "winston";

const DEFAULT_INTERVAL = 1_000;

export abstract class RunPoolRpcs {
  private statRepo?: StatRepository;
  private interval: number;

  protected abstract set(): Promise<void>;
  protected abstract report(): void;

  constructor(repositories: Map<string | string[], any>, interval: number = DEFAULT_INTERVAL) {
    this.statRepo = repositories.get("stats-repo");
    this.interval = interval;
  }

  public async run(): Promise<void> {
    try {
      this.report();

      const poolStartTime = performance.now();

      await this.set();

      const poolEndTime = performance.now();
      const poolExecutionTime = Number(((poolEndTime - poolStartTime) / 1000).toFixed(2));

      this.statRepo?.measure("pool_execution_time", poolExecutionTime, { job: "run-pool-config" });
    } catch (e: Error | any) {}
  }
}
