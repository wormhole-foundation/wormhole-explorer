import { getLogger, WormholeLogger } from '../../utils/logger';
import { SNSImplementation, SNSInput, SNSPublishMessageOutput } from './types';

abstract class BaseSNS implements SNSImplementation {
  public logger: WormholeLogger;

  constructor(private readonly snsTypeName: string = '') {
    console.log('[SNS]', `Initializing as ${this.snsTypeName}...`);

    this.logger = getLogger(snsTypeName || 'sns');
  }

  abstract publishMessage(message: SNSInput): Promise<SNSPublishMessageOutput>;
  abstract publishMessages(messages: SNSInput[]): Promise<SNSPublishMessageOutput>;
}

export default BaseSNS;
