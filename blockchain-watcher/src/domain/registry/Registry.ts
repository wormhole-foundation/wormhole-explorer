import {
  SqsEventRepository,
  SQSLogMessagePublished,
} from "../../infrastructure/repositories/target/SqsEventRepository";
import { MetadataRepository, StatRepository } from "../repositories";
import winston from "winston";

export class Registry {
  private isRunning: boolean = false;
  private readonly sqsRepository: SqsEventRepository;
  private readonly metadataRepo: any; // TODO
  private readonly statsRepo: StatRepository;
  private readonly logger: winston.Logger;
  private readonly cfg: RegistryConfig;

  constructor(
    statsRepo: StatRepository,
    metadataRepo: MetadataRepository<any>,
    sqsRepository: SqsEventRepository,
    cfg: RegistryConfig
  ) {
    this.logger = winston.child({ module: "Registry", label: "registry" });
    this.sqsRepository = sqsRepository;
    this.metadataRepo = metadataRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
  }

  protected async execute(): Promise<void> {
    this.isRunning = true;
    this.logger.info("Starting SQS consumer");

    while (this.isRunning) {
      try {
        const messages = await this.sqsRepository.handleMessage();
        for (const message of messages) {
          await this.processMessage(message);
          await this.sqsRepository.deleteMessage(message.receiptHandle);
        }
      } catch (error: any) {
        this.logger.error(`Error processing messages: ${error.message}`);
        // Add a small delay to prevent tight looping in case of persistent errors
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
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
