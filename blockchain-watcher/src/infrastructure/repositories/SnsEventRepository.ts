import { LogFoundEvent } from "../../domain/entities";
import crypto from "node:crypto";
import {
  SNSClient,
  PublishBatchCommand,
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
} from "@aws-sdk/client-sns";
import winston from "../log";
import { SnsEvent } from "../events/SnsEvent";
import { SnsRepository } from "../../domain/repositories";

const CHUNK_SIZE = 10;

export class SnsEventRepository implements SnsRepository{
  private client: SNSClient;
  private cfg: SnsConfig;
  private logger: winston.Logger;

  constructor(snsClient: SNSClient, cfg: SnsConfig) {
    this.client = snsClient;
    this.cfg = cfg;
    this.logger = winston.child({ module: "SnsEventRepository" });
    this.logger.info(`[Sns] Created for topic ${cfg.topicArn}`);
  }

  async publish(events: LogFoundEvent<any>[]): Promise<SnsPublishResult> {
    if (!events.length) {
      this.logger.warn("[publish] No events to publish, continuing...");
      return {
        status: "success",
      };
    }

    const batches: PublishBatchCommandInput[] = [];
    const inputs: PublishBatchRequestEntry[] = events
      .map(SnsEvent.fromLogFoundEvent)
      .map((event) => ({
        Id: crypto.randomUUID(),
        Subject: this.cfg.subject ?? "blockchain-watcher",
        Message: JSON.stringify(event),
        MessageGroupId: this.cfg.groupId ?? "blockchain-watcher",
        MessageDeduplicationId: event.trackId,
      }));

    // PublishBatchCommand: only supports max 10 items per batch
    for (let i = 0; i < inputs.length; i += CHUNK_SIZE) {
      const batch: PublishBatchCommandInput = {
        TopicArn: this.cfg.topicArn,
        PublishBatchRequestEntries: inputs.slice(i, i + CHUNK_SIZE),
      };

      batches.push(batch);
    }

    try {
      const promises = [];
      const errors = [];
      for (const batch of batches) {
        const command = new PublishBatchCommand(batch);
        promises.push(this.client.send(command));
      }

      const results = await Promise.allSettled(promises);

      for (const result of results) {
        if (result.status !== "fulfilled") {
          this.logger.error(`[publish] ${result.reason}`);
          errors.push(result.reason);
        }
      }

      if (errors.length > 0) {
        return {
          status: "error",
          reasons: errors,
        };
      }
    } catch (error: unknown) {
      this.logger.error(`[publish] ${error}`);

      return {
        status: "error",
      };
    }

    this.logger.info(`[publish] Published ${events.length} events to SNS`);
    return {
      status: "success",
    };
  }

  async asTarget(): Promise<(events: LogFoundEvent<any>[]) => Promise<void>> {
    return async (events: LogFoundEvent<any>[]) => {
      const result = await this.publish(events);
      if (result.status === "error") {
        this.logger.error(`[asTarget] Error publishing events to SNS: ${result.reason ?? result.reasons}`);
        throw new Error(`Error publishing events to SNS: ${result.reason}`);
      }
    };
  }
}

export type SnsConfig = {
  region: string;
  topicArn: string;
  subject?: string;
  groupId: string;
  credentials?: {
    accessKeyId: string;
    secretAccessKey: string;
    url: string;
  };
};

export type SnsPublishResult = {
  status: "success" | "error";
  reason?: string;
  reasons?: string[];
};
