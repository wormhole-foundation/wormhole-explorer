import { EvmTransaction } from "../../../../domain/entities";
import { NTTTransfer } from "./ntt";

interface ProcessQueueData {
  targetTxn: EvmTransaction;
  nttTransferInfo: NTTTransfer;
}

const TARGET_TO_SOURCE_CHAIN_MAPPING = {
  "transfer-redeemed": {
    sourceChainEventName: "transfer-sent",
    topic: "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf",
  },
  "received-relayed-message": {
    sourceChainEventName: "send-transceiver-message",
    topic: "0x9716fe52fe4e02cf924ae28f19f5748ef59877c6496041b986fbad3dae6a8ecf",
  },
};

// messageId (numberic) -> {sourceChainTxn, nttTransferInfo}

class TimeDistributionMetrics {
  private processQueue: Map<number, ProcessQueueData> = new Map();

  constructor() {
    this.scanProcessQueue();
  }

  public addProcessQueueData(targetTxn: EvmTransaction, nttTransferInfo: NTTTransfer) {
    const currentProcessQueueSize = this.processQueue.size;
    this.processQueue.set(nttTransferInfo.messageId, { targetTxn, nttTransferInfo });
    if (currentProcessQueueSize === 0) {
      this.scanProcessQueue();
    }
  }

  private async scanProcessQueue() {
    while (this.processQueue.size > 0) {
      // scan source chain for corresponding TOPIC using blockRepo.getFilteredLogs
    }
  }
}
