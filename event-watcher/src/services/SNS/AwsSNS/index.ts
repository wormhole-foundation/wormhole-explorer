import crypto from 'node:crypto';
import {
  SNSClient,
  PublishBatchCommand,
  PublishBatchCommandInput,
  PublishBatchRequestEntry,
} from '@aws-sdk/client-sns';
import {
  AwsSNSConfig,
  SNSInput,
  SNSPublishMessageOutput,
  WhEventType,
  WhTxSNSMessage,
} from '../types';
import BaseSNS from '../BaseSNS';
import { env } from '../../../config';
import { WHTransaction, WHTransferRedeemed } from '../../../databases/types';
import { makeRedeemedTxSnsMessage, makeWhTxSnsMessage } from '../utils';
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

  makeSNSInput(data: WHTransaction | WHTransferRedeemed, eventType: WhEventType): SNSInput {
    let snsMessage;
    let deduplicationId;
    if (eventType === 'whTx') {
      const whTx = data as WHTransaction;
      snsMessage = makeWhTxSnsMessage(whTx, this.metadata);
      deduplicationId = whTx.id;
    }
    if (eventType === 'redeemedTx') {
      const redeemedTx = data as WHTransferRedeemed;
      snsMessage = makeRedeemedTxSnsMessage(redeemedTx, this.metadata);
      deduplicationId = 'redeemedTx.id';
    }

    return {
      message: JSON.stringify(snsMessage),
      subject: env.AWS_SNS_SUBJECT,
      groupId: env.AWS_SNS_SUBJECT,
      deduplicationId,
    };
  }

  async createMessages(
    txs: WHTransaction[] | WHTransferRedeemed[],
    eventType: WhEventType,
    fifo: boolean = false,
  ) {
    const messages: SNSInput[] = txs.map((tx) => this.makeSNSInput(tx, eventType));

    this.publishMessages(messages, eventType, fifo);
  }

  override async publishMessages(
    messages: SNSInput[],
    eventType: WhEventType,
    fifo: boolean = false,
  ): Promise<SNSPublishMessageOutput> {
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
                let snsMessage;

                if (eventType === 'whTx') {
                  snsMessage = JSON.parse(Message) as WhTxSNSMessage;
                  const { payload } = snsMessage;
                  const { id, emitterChain, txHash } = payload;
                  const chainName = coalesceChainName(emitterChain as ChainId);

                  this.logger.info({
                    id,
                    emitterChain,
                    chainName,
                    txHash,
                    message: 'Publish Wormhole Transaction Event Log to SNS',
                  });
                }
                if (eventType === 'redeemedTx') {
                  this.logger.info({
                    message: 'Publish Wormhole Transfer Redeemed Event Log to SNS',
                  });
                }
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
