import { MetadataRepository, StatRepository } from "../repositories";
import { RunDeploymentContract } from "../actions/RunDeploymentContract";
import winston from "winston";
import {
  SQSLogMessagePublished,
  SqsEventRepository,
} from "../../infrastructure/repositories/target/SqsEventRepository";

export class DeploymentContract extends RunDeploymentContract {
  private isRunning: boolean = false;
  private readonly metadataRepo: any; // TODO
  private readonly statsRepo: StatRepository;
  protected readonly logger: winston.Logger;
  private readonly cfg: RegistryConfig;

  constructor(
    statsRepo: StatRepository,
    metadataRepo: MetadataRepository<any>,
    sqsRepository: SqsEventRepository,
    cfg: RegistryConfig
  ) {
    super(statsRepo, metadataRepo, sqsRepository, cfg);
    this.logger = winston.child({ module: "DeploymentContract", label: "deployment-contract" });
    this.metadataRepo = metadataRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
  }

  protected async execute(): Promise<void> {
    try {
      return;
    } catch (e) {
      this.logger.error(`Error setting providers: ${e}`);
    }
  }

  private async processMessage(message: SQSLogMessagePublished) {
    this.logger.info(`Processing message: ${JSON.stringify(message)}`);
    // Example: You can add your business logic here, such as saving to a database,
    // triggering other services, etc.
  }

  protected report(): void {}
}

export interface RegistryConfig {
  environment: string;
  commitment: string;
  interval?: number;
  chainId: number;
  chain: string;
  id: string;
}
