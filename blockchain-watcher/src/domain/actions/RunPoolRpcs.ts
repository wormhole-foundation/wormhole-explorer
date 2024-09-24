import { StatRepository } from "../repositories";
import { Repos } from "../../infrastructure/repositories";

export abstract class RunPoolRpcs {
  private statRepo?: StatRepository;

  protected abstract set(): Promise<void>;
  protected abstract report(): void;

  constructor(repositories: Repos) {
    this.statRepo = repositories.statsRepo;
  }

  public async run(): Promise<void> {
    try {
      this.report();

      const poolStartTime = performance.now();

      await this.set();

      const poolEndTime = performance.now();
      const poolExecutionTime = Number(((poolEndTime - poolStartTime) / 1000).toFixed(2));

      this.statRepo?.measure("pool_execution_time", poolExecutionTime, { job: "pool-rpcs" });
    } catch (e: Error | any) {}
  }
}
