import { HttpClient } from "../../rpc/http/HttpClient";
import {
  EvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepositoryCfg,
} from "./EvmJsonRPCBlockRepository";

export class PolygonJsonRPCBlockRepository extends EvmJsonRPCBlockRepository {
  constructor(cfg: EvmJsonRPCBlockRepositoryCfg, httpClient: HttpClient) {
    super(cfg, httpClient);
  }
}
