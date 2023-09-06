import { getLogger, WormholeLogger } from '../../utils/logger';
import { SNSImplementation, SNSInput, SNSPublishMessageOutput } from './types';
import { VaaLog } from '../../databases/types';

abstract class BaseSNS implements SNSImplementation {
  public logger: WormholeLogger;
  public metadata = {
    source: 'event-watcher',
    type: 'published-log-message',
  } as const;

  constructor(private readonly snsTypeName: string = '') {
    this.logger = getLogger(snsTypeName || 'sns');
    this.logger.info(`Initializing as ${this.snsTypeName}...`);
  }

  abstract makeSNSInput(vaaLog: VaaLog): SNSInput;
  abstract publishMessage(message: VaaLog, fifo?: boolean): Promise<SNSPublishMessageOutput>;
  abstract publishMessages(messages: VaaLog[], fifo?: boolean): Promise<SNSPublishMessageOutput>;
}

export default BaseSNS;
