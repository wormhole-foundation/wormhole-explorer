import { EvmTransactions, FailedRedeemed, TransactionFoundEvent } from "../../domain/entities";
import { BigNumber } from "ethers";

/*
 * Delivery (index_topic_1 address recipientContract, index_topic_2 uint16 sourceChain, index_topic_3 uint64 sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)
 */
export const evmFailedRedeemedMapper = (
  transaction: EvmTransactions,
  args: ReadonlyArray<any>
): TransactionFoundEvent<FailedRedeemed> => {
  return {
    name: "failed-redeemed",
    address: transaction.to,
    txHash: transaction.hash, // TODO
    blockHeight: transaction.blockNumber, // TODO
    blockTime: 1111, // TODO
    attributes: {
      hash: transaction.hash,
      from: transaction.from,
      to: transaction.to,
      status: transaction.status,
      blockNumber: transaction.blockNumber,
      blockTimestamp: transaction.blockTimestamp,
      input: transaction.input,
      methodsByAddress: transaction.methodsByAddress,
    },
  };
};
