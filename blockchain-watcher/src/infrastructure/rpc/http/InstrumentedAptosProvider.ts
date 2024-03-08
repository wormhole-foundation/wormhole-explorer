import { AptosClient } from "aptos";
import { Block } from "../../../domain/actions/aptos/PollAptos";
import winston from "winston";

type InstrumentedAptosProviderOptions = Required<Pick<HttpClientOptions, "url" | "chain">> &
  HttpClientOptions;

export class InstrumentedAptosProvider {
  private initialDelay: number = 1_000;
  private maxDelay: number = 60_000;
  private retries: number = 0;
  private timeout: number = 5_000;
  private url: string;
  client: AptosClient;

  private logger: winston.Logger = winston.child({ module: "InstrumentedAptosProvider" });

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
    fromBlock?: number,
    toBlock: number = 100
  ): Promise<any[]> {
    try {
      const params = fromBlock ? { start: fromBlock, limit: toBlock } : { limit: toBlock };

      const result = await this.client.getEventsByEventHandle(
        address,
        eventHandle,
        fieldName!,
        params
      );
      return result;
    } catch (e) {
      throw e;
    }
  }

  public async getTransactionByVersion(version: number): Promise<any> {
    try {
      const result = await this.client.getTransactionByVersion(version);
      return result;
    } catch (e) {
      throw e;
    }
  }

  public async getBlockByHeight(
    blockHeight: number,
    withTransactions?: boolean | undefined
  ): Promise<any> {
    try {
      const result = await this.client.getBlockByHeight(blockHeight, withTransactions);
      return result;
    } catch (e) {
      throw e;
    }
  }

  public async getBlockByVersion(version: number): Promise<any> {
    try {
      const result = await this.client.getBlockByVersion(version);
      return result;
    } catch (e) {
      throw e;
    }
  }

  public async getTransactions(block: Block): Promise<any[]> {
    try {
      const params = block.fromBlock
        ? { start: block.fromBlock, limit: block.toBlock }
        : { limit: block.toBlock };

      const result = await this.client.getTransactions(params);
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
