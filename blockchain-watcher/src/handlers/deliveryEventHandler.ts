//TODO refactor this to use the new event handler system
export {};
// import { TypedEvent } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts/common";
// import { ethers } from "ethers";
// import { getEnvironment } from "../environment";
// import { CHAIN_ID_TO_NAME, ChainId, Network } from "@certusone/wormhole-sdk";
// import AbstractHandler from "./AbstractHandler";

// //TODOD consider additional fields:
// // - timestamp
// // - block number
// // - call data
// // - transaction cost
// // - full transaction receipt
// export type WormholeRelayerDeliveryEventRecord = {
//   environment: string;
//   chainId: ChainId;
//   txHash: string;
//   recipientContract: string;
//   sourceChain: number;
//   sequence: string;
//   deliveryVaaHash: string;
//   status: number;
//   gasUsed: string;
//   refundStatus: number;
//   additionalStatusInfo: string;
//   overridesInfo: string;
// };

// //TODO implement this such that it pushes the event to a database
// async function persistRecord(record: WormholeRelayerDeliveryEventRecord) {
//   console.log(JSON.stringify(record));
// }

// function getEventAbi(): string[] {
//   return [
//     "event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
//   ];
// }

// function getEventSignatureEvm(): string {
//   return "Delivery(address,uint16,uint64,bytes32,uint8,uint256,uint8,bytes,bytes)";
// }

// async function handleEventEvm(
//   chainId: ChainId,
//   eventObj: ethers.Event
// ): Promise<SyntheticEvent<WormholeRelayerDeliveryEventRecord>> {
//   console.log(
//     `Received Delivery event for Wormhole Relayer Contract, txHash: ${eventObj.transactionHash}`
//   );
//   const environment = await getEnvironment();
//   const txHash = eventObj.transactionHash;
//   var abi = getEventAbi();
//   var iface = new ethers.utils.Interface(abi);
//   var parsedLog = iface.parseLog(eventObj);

//   const recipientContract = parsedLog.args[0];
//   const sourceChain = parsedLog.args[1];
//   const sequence = parsedLog.args[2].toString();
//   const deliveryVaaHash = parsedLog.args[3];
//   const status = parsedLog.args[4];
//   const gasUsed = parsedLog.args[5].toString();
//   const refundStatus = parsedLog.args[6];
//   const additionalStatusInfo = parsedLog.args[7];
//   const overridesInfo = parsedLog.args[8];

//   return AbstractHandler.prototype.wrapEvent(chainId, 1, {
//     environment: environment.network,
//     chainId,
//     txHash,
//     recipientContract,
//     sourceChain,
//     sequence,
//     deliveryVaaHash,
//     status,
//     gasUsed,
//     refundStatus,
//     additionalStatusInfo,
//     overridesInfo,
//   });
// }

// function getContractAddressEvm(network: Network, chainId: ChainId): string {
//   return ""; //TODO //getWormholeRelayerAddressWrapped(CHAIN_ID_TO_NAME[chainId], network);
// }

// function shouldSupportChain(network: Network, chainId: ChainId): boolean {
//   return true; //TODO currently the supported chains are determined by the relayer contract, so this is trivially true.
//   //It might not be true in the future.
// }

// const WormholeRelayerEventHandler: AbstractHandler =
//   {
//     name: "Wormhole Relayer Delivery Event Handler",
//     getEventSignatureEvm,
//     getEventAbiEvm: getEventAbi,
//     handleEventEvm,
//     persistRecord,
//     getContractAddressEvm,
//     shouldSupportChain,
//     getEventListener: AbstractHandler.prototype.getEventListener, //TODO not any of this
//   };

// export default WormholeRelayerEventHandler;
