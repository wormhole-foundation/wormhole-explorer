import { ChainId, Network } from "@certusone/wormhole-sdk";
import { EventHandler } from "../handlers/EventHandler";
import AbstractWatcher from "./AbstractWatcher";

export default class EvmWatcher extends AbstractWatcher {
  constructor(
    watcherName: string,
    environment: Network,
    events: EventHandler<any>[],
    chain: ChainId,
    rpcs: string[],
    logger: any
  ) {
    super(watcherName, environment, events, chain, rpcs, logger);
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
