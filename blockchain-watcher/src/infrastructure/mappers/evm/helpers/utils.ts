import { EvmTransactionLog, TxStatus } from "../../../../domain/entities";

const TX_STATUS_CONFIRMED = "0x1";
const TX_STATUS_FAILED = "0x0";

export const mappedTxnStatus = (txStatus: string | undefined): string => {
  switch (txStatus) {
    case TX_STATUS_CONFIRMED:
      return TxStatus.Confirmed;
    case TX_STATUS_FAILED:
      return TxStatus.Failed;
    default:
      return TxStatus.Unkonwn;
  }
};

export type LogToNTTTransfer<T> = (log: EvmTransactionLog, ...args: any) => T | undefined;

export type Topics<T> = { [key: string]: LogToNTTTransfer<T> };

export const mapLogDataByTopic = <T>(
  TOPICS: Topics<T>,
  logs: EvmTransactionLog[],
  ...args: any
) => {
  const filterLogs = logs.filter((log) => {
    return TOPICS[log.topics[0]];
  });

  if (!filterLogs) return undefined;

  for (const log of filterLogs) {
    const mapper = TOPICS[log.topics[0]];
    const info = mapper(log, ...args);

    if (info) {
      return info;
    }
  }
};
