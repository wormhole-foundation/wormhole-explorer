import { InstrumentedAptosProvider } from "../../rpc/http/InstrumentedAptosProvider";
import { coalesceChainId } from "@certusone/wormhole-sdk/lib/cjs/utils/consts";
import { AptosEvent } from "../../../domain/entities/aptos";
import { Sequence } from "../../../domain/actions/aptos/PollAptosTransactions";
import winston from "winston";

const CORE_BRIDGE_ADDRESS = "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625";
const EVENT_HANDLE = `${CORE_BRIDGE_ADDRESS}::state::WormholeMessageHandle`;
const FIELD_NAME = "event";

export class AptosJsonRPCBlockRepository {
  private readonly logger: winston.Logger;

  constructor(private readonly client: InstrumentedAptosProvider) {
    this.logger = winston.child({ module: "AptossonRPCBlockRepository" });
  }

  async getSequenceNumber(range: Sequence | undefined): Promise<AptosEvent[]> {
    try {
      const fromSequence = range?.fromSequence ? Number(range?.fromSequence) : undefined;
      const toSequence = range?.toSequence ? Number(range?.toSequence) : undefined;

      const results = await this.client.getEventsByEventHandle(
        CORE_BRIDGE_ADDRESS,
        EVENT_HANDLE,
        FIELD_NAME,
        fromSequence,
        toSequence
      );
      return results;
    } catch (e) {
      this.handleError(e, "getSequenceNumber");
      throw e;
    }
  }

  async getTransactionsForVersions(events: AptosEvent[]): Promise<TransactionsByVersion[]> {
    const transactionsByVersion: TransactionsByVersion[] = [];

    for (const event of events) {
      const result = await this.client.getTransactionByVersion(Number(event.version));
      const tx = {
        consistencyLevel: event.data.consistency_level,
        timestamp: result.timestamp,
        sequence: result.sequence_number,
        version: result.version,
        payload: result.payload,
        sender: result.sender,
        status: result.success,
        events: result.events,
        nonce: event.data.nonce,
        hash: result.hash,
      };
      transactionsByVersion.push(tx);
    }

    return transactionsByVersion;
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[aptos] Error calling ${method}: ${e.message}`);
  }
}

type ResultBlocksHeight = {
  sequence_number: bigint;
};

export type TransactionsByVersion = {
  consistencyLevel?: number;
  timestamp?: string;
  sequence?: string;
  version?: string;
  payload?: string;
  sender?: string;
  status: boolean;
  events: any;
  nonce?: string;
  hash: string;
};

const makeVaaKey = (transactionHash: string, emitter: string, seq: string): string =>
  `${transactionHash}:${coalesceChainId("aptos")}/${emitter}/${seq}`;
