import { LogFoundEvent } from "../../domain/entities";
import crypto from "node:crypto";
import {
  SNSClient,
  PublishBatchCommand,
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
} from "@aws-sdk/client-sns";
import winston from "winston";

const CHUNK_SIZE = 10;

export class SnsEventRepository {
  private client: SNSClient;
  private cfg: SnsConfig;
  private logger: typeof winston;

  constructor(snsClient: SNSClient, cfg: SnsConfig) {
    this.client = snsClient;
    this.cfg = cfg;
    this.logger = winston;
  }

  async publish(events: LogFoundEvent<any>[]): Promise<SnsPublishResult> {
    const batches: PublishBatchCommandInput[] = [];
    const inputs: PublishBatchRequestEntry[] = events.map((event) => ({
      Id: crypto.randomUUID(),
      Subject: this.cfg.subject ?? "BlockchainWatcher",
      Message: JSON.stringify(event),
      MessageGroupId: this.cfg.groupId,
      MessageDeduplicationId: `${event.chainId}-${event.txHash}-${event.blockHeight}-${event.name}`,
    }));

    // PublishBatchCommand: only supports max 10 items per batch
    for (let i = 0; i <= inputs.length; i += CHUNK_SIZE) {
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
          this.logger.error(result.reason);
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
      this.logger.error(error);

      return {
        status: "error",
      };
    }

    return {
      status: "success",
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
    return new SnsEvent(
      `chain-event-${logFoundEvent.txHash}-${logFoundEvent.blockHeight}`,
      "blockchain-watcher",
      logFoundEvent.name,
      new Date().toISOString(),
      "1",
      {
        chainId: logFoundEvent.chainId,
        emitterAddress: logFoundEvent.name,
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
};

export type SnsPublishResult = {
  status: "success" | "error";
  reason?: string;
  reasons?: string[];
};

/*
 {
  "trackId": "chain-event-{txId}-{position}",
  "source": "blockchain-watcher",
  "event": "log-message-published",
  "timestamp": string (timestamp in RFC3339 format)
  "version": "1",
  "data": {
      "chainId": number,
	  "emitterAddress": string,
	  "txHash": string,
	  "blockHeight": string,
	  "blockTime": string (timestamp in RFC3339 format),
	  "attributes": {
			"sender": string,
	        "sequence": number,
            "nonce": number,
			"payload": bytes,
			"consistencyLevel": number
	   }
    }
 }
 */