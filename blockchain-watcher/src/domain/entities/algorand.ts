export interface AlgorandTransaction {
  applicationId: string;
  blockNumber: number;
  timestamp: number;
  innerTxs?: { sender: string; logs?: string[] }[];
  payload: string;
  sender: string;
  hash: string;
}
