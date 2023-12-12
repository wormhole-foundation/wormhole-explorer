import { HttpClient } from "../../rpc/http/HttpClient";
import { setTimeout } from 'timers/promises'
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

const GROW_SLEEP_TIME = 50;
const MAX_ATTEMPTS = 20;

export class MoonbeamEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  override readonly logger = winston.child({ module: "MoonbeamEvmJsonRPCBlockRepository" });
  private isBlockFinalized = false;
  private sleepTime = 100;
  private attempts = 0;

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const chainCfg = this.getCurrentChain(chain);
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);

    while (!this.isBlockFinalized && this.attempts <= MAX_ATTEMPTS) {
      try {
        await this.sleep();

        const { hash } = await super.getBlock(chain, blockNumber);

        const { result } = await this.httpClient.post<BlockIsFinalizedResult>(
          chainCfg.rpc.href,
          {
            jsonrpc: "2.0",
            id: 1,
            method: "moon_isBlockFinalized",
            params: [hash],
          },
          { timeout: chainCfg.timeout, retries: chainCfg.retries }
        );

        this.isBlockFinalized = result ?? false;
        this.attempts++;
      } catch (e) {
        this.handleError(chain, e, "getBlockHeight", "eth_getBlockByNumber");
        this.attempts++;
      }
    }

    if (this.attempts > MAX_ATTEMPTS)
      this.logger.error(`[getBlockHeight] The block ${blockNumber} never ended`);

    return blockNumber;
  }

  private async sleep() {
    this.sleepTime = this.sleepTime + GROW_SLEEP_TIME;
    await setTimeout(this.sleepTime, null, {ref: false})
  }
}

type BlockIsFinalizedResult = {
  result: boolean;
};
