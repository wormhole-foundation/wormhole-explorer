import { ChainId, Network } from "@certusone/wormhole-sdk";
import AbstractHandler, { SyntheticEvent } from "./AbstractHandler";
import { Environment } from "../environment";
import { ethers } from "ethers";

const CURRENT_VERSION = 1;

type LogMessagePublishedConfig = {
  chains: {
    chainId: ChainId;
    coreContract: string;
  }[];
};

//VAA structure is the same on all chains.
//therefore, as long as the content of the VAA is readable on-chain, we should be able to create this object for all ecosystems
type LogMessagePublished = {
  timestamp: number;
  nonce: number;
  emitterChain: ChainId;
  emitterAddress: string;
  sequence: number;
  consistencyLevel: number;
  payload: string;
  hash: string;
};

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
  public async handleEventEvm(
    chainId: ChainId,
    event: ethers.Event
  ): Promise<SyntheticEvent<LogMessagePublished>> {
    const abi = this.getEventAbiEvm() as string[];
    const iface = new ethers.utils.Interface(abi);
    const parsedLog = iface.parseLog(event);

    const timestamp = (await event.getBlock()).timestamp; //TODO see if there's a way we can do this without pulling the block header
    const nonce = parsedLog.args[2].toNumber();
    const emitterChain = chainId;
    const emitterAddress = parsedLog.args[0].toString("hex"); //TODO unsure if this is correct
    const sequence = parsedLog.args[1].toNumber();
    const consistencyLevel = parsedLog.args[4].toNumber();
    const payload = parsedLog.args[3].toString("hex"); //TODO unsure if this is correct

    //Encoding from Wormhole ts-sdk
    // timestamp: body.readUInt32BE(0),
    // nonce: body.readUInt32BE(4),
    // emitterChain: body.readUInt16BE(8),
    // emitterAddress: body.subarray(10, 42),
    // sequence: body.readBigUInt64BE(42),
    // consistencyLevel: body[50],
    // payload: body.subarray(51),

    const body = ethers.utils.defaultAbiCoder.encode(
      ["uint32", "uint32", "uint16", "bytes32", "uint64", "uint8", "bytes"],
      [timestamp, nonce, chainId, emitterAddress, sequence, consistencyLevel, payload]
    );
    const hash = this.keccak256(body).toString("hex");

    const parsedEvent = {
      timestamp,
      nonce,
      emitterChain,
      emitterAddress,
      sequence,
      consistencyLevel,
      payload,
      hash,
    };

    return Promise.resolve(this.wrapEvent(chainId, CURRENT_VERSION, parsedEvent));
  }
  public getContractAddressEvm(network: Network, chainId: ChainId): string {
    const found = this.config.chains.find((c: any) => c.chainId === chainId);
    if (found === undefined) {
      throw new Error("Chain not supported");
    }
    return found.coreContract;
  }

  //TODO move to utils
  private keccak256(data: ethers.BytesLike): Buffer {
    return Buffer.from(ethers.utils.arrayify(ethers.utils.keccak256(data)));
  }
}
