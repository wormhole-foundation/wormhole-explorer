import { EvmTag } from "../../../domain/entities";
import { HttpClient } from "../../rpc/http/HttpClient";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

export class BscEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);
    const lastBlock = Math.max(Number(blockNumber) - 15, 0);
    return BigInt(lastBlock);
  }
}
