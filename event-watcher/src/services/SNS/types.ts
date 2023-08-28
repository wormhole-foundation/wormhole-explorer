import AwsSNS from './AwsSNS';

export type SNSOptionTypes = AwsSNS | null;
export interface SNSImplementation {
  publishMessage(message: SNSInput): Promise<SNSPublishMessageOutput>;
  publishMessages(messages: SNSInput[]): Promise<SNSPublishMessageOutput>;
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

export interface SNSPublishMessageOutput {
  status: 'success' | 'error';
  reason?: string;
  reasons?: string[];
}
