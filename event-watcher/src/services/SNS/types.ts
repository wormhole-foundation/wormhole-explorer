import { WHTransaction, WHTransferRedeemed } from '../../databases/types';
import AwsSNS from './AwsSNS';

export type SNSOptionTypes = AwsSNS | null;
export interface SNSImplementation {
  createMessages(
    txs: WHTransaction[] | WHTransferRedeemed[],
    eventType: WhEventType,
    fifo?: boolean,
  ): Promise<void>;

  publishMessages(
    messages: SNSInput[],
    eventType: WhEventType,
    fifo?: boolean,
  ): Promise<SNSPublishMessageOutput>;
}

export interface AwsSNSConfig {
  region: string;
  topicArn: string;
  subject: string;
  credentials: {
    accessKeyId: string;
    secretAccessKey: string;
  };
}

export interface SNSInput {
  message: string;
  subject?: string;
  groupId?: string;
  deduplicationId?: string;
}

export interface WhTxSNSMessage {
  trackId: string;
  source: string;
  type: string;
  payload: {
    id: string;
    emitterChain: number;
    emitterAddr: string;
    sequence: number;
    timestamp: Date | string | number;
    vaa: Uint8Array | Buffer | string;
    txHash: string;
  };
}

export type RedeemedTxSNSMessage = object;
export interface SNSPublishMessageOutput {
  status: 'success' | 'error';
  reason?: string;
  reasons?: string[];
}

export type WhEventType = 'whTx' | 'redeemedTx';
