import { BigNumber } from "ethers";
import { EvmLog, LogFoundEvent, TransferRedeemed } from "../../domain/entities";

export const evmTransferRedeemedMapper = (
  log: EvmLog,
  _: ReadonlyArray<any>
): LogFoundEvent<TransferRedeemed> => {
  if (!log.blockTime) {
    throw new Error(`Block time is missing for log ${log.logIndex} in tx ${log.transactionHash}`);
  }

  return {
    name: "transfer-redeemed",
    address: log.address,
    chainId: log.chainId,
    txHash: log.transactionHash,
    blockHeight: log.blockNumber,
    blockTime: log.blockTime,
    attributes: {
      emitterChainId: Number(log.topics[1]),
      emitterAddress: log.topics[2],
      sequence: BigNumber.from(log.topics[3]).toNumber(),
    },
  };
};
