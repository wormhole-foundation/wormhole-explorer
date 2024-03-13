import { AptosClient } from "aptos";
import { Range } from "../../../domain/actions/aptos/PollAptos";
import {
  AptosTransactionByVersion,
  AptosTransactionByRange,
  AptosBlockByVersion,
  AptosEvent,
} from "../../../domain/entities/aptos";

type InstrumentedAptosProviderOptions = Required<Pick<HttpClientOptions, "url" | "chain">> &
  HttpClientOptions;

export class InstrumentedAptosProvider {
  private initialDelay: number = 1_000;
  private maxDelay: number = 60_000;
  private retries: number = 0;
  private timeout: number = 5_000;
  private url: string;
  client: AptosClient;

  constructor(options: InstrumentedAptosProviderOptions) {
    options?.initialDelay && (this.initialDelay = options.initialDelay);
    options?.maxDelay && (this.maxDelay = options.maxDelay);
    options?.retries && (this.retries = options.retries);
    options?.timeout && (this.timeout = options.timeout);

    if (!options.url) throw new Error("URL is required");
    this.url = options.url;

    this.client = new AptosClient(this.url);
  }

  public async getEventsByEventHandle(
    address: string,
    eventHandle: string,
    fieldName?: string,
    from?: number,
    limit: number = 100
  ): Promise<AptosEvent[]> {
    try {
      const params = from ? { start: from, limit } : { limit };

      const results = (await this.client.getEventsByEventHandle(
        address,
        eventHandle,
        fieldName!,
        params
      )) as EventsByEventHandle[];

      // Mapped to AptosEvent internal entity
      const aptosEvents: AptosEvent[] = results.map((result) => ({
        guid: result.guid,
        sequence_number: result.data.sequence,
        type: result.type,
        data: result.data,
        version: result.version,
      }));

      return aptosEvents;
    } catch (e) {
      throw e;
    }
  }

  public async getTransactions(block: Range): Promise<AptosTransactionByRange[]> {
    try {
      const params = block.from
        ? { start: block.from, limit: block.limit }
        : { limit: block.limit };

      const results = await this.client.getTransactions(params);

      // Mapped to AptosTransactionByRange internal entity
      const aptosEvents = results
        .map((result: AptosEventRepository) => {
          if (result.events && result.events[0].guid) {
            return {
              version: result.version,
              guid: result.events?.[0]?.guid!,
              sequence_number: result.sequence_number!,
              type: result.type,
              data: result.data,
              events: result.events,
              hash: result.hash,
              payload: result.payload,
            };
          }
        })
        .filter((x) => x !== undefined) as AptosTransactionByRange[];

      return aptosEvents;
    } catch (e) {
      throw e;
    }
  }

  public async getTransactionByVersion(version: number): Promise<AptosTransactionByVersion> {
    try {
      const result = await this.client.getTransactionByVersion(version);
      return result;
    } catch (e) {
      throw e;
    }
  }

  public async getBlockByVersion(version: number): Promise<AptosBlockByVersion> {
    try {
      const result = await this.client.getBlockByVersion(version);
      return result;
    } catch (e) {
      throw e;
    }
  }
}

export type HttpClientOptions = {
  chain?: string;
  url?: string;
  initialDelay?: number;
  maxDelay?: number;
  retries?: number;
  timeout?: number;
};

type AptosEventRepository = {
  sequence_number?: string;
  timestamp?: string;
  success?: boolean;
  version?: string;
  payload?: any;
  events?: any[];
  sender?: string;
  hash: string;
  data?: any;
  type: string;
};

type EventsByEventHandle = {
  version?: string;
  guid: any;
  sequence_number: string;
  type: string;
  data: any;
};
