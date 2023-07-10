import { TypedEvent } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts/common";
import { ethers } from "ethers";
import { getEnvironment } from "../environment";
import { ChainId } from "@certusone/wormhole-sdk";

//TODOD consider additional fields:
// - timestamp
// - block number
// - call data
// - transaction cost
// - full transaction receipt
export type WormholeRelayerDeliveryEventRecord = {
  environment: string;
  chainId: ChainId;
  txHash: string;
  recipientContract: string;
  sourceChain: number;
  sequence: string;
  deliveryVaaHash: string;
  status: number;
  gasUsed: string;
  refundStatus: number;
  additionalStatusInfo: string;
  overridesInfo: string;
};

//TODO implement this such that it pushes the event to a database
async function persistRecord(record: WormholeRelayerDeliveryEventRecord) {
  console.log(JSON.stringify(record));
}

export function handleDeliveryEvent(
  chainId: ChainId,
  recipientContract: string,
  sourceChain: number,
  sequence: ethers.BigNumber,
  deliveryVaaHash: string,
  status: number,
  gasUsed: ethers.BigNumber,
  refundStatus: number,
  additionalStatusInfo: string,
  overridesInfo: string,
  typedEvent: TypedEvent<
    [
      string,
      number,
      ethers.BigNumber,
      string,
      number,
      ethers.BigNumber,
      number,
      string,
      string
    ] & {
      recipientContract: string;
      sourceChain: number;
      sequence: ethers.BigNumber;
      deliveryVaaHash: string;
      status: number;
      gasUsed: ethers.BigNumber;
      refundStatus: number;
      additionalStatusInfo: string;
      overridesInfo: string;
    }
  >
) {
  console.log(
    `Received Delivery event for Wormhole Relayer Contract, txHash: ${typedEvent.transactionHash}`
  );
  (async () => {
    try {
      const environment = await getEnvironment();
      const txHash = typedEvent.transactionHash;
      const recipientContract = typedEvent.args.recipientContract;
      const sourceChain = typedEvent.args.sourceChain;
      const sequence = typedEvent.args.sequence.toString();
      const deliveryVaaHash = typedEvent.args.deliveryVaaHash;
      const status = typedEvent.args.status;
      const gasUsed = typedEvent.args.gasUsed.toString();
      const refundStatus = typedEvent.args.refundStatus;
      const additionalStatusInfo = typedEvent.args.additionalStatusInfo;
      const overridesInfo = typedEvent.args.overridesInfo;

      const record: WormholeRelayerDeliveryEventRecord = {
        environment,
        chainId,
        txHash,
        recipientContract,
        sourceChain,
        sequence,
        deliveryVaaHash,
        status,
        gasUsed,
        refundStatus,
        additionalStatusInfo,
        overridesInfo,
      };

      await persistRecord(record);
      console.log(
        `Successfully persisted delivery record for transaction: ${txHash}`
      );
    } catch (e) {
      console.error(e);
    }
  })();
}
