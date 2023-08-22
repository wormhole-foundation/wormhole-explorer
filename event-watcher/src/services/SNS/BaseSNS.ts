import { SNSImplementation, SNSInput, SNSPublishMessageOutput } from './types';

abstract class BaseSNS implements SNSImplementation {
  abstract publishMessage(message: SNSInput): Promise<SNSPublishMessageOutput>;
  abstract publishMessages(messages: SNSInput[]): Promise<SNSPublishMessageOutput>;
}

export default BaseSNS;
