import { getLogger, WormholeLogger } from '../../utils/logger';
import { SNSImplementation, SNSInput, SNSPublishMessageOutput, WhEventType } from './types';
import { WHTransaction, WHTransferRedeemed } from '../../databases/types';

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

  abstract makeSNSInput(
    data: WHTransaction | WHTransferRedeemed,
    type: 'whTx' | 'redeemedTx'
  ): SNSInput;

  abstract createMessages(
    txs: WHTransaction[] | WHTransferRedeemed[],
    eventType: WhEventType,
    fifo?: boolean
  ): Promise<void>;

  abstract publishMessages(
    messages: SNSInput[],
    eventType: WhEventType,
    fifo?: boolean
  ): Promise<SNSPublishMessageOutput>;
}

export default BaseSNS;
