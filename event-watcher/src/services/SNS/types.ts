import { VaaLog } from '../../databases/types';
import AwsSNS from './AwsSNS';

export type SNSOptionTypes = AwsSNS | null;
export interface SNSImplementation {
  publishMessage(message: VaaLog, fifo?: boolean): Promise<SNSPublishMessageOutput>;
  publishMessages(message: VaaLog[], fifo?: boolean): Promise<SNSPublishMessageOutput>;
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
    sequence: string | number;
    timestamp: string | number;
    vaa: string | null;
    txHash: string;
  };
}

export interface SNSPublishMessageOutput {
  status: 'success' | 'error';
  reason?: string;
  reasons?: string[];
}
