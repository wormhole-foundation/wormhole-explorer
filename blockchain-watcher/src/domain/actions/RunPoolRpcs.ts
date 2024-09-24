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
      setInterval(async () => {
        const poolStartTime = performance.now();

        await this.set();
        this.report();

        const poolEndTime = performance.now();
        const poolExecutionTime = Number(((poolEndTime - poolStartTime) / 1000).toFixed(2));

        this.statRepo?.measure("pool_execution_time", poolExecutionTime, { job: "pool-rpcs" });
      }, 1 * 60 * 60 * 1000); // 1 hour
    } catch (e: Error | any) {}
  }
}
