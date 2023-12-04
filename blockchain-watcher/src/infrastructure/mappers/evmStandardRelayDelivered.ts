import { BigNumber } from "ethers";
import { EvmLog, LogFoundEvent, StandardRelayDelivered } from "../../domain/entities";

/*
 * Delivery (index_topic_1 address recipientContract, index_topic_2 uint16 sourceChain, index_topic_3 uint64 sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)
 */
export const evmStandardRelayDelivered = (
  log: EvmLog,
  args: ReadonlyArray<any>
): LogFoundEvent<StandardRelayDelivered> => {
  if (!log.blockTime) {
    throw new Error(`Block time is missing for log ${log.logIndex} in tx ${log.transactionHash}`);
  }

  return {
    name: "standard-relay-delivered",
    address: log.address,
    chainId: log.chainId,
    txHash: log.transactionHash,
    blockHeight: log.blockNumber,
    blockTime: log.blockTime,
    attributes: {
      recipientContract: args[0],
      sourceChain: BigNumber.from(args[1]).toNumber(),
      sequence: BigNumber.from(args[2]).toNumber(),
      deliveryVaaHash: args[3],
      status: BigNumber.from(args[4]).toNumber(),
      gasUsed: BigNumber.from(args[5]).toNumber(),
      refundStatus: BigNumber.from(args[6]).toNumber(),
      additionalStatusInfo: args[7],
      overridesInfo: args[8],
    },
  };
};
