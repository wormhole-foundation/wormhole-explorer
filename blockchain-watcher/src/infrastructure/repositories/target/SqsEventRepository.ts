import { DeleteMessageCommand, ReceiveMessageCommand, SQSClient } from "@aws-sdk/client-sqs";
import winston from "../../log";

export class SqsEventRepository {
  private receiveMessageCommand: ReceiveMessageCommand;
  private sqsClient: SQSClient;
  private queueUrl: string;
  private logger: winston.Logger;

  constructor(config: SQSConfig) {
    if (!config.queueUrl || !config.region) throw new Error("Queue url and region are required");
    this.logger = winston.child({ module: "SqsEventRepository" });

    this.queueUrl = config.queueUrl;

    this.receiveMessageCommand = new ReceiveMessageCommand({
      MaxNumberOfMessages: config.batchSize,
      WaitTimeSeconds: config.waitTimeSeconds,
      QueueUrl: config.queueUrl,
    });

    this.sqsClient = new SQSClient({
      region: config.region!,
    });
  }

  async handleMessage() {
    try {
      const response = await this.sqsClient.send(this.receiveMessageCommand);
      const messages = [];

      for (const message of response.Messages || []) {
        if (!message.ReceiptHandle) {
          this.logger.warn(`Message does not contain receipt handle ${message.MessageId}`);
          continue;
        }

        if (message && message.Body) {
          const msg = JSON.parse(message.Body) as SQSMessageWrapper;

          switch (msg.MessageAttributes.event.Value) {
            case "log-message-published":
              const proposalCreatedMessage = JSON.parse(msg.Message) as SQSLogMessagePublished;
              proposalCreatedMessage.receiptHandle = msg.ReceiptHandle!; // Add receipt handle to message to delete it later
              messages.push(proposalCreatedMessage);
              break;
            default:
              this.logger.warn(`Unknown event type: ${msg.MessageAttributes.event.Value}`);
              break;
          }
        }
      }
      return messages;
    } catch (error: any) {
      this.logger.error(`Error getting messages ${error.message}`);
      throw error;
    }
  }

  async deleteMessage(receiptHandle: string) {
    const command = new DeleteMessageCommand({
      ReceiptHandle: receiptHandle,
      QueueUrl: this.queueUrl,
    });

    try {
      await this.sqsClient.send(command);
    } catch (error: any) {
      this.logger.error(`Error deleting message ${error.message}`);
      throw error;
    }
  }
}

export interface SQSConfig {
  waitTimeSeconds?: number;
  batchSize?: number;
  queueUrl?: string;
  region?: string;
}

interface MessageAttributes {
  event: {
    Type: string;
    Value: string;
  };
}

interface SQSMessageWrapper {
  Message: any;
  Timestamp: string;
  UnsubscribeURL: string;
  MessageAttributes: MessageAttributes;
  ReceiptHandle?: string;
}

// SQS message structure by proposal created event
interface Data {
  chainId: number;
  emitter: string;
  txHash: string;
  blockHeight: string;
  blockTime: string;
  attributes: any;
}

export interface SQSLogMessagePublished {
  trackId: string;
  source: string;
  event: string;
  timestamp: string;
  version: string;
  data: Data;
  receiptHandle: string;
}
