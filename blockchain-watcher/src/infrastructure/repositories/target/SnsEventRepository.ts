import { LogFoundEvent } from "../../../domain/entities";
import { v4 as uuidv4 } from "uuid";
import winston from "../../log";
import crypto from "node:crypto";
import {
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
  PublishBatchCommand,
  SNSClient,
} from "@aws-sdk/client-sns";

const CHUNK_SIZE = 10;

export class SnsEventRepository {
  private client: SNSClient;
  private cfg: SnsConfig;
  private logger: winston.Logger;

  constructor(snsClient: SNSClient, cfg: SnsConfig) {
    this.client = snsClient;
    this.cfg = cfg;
    this.logger = winston.child({ module: "SnsEventRepository" });
    this.logger.info(`Created for topic ${cfg.topicArn}`);
  }

  async publish(events: LogFoundEvent<any>[]): Promise<SnsPublishResult> {
    if (!events.length) {
      this.logger.debug("[publish] No events to publish, continuing...");
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
        MessageAttributes: {
          event: {
            DataType: "String",
            StringValue: event.event,
          },
        },
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
        this.logger.error(
          `[asTarget] Error publishing events to SNS: ${result.reason ?? result.reasons}`
        );
        throw new Error(`Error publishing events to SNS: ${result.reason}`);
      }
    };
  }
}

export class SnsEvent {
  trackId: string;
  source: string;
  event: string;
  timestamp: string;
  version: string;
  data: Record<string, any>;

  constructor(
    trackId: string,
    source: string,
    event: string,
    timestamp: string,
    version: string,
    data: Record<string, any>
  ) {
    this.trackId = trackId;
    this.source = source;
    this.event = event;
    this.timestamp = timestamp;
    this.version = version;
    this.data = data;
  }

  static fromLogFoundEvent<T>(logFoundEvent: LogFoundEvent<T>): SnsEvent {
    const prefix = `chain-event-${uuidv4()}-${logFoundEvent.txHash}`;
    // SNS message attributes have a limit of 127 characters
    const trackId = prefix.length > 127 ? prefix.substring(0, 127) : prefix;

    return new SnsEvent(
      trackId,
      "blockchain-watcher",
      logFoundEvent.name,
      new Date().toISOString(),
      "1",
      {
        chainId: logFoundEvent.chainId,
        emitter: logFoundEvent.address,
        txHash: logFoundEvent.txHash,
        blockHeight: logFoundEvent.blockHeight.toString(),
        blockTime: new Date(logFoundEvent.blockTime * 1000).toISOString(),
        attributes: logFoundEvent.attributes,
      }
    );
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
