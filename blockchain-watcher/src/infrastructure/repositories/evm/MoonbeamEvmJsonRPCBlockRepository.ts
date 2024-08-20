import { JsonRPCBlockRepositoryCfg, ProviderPoolMap } from "../RepositoriesBuilder";
import { EvmJsonRPCBlockRepository } from "./EvmJsonRPCBlockRepository";
import { getChainProvider } from "../common/utils";
import { setTimeout } from "timers/promises";
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";

const GROW_SLEEP_TIME = 350;
const MAX_ATTEMPTS = 10;

export class MoonbeamEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  override readonly logger = winston.child({ module: "MoonbeamEvmJsonRPCBlockRepository" });

  constructor(cfg: JsonRPCBlockRepositoryCfg, pools: ProviderPoolMap) {
    super(cfg, pools);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    let isBlockFinalized = false;
    let sleepTime = 100;
    let attempts = 0;

    const chainCfg = this.getCurrentChain(chain);
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);

    while (!isBlockFinalized && attempts <= MAX_ATTEMPTS) {
      const provider = getChainProvider(chain, this.pool);
      try {
        await this.sleep(sleepTime);

        const { hash } = await super.getBlock(chain, blockNumber);
        const { result } = await provider.post<BlockIsFinalizedResult>(
          {
            jsonrpc: "2.0",
            id: 1,
            method: "moon_isBlockFinalized",
            params: [hash],
          },
          { timeout: chainCfg.timeout, retries: chainCfg.retries }
        );

        if (!result) {
          isBlockFinalized = false;
          provider.setProviderOffline();
        } else {
          isBlockFinalized = result;
        }

        sleepTime += GROW_SLEEP_TIME; // Increment sleep time
        attempts++; // Increment attempts
      } catch (e) {
        provider.setProviderOffline();
        this.handleError(chain, e, "getBlockHeight", "eth_getBlockByNumber");
        sleepTime = sleepTime += GROW_SLEEP_TIME;
        attempts++;
      }
    }

    if (attempts > MAX_ATTEMPTS) {
      this.logger.warn(`[getBlockHeight] The block ${blockNumber} never ended`);
      throw new Error(`The block ${blockNumber} never ended`);
    }

    return blockNumber;
  }

  private async sleep(sleepTime: number) {
    await setTimeout(sleepTime, null, { ref: false });
  }
}

type BlockIsFinalizedResult = {
  result: boolean;
};
