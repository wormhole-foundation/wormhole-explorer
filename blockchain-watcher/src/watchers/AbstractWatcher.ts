import { ChainId, Network } from "@certusone/wormhole-sdk";
import AbstractHandler from "../handlers/AbstractHandler";

export default abstract class AbstractWatcher {
  //store class fields from constructor
  public watcherName: string;
  public environment: Network;
  public events: AbstractHandler<any>[];
  public chain: ChainId;
  public rpcs: string[];
  public logger: any;

  //TODO add persistence module(s) as class fields
  //or, alternatively, pull necessary config from the persistence module here
  //TODO resumeBlock is needed for the query processor
  constructor(
    watcherName: string,
    environment: Network,
    events: AbstractHandler<any>[],
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
