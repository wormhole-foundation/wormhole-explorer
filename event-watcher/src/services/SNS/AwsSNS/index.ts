import crypto from 'node:crypto';
import {
  SNSClient,
  PublishCommand,
  PublishCommandInput,
  PublishBatchCommand,
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
} from '@aws-sdk/client-sns';
import { AwsSNSConfig, SNSInput, SNSPublishMessageOutput } from '../types';
import BaseSNS from '../BaseSNS';

class AwsSNS extends BaseSNS {
  private client: SNSClient;
  private subject: string;
  private topicArn: string;

  constructor(private config: AwsSNSConfig) {
    super();

    const { region, credentials, subject, topicArn } = this.config;

    this.subject = subject;
    this.topicArn = topicArn;
    this.client = new SNSClient({
      region,
      credentials,
    });
  }

  override async publishMessage({ subject, message }: SNSInput): Promise<SNSPublishMessageOutput> {
    const input: PublishCommandInput = {
      TopicArn: this.topicArn!,
      Subject: subject ?? this.subject!,
      Message: message,
    };

    try {
      const command = new PublishCommand(input);
      await this.client?.send(command);
    } catch (error) {
      console.error(error);

      return {
        status: 'error',
      };
    }

    return {
      status: 'success',
    };
  }

  override async publishMessages(messages: SNSInput[]): Promise<SNSPublishMessageOutput> {
    const CHUNK_SIZE = 10;
    const batches: PublishBatchCommandInput[] = [];
    const inputs: PublishBatchRequestEntry[] = messages.map(({ subject, message }) => ({
      Id: crypto.randomUUID(),
      Subject: subject ?? this.subject!,
      Message: message,
    }));

    // PublishBatchCommand: only supports max 10 items per batch
    for (let i = 0; i <= inputs.length; i += CHUNK_SIZE) {
      const batch: PublishBatchCommandInput = {
        TopicArn: this.topicArn!,
        PublishBatchRequestEntries: inputs.slice(i, i + CHUNK_SIZE),
      };

      batches.push(batch);
    }

    try {
      const promises = [];
      const errors = [];
      for (const batch of batches) {
        const command = new PublishBatchCommand(batch);
        promises.push(this.client?.send(command));
      }

      const results = await Promise.allSettled(promises);

      for (const result of results) {
        if (result.status !== 'fulfilled') {
          console.error(result.reason);
          errors.push(result.reason);
        }
      }

      if (errors.length > 0) {
        console.error(errors);

        return {
          status: 'error',
          reasons: errors,
        };
      }
    } catch (error) {
      console.error(error);

      return {
        status: 'error',
      };
    }

    return {
      status: 'success',
    };
  }
}

export default AwsSNS;
