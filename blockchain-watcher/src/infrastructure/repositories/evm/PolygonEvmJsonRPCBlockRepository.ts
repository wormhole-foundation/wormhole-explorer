import { JsonRPCBlockRepositoryCfg, ProviderPoolMap } from "../RepositoriesBuilder";
import { EvmJsonRPCBlockRepository } from "./EvmJsonRPCBlockRepository";
import { BytesLike, ethers } from "ethers";
import { getChainProvider } from "../common/utils";
import { EvmTag } from "../../../domain/entities";

const POLYGON_ROOT_CHAIN_ADDRESS = "0x86E4Dc95c7FBdBf52e33D563BbDB00823894C287";
const FINALIZED = "finalized";

export class PolygonJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: JsonRPCBlockRepositoryCfg, pools: ProviderPoolMap) {
    super(cfg, pools);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    if (finality == FINALIZED) {
      const provider = getChainProvider(chain, this.pool);
      try {
        const rootChain = new ethers.utils.Interface([
          `function getLastChildBlock() external view returns (uint256)`,
        ]);
        const callData = rootChain.encodeFunctionData("getLastChildBlock");

        const callResult: CallResult[] = await provider.post([
          {
            jsonrpc: "2.0",
            id: 1,
            method: "eth_call",
            params: [{ to: POLYGON_ROOT_CHAIN_ADDRESS, data: callData }, FINALIZED],
          },
        ]);

        const block = rootChain.decodeFunctionResult("getLastChildBlock", callResult[0].result)[0];
        return BigInt(block);
      } catch (e) {
        provider.setProviderOffline();
        this.handleError(chain, e, "getBlockHeight", "eth_call");
        throw new Error(`Unable to parse result of eth_call, ${e}`);
      }
    }

    return await super.getBlockHeight(chain, finality);
  }
}

type CallResult = {
  result: BytesLike;
};
