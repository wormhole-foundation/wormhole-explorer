import { HttpClient } from "../../rpc/http/HttpClient";
import { EvmTag } from "../../../domain/entities";
import winston from "../../log";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

export class MoonbeamEvmJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
    override readonly logger = winston.child({ module: "ArbitrumEvmJsonRPCBlockRepository" });

  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    
    super(cfg, httpClient);
  }

  async getBlockHeight(chain: string, finality: EvmTag): Promise<bigint> {
    const chainCfg = this.getCurrentChain(chain);
    const blockNumber: bigint = await super.getBlockHeight(chain, finality);
    let response: { result: boolean };

    let isBlockFinalized = false;
    while (!isBlockFinalized) {
      //await sleep(100);
      // refetch the block by number to get an up-to-date hash
      try {
        const blockFromNumber = await super.getBlock(chain, "latest");

        response = await this.httpClient.post<typeof response>(
          chainCfg.rpc.href,
          {
            jsonrpc: "2.0",
            id: 1,
            method: "moon_isBlockFinalized",
            params: [blockFromNumber.hash],
          },
          { timeout: chainCfg.timeout, retries: chainCfg.retries }
        );
        isBlockFinalized = response.result ?? false;
      } catch (e) {
        this.handleError(chain, e, "getBlockHeight", "eth_getBlockByNumber");
      }
    }
    return blockNumber;
  }
}
