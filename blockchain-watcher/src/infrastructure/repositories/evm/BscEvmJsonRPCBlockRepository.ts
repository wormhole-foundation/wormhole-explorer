import { JsonRPCBlockRepositoryCfg, ProviderPoolMap } from "../RepositoriesBuilder";
import { EvmJsonRPCBlockRepository } from "./EvmJsonRPCBlockRepository";
import { EvmTag } from "../../../domain/entities";

export class BscEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: JsonRPCBlockRepositoryCfg, pools: ProviderPoolMap) {
    super(cfg, pools);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);
    const lastBlock = Math.max(Number(blockNumber) - 15, 0);
    return BigInt(lastBlock);
  }
}
