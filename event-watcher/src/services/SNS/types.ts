import { WHTransaction } from '../../databases/types';
import AwsSNS from './AwsSNS';

export type SNSOptionTypes = AwsSNS | null;
export interface SNSImplementation {
  publishMessage(message: WHTransaction, fifo?: boolean): Promise<SNSPublishMessageOutput>;
  publishMessages(message: WHTransaction[], fifo?: boolean): Promise<SNSPublishMessageOutput>;
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

export interface SNSMessage {
  trackId: string;
  source: string;
  type: string;
  payload: {
    id: string;
    emitterChain: number;
    emitterAddr: string;
    sequence: number;
    timestamp: Date | string | number;
    vaa: Uint8Array | Buffer;
    txHash: string;
  };
}

export interface SNSPublishMessageOutput {
  status: 'success' | 'error';
  reason?: string;
  reasons?: string[];
}
