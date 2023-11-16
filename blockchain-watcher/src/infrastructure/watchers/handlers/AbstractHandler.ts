import { ChainId, Network } from "@certusone/wormhole-sdk";
import { v4 as uuidv4 } from "uuid";
import { Environment } from "../environment";
const { createHash } = require("crypto");

export type SyntheticEvent<T> = {
  eventName: string;
  eventVersion: number;
  eventChain: ChainId;
  observationTimestamp: number;
  uuid: string; //UUID for the event, good for deduping
  dataHash: string; //sha256 hash of the event data, good for deduping
  data: T;
};

export default abstract class AbstractHandler<T> {
  public name: string;
  public environment: Environment;
  public config: any;

  constructor(name: string, environment: Environment, config: any) {
    this.name = name;
    this.environment = environment;
    this.config = config;
  }

  //These top level functions must always be implemented
  public abstract shouldSupportChain(network: Network, chainId: ChainId): boolean;

  //These functions must be implemented if an EVM chain is supported.

  //Event to be listened for in ABI format. Example:
  //"event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
  public abstract getEventAbiEvm(): string[] | null;

  //Event to be listened for in signature format. Example:
  //"Delivery(address,uint16,uint64,bytes32,uint8,uint256,uint8,bytes,bytes)"
  public abstract getEventSignatureEvm(): string | null;

  //This function will be called when a subscribed event is received from the ethers provider.
  //TODO pretty sure the ...args is always an ethers.Event object
  public abstract handleEventEvm(chainId: ChainId, ...args: any): Promise<SyntheticEvent<T>>;
  public abstract getContractAddressEvm(network: Network, chainId: ChainId): string;

  //*** Non-abstract functions

  //Wrapper function to hand into EVM rpc provider.
  //The wrapper is necessary otherwise we can't figure out which chain ID the event came from.
  public getEventListener(handler: AbstractHandler<T>, chainId: ChainId) {
    //@ts-ignore
    return (...args) => {
      // @ts-ignore
      return handler
        .handleEventEvm(chainId, ...args)
        .then((records) => {
          if (records) {
            //TODO persist records. Unsure how exactly this happens atm.
            //handler.persistRecord(record);
          }
        })
        .catch((e) => {
          console.error("Unexpected error processing the following event: ", chainId, ...args);
          console.error(e);
        });
    };
  }

  public getName(): string {
    return this.name;
  }

  public generateUuid(): string {
    return uuidv4();
  }

  public getEnvironment(): Environment {
    return this.environment;
  }

  public getConfig(): any {
    return this.config;
  }

  protected wrapEvent(chainId: ChainId, version: number, data: T): SyntheticEvent<T> {
    return {
      eventName: this.name,
      eventVersion: version,
      eventChain: chainId,
      observationTimestamp: Date.now(),
      uuid: this.generateUuid(),
      dataHash: createHash("sha256").update(JSON.stringify(data)).digest("hex"),
      data: data,
    };
  }
}
