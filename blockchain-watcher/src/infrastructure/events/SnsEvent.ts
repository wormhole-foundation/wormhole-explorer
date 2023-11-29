import { LogFoundEvent } from "../../domain/entities";

export class SnsEvent {
    trackId: string;
    source: string;
    event: string;
    timestamp: string;
    version: string;
    data: Record<string, any>;
  
    constructor(
      trackId: string,
      source: string,
      event: string,
      timestamp: string,
      version: string,
      data: Record<string, any>
    ) {
      this.trackId = trackId;
      this.source = source;
      this.event = event;
      this.timestamp = timestamp;
      this.version = version;
      this.data = data;
    }
  
    static fromLogFoundEvent<T>(logFoundEvent: LogFoundEvent<T>): SnsEvent {
      return new SnsEvent(
        `chain-event-${logFoundEvent.txHash}-${logFoundEvent.blockHeight}`,
        "blockchain-watcher",
        logFoundEvent.name,
        new Date().toISOString(),
        "1",
        {
          chainId: logFoundEvent.chainId,
          emitter: logFoundEvent.address,
          txHash: logFoundEvent.txHash,
          blockHeight: logFoundEvent.blockHeight.toString(),
          blockTime: new Date(logFoundEvent.blockTime * 1000).toISOString(),
          attributes: logFoundEvent.attributes,
        }
      );
    }
  }