import { ChainId, Network } from "@certusone/wormhole-sdk";
import AbstractWatcher from "./AbstractWatcher";
import AbstractHandler from "../handlers/AbstractHandler";

export default class EvmWatcher extends AbstractWatcher {
  constructor(
    watcherName: string,
    environment: Network,
    events: AbstractHandler<any>[],
    chain: ChainId,
    rpc: string,
    logger: any
  ) {
    super(watcherName, environment, events, chain, rpc, logger);
  }

  async startWebsocketProcessor(): Promise<void> {
    throw new Error("Method not implemented.");
  }
  async startQueryProcessor(): Promise<void> {
    throw new Error("Method not implemented.");
  }
  async startGapProcessor(): Promise<void> {
    throw new Error("Method not implemented.");
  }
}
