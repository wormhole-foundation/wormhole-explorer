import { BytesLike, ethers } from "ethers";
import { EvmTag } from "../../../domain/entities";
import { HttpClient } from "../../rpc/http/HttpClient";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

const POLYGON_ROOT_CHAIN_ADDRESS = "0x86E4Dc95c7FBdBf52e33D563BbDB00823894C287";
const POLYGON_ROOT_CHAIN_RPC = "https://rpc.ankr.com/eth";
const FINALIZED = "finalized";

export class PolygonJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    if (finality == FINALIZED) {
      try {
        const rootChain = new ethers.utils.Interface([
          `function getLastChildBlock() external view returns (uint256)`,
        ]);
        const callData = rootChain.encodeFunctionData("getLastChildBlock");

        const callResult: CallResult[] = await this.httpClient.post(POLYGON_ROOT_CHAIN_RPC, [
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
