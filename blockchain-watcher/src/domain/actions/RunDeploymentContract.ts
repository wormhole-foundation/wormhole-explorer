import { MetadataRepository, StatRepository } from "../repositories";
import { SqsEventRepository } from "../../infrastructure/repositories/target/SqsEventRepository";
import { RegistryConfig } from "../deploymentContract/DeploymentContract";
import winston from "winston";

export abstract class RunDeploymentContract {
  private statRepo?: StatRepository;

  protected abstract logger: winston.Logger;
  protected abstract execute(): Promise<void>;
  protected abstract report(): void;

  constructor(
    statsRepo: StatRepository,
    metadataRepo: MetadataRepository<any>,
    sqsRepository: SqsEventRepository,
    cfg: RegistryConfig
  ) {
    this.statRepo = statsRepo;
  }

  public async run(): Promise<void> {
    try {
      await this.execute();
      this.report();
    } catch (e: Error | any) {
      this.logger.error("[run] Error starting interval for rpcHealthCheck providers", e);
      this.statRepo?.count("rpc_healthcheck_runs_total", {
        id: "rpc-healthcheck",
        status: "error",
      });
    }
  }
}
