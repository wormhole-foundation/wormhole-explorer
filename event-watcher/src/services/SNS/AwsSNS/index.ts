import crypto from 'node:crypto';
import {
  SNSClient,
  PublishCommand,
  PublishCommandInput,
  PublishBatchCommand,
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
} from '@aws-sdk/client-sns';
import { AwsSNSConfig, SNSInput, SNSMessage, SNSPublishMessageOutput } from '../types';
import BaseSNS from '../BaseSNS';
import { env } from '../../../config';
import { WHTransaction } from '../../../databases/types';
import { makeSnsMessage } from '../utils';
import { ChainId, coalesceChainName } from '@certusone/wormhole-sdk';

const isDev = env.NODE_ENV !== 'production';

class AwsSNS extends BaseSNS {
  private client: SNSClient;
  private subject: string;
  private topicArn: string;

  constructor(private config: AwsSNSConfig) {
    super('AwsSNS');

    const { region, credentials, subject, topicArn } = this.config;

    this.subject = subject;
    this.topicArn = topicArn;
    const credentialsConfig = {
      region,
      ...(isDev && {
        credentials,
      }),
    };

    this.client = new SNSClient(credentialsConfig);
    this.logger.info('Client initialized');
  }

  makeSNSInput(whTx: WHTransaction): SNSInput {
    const snsMessage = makeSnsMessage(whTx, this.metadata);

    return {
      message: JSON.stringify(snsMessage),
      subject: env.AWS_SNS_SUBJECT,
      groupId: env.AWS_SNS_SUBJECT,
      deduplicationId: whTx.id,
    };
  }

  override async publishMessage(
    whTx: WHTransaction,
    fifo: boolean = false,
  ): Promise<SNSPublishMessageOutput> {
    const { message, subject, groupId, deduplicationId } = this.makeSNSInput(whTx);
    const input: PublishCommandInput = {
      TopicArn: this.topicArn!,
      Subject: subject ?? this.subject!,
      Message: message,
      ...(fifo && { MessageGroupId: groupId }),
      ...(fifo && { MessageDeduplicationId: deduplicationId }),
    };

    try {
      const command = new PublishCommand(input);
      await this.client?.send(command);

      if (input) {
        const { Message } = input;
        if (Message) {
          const snsMessage: SNSMessage = JSON.parse(Message);
          const { payload } = snsMessage;
          const { id, emitterChain, txHash } = payload;
          const chainName = coalesceChainName(emitterChain as ChainId);

          this.logger.info({
            id,
            emitterChain,
            chainName,
            txHash,
            message: 'Publish VAA log to SNS',
          });
        }
      }
    } catch (error: unknown) {
      this.logger.error(error);

      return {
        status: 'error',
      };
    }

    return {
      status: 'success',
    };
  }

  override async publishMessages(
    whTxs: WHTransaction[],
    fifo: boolean = false,
  ): Promise<SNSPublishMessageOutput> {
    const messages: SNSInput[] = whTxs.map((whTx) => this.makeSNSInput(whTx));
    const CHUNK_SIZE = 10;
    const batches: PublishBatchCommandInput[] = [];
    const inputs: PublishBatchRequestEntry[] = messages.map(
      ({ message, subject, groupId, deduplicationId }) => ({
        Id: crypto.randomUUID(),
        Subject: subject ?? this.subject!,
        Message: message,
        ...(fifo && { MessageGroupId: groupId }),
        ...(fifo && { MessageDeduplicationId: deduplicationId }),
      }),
    );

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
          this.logger.error(result.reason);
          errors.push(result.reason);
        } else {
          result.value?.Successful?.forEach((item) => {
            const { Id } = item;
            const input: PublishBatchRequestEntry | undefined = inputs?.find(
              (input) => input.Id === Id,
            );
            if (input) {
              const { Message } = input;
              if (Message) {
                const snsMessage: SNSMessage = JSON.parse(Message);
                const { payload } = snsMessage;
                const { id, emitterChain, txHash } = payload;
                const chainName = coalesceChainName(emitterChain as ChainId);

                this.logger.info({
                  id,
                  emitterChain,
                  chainName,
                  txHash,
                  message: 'Publish VAA log to SNS',
                });
              }
            }
          });
        }
      }

      if (errors.length > 0) {
        return {
          status: 'error',
          reasons: errors,
        };
      }
    } catch (error: unknown) {
      this.logger.error(error);

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
