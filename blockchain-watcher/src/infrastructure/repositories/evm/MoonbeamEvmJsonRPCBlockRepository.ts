import { HttpClient } from "../../rpc/http/HttpClient";
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

const MAX_ATTEMPTS = 30;
let isBlockFinalized = false;
let sleepTime = 100;
let attempts = 0;

export class MoonbeamEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  override readonly logger = winston.child({ module: "ArbitrumEvmJsonRPCBlockRepository" });

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const chainCfg = this.getCurrentChain(chain);
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);

    while (!isBlockFinalized && attempts < MAX_ATTEMPTS) {
      try {
        this.sleep();

        const { hash } = await super.getBlock(chain, blockNumber);

        // Consult the block with your hash, asking your status
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

        isBlockFinalized = result ?? false;
        attempts++;
      } catch (e) {
        this.handleError(chain, e, "getBlockHeight", "eth_getBlockByNumber");
        attempts++;
      }
    }

    if (attempts > MAX_ATTEMPTS)
      this.logger.warn(`[getBlockHeight] The block ${blockNumber} never ended`);

    return blockNumber;
  }

  private sleep() {
    sleepTime = sleepTime + 50;
    return new Promise((resolve) => setTimeout(resolve, sleepTime));
  }
}

type BlockIsFinalizedResult = {
  result: boolean;
};
