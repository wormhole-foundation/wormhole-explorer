import { ChainId, Network } from "@certusone/wormhole-sdk";
import AbstractHandler, { SyntheticEvent } from "./AbstractHandler";
import { Environment } from "../environment";

const CURRENT_VERSION = 1;

type LogMessagePublishedConfig = {
  chains: {
    chainId: ChainId;
    coreContract: string;
  }[];
};

//VAA structure is the same on all chains.
//therefore, as long as the content of the VAA is readable on-chain, we should be able to create this object for all ecosystems
type LogMessagePublished = {};

export default class LogMessagePublishedHandler extends AbstractHandler<LogMessagePublished> {
  constructor(env: Environment, config: any) {
    super("LogMessagePublished", env, config);
  }
  public shouldSupportChain(network: Network, chainId: ChainId): boolean {
    const found = this.config.chains.find((c: any) => c.chainId === chainId);
    return found !== undefined;
  }
  public getEventAbiEvm(): string[] | null {
    return [
      "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel);",
    ];
  }
  public getEventSignatureEvm(): string | null {
    return "LogMessagePublished(address,uint64,uint32,bytes,uint8)";
  }
  public handleEventEvm(
    chainId: ChainId,
    ...args: any
  ): Promise<SyntheticEvent<LogMessagePublished>> {
    //TODO parse event
    const parsedEvent = {};
    return Promise.resolve(this.wrapEvent(chainId, CURRENT_VERSION, parsedEvent));
  }
  public getContractAddressEvm(network: Network, chainId: ChainId): string {
    const found = this.config.chains.find((c: any) => c.chainId === chainId);
    if (found === undefined) {
      throw new Error("Chain not supported");
    }
    return found.coreContract;
  }
}
