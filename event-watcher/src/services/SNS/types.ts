export interface SNSImplementation {
  publishMessage(message: SNSInput): Promise<SNSPublishMessageOutput>;
  publishMessages(messages: SNSInput[]): Promise<SNSPublishMessageOutput>;
}

export interface SNSConfig {
  region: string;
  topicArn: string;
  subject: string;
  credentials: {
    accessKeyId: string;
    secretAccessKey: string;
  };
}

export interface SNSInput {
  subject?: string;
  message: string;
}

export interface SNSPublishMessageOutput {
  status: 'success' | 'error';
  reason?: string;
  reasons?: string[];
}
