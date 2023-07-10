import { TypedEvent } from "@certusone/wormhole-sdk/lib/cjs/ethers-contracts/common";
import { ethers } from "ethers";
import { getEnvironment } from "../environment";
import { ChainId } from "@certusone/wormhole-sdk";

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

export function handleSendEvent(
  chainId: ChainId,
  sequence: ethers.BigNumber,
  deliveryQuote: ethers.BigNumber,
  paymentForExtraReceiverValue: ethers.BigNumber,
  typedEvent: TypedEvent<
    [ethers.BigNumber, ethers.BigNumber, ethers.BigNumber] & {
      sequence: ethers.BigNumber;
      deliveryQuote: ethers.BigNumber;
      paymentForExtraReceiverValue: ethers.BigNumber;
    }
  >
) {
  console.log(
    `Received Send event for Wormhole Relayer Contract, txHash: ${typedEvent.transactionHash}`
  );
  (async () => {
    try {
      const environment = await getEnvironment();
      const txHash = typedEvent.transactionHash;

      const sequence = typedEvent.args.sequence.toString();
      const deliveryQuote = typedEvent.args.deliveryQuote.toString();
      const paymentForExtraReceiverValue =
        typedEvent.args.paymentForExtraReceiverValue.toString();

      const record: WormholeRelayerSendEventRecord = {
        environment,
        chainId,
        txHash,
        sequence,
        deliveryQuote,
        paymentForExtraReceiverValue,
      };

      await persistRecord(record);
      console.log(`Successfully persisted record for transaction: ${txHash}`);
    } catch (e) {
      console.error(e);
    }
  })();
}
