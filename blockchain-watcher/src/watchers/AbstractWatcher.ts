import { ChainId, Network } from "@certusone/wormhole-sdk";
import { EventHandler } from "../handlers/EventHandler";

export default abstract class AbstractWatcher {
  //store class fields from constructor
  public watcherName: string;
  public environment: Network;
  public events: EventHandler<any>[];
  public chain: ChainId;
  public rpcs: string[];
  public logger: any;

  //TODO add persistence module(s) as class fields
  //or, alternatively, pull necessary config from the persistence module here
  //TODO resumeBlock is needed for the query processor
  constructor(
    watcherName: string,
    environment: Network,
    events: EventHandler<any>[],
    chain: ChainId,
    rpcs: string[],
    logger: any
  ) {
    this.watcherName = watcherName;
    this.environment = environment;
    this.events = events;
    this.chain = chain;
    this.rpcs = rpcs;
    this.logger = logger;
  }

  abstract startWebsocketProcessor(): Promise<void>;

  abstract startQueryProcessor(): Promise<void>;

  abstract startGapProcessor(): Promise<void>;
}
