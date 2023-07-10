import { TypedEvent } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts/common";
import { ethers } from "ethers";
import {
  getEnvironment,
  getWormholeRelayerAddressWrapped,
} from "../environment";
import { CHAIN_ID_TO_NAME, ChainId, Network } from "@certusone/wormhole-sdk";
import { EventHandler } from "./EventHandler";

//TODOD consider additional fields:
// - timestamp
// - entire transaction receipt
// - deduplication info
export type WormholeRelayerSendEventRecord = {
  environment: string;
  chainId: ChainId;
  txHash: string;
  sequence: string;
  deliveryQuote: string;
  paymentForExtraReceiverValue: string;
};

//TODO implement this such that it pushes the event to a database
async function persistRecord(record: WormholeRelayerSendEventRecord) {
  console.log(JSON.stringify(record));
}

async function handleEventEvm(
  chainId: ChainId,
  eventObj: ethers.Event
): Promise<WormholeRelayerSendEventRecord | null> {
  console.log(
    `Received Send event for Wormhole Relayer Contract, txHash: ${eventObj.transactionHash}`
  );
  const abi = [
    "event SendEvent(uint64 indexed sequence, uint256 deliveryQuote, uint256 paymentForExtraReceiverValue)",
  ];
  var iface = new ethers.utils.Interface(abi);
  var parsedLog = iface.parseLog(eventObj);

  return {
    environment: await getEnvironment(),
    chainId: chainId,
    txHash: eventObj.transactionHash,
    sequence: parsedLog.args[0].toString(),
    deliveryQuote: parsedLog.args[1].toString(),
    paymentForExtraReceiverValue: parsedLog.args[2].toString(),
  };
}

function getContractAddressEvm(network: Network, chainId: ChainId): string {
  return getWormholeRelayerAddressWrapped(CHAIN_ID_TO_NAME[chainId], network);
}

function shouldSupportChain(network: Network, chainId: ChainId): boolean {
  return true; //TODO currently the supported chains are determined by the relayer contract, so this is trivially true.
  //It might not be true in the future.
}

function getEventSignatureEvm(): string {
  return "SendEvent(uint64,uint256,uint256)";
}

const WormholeRelayerSendEventHandler: EventHandler<WormholeRelayerSendEventRecord> =
  {
    name: "Wormhole Relayer Send Event Handler",
    getEventSignatureEvm,
    handleEventEvm,
    persistRecord,
    getContractAddressEvm,
    shouldSupportChain,
  };

export default WormholeRelayerSendEventHandler;
