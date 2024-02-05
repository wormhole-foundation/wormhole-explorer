import { EvmTag } from "../../../domain/entities";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
  ProviderPoolMap,
} from "./EvmJsonRPCBlockRepository";

export class BscEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, pools: ProviderPoolMap) {
    super(cfg, pools);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);
    const lastBlock = Math.max(Number(blockNumber) - 15, 0);
    return BigInt(lastBlock);
  }
}
