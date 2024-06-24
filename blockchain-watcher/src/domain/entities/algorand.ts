export interface AlgorandTransaction {
  applicationId: string;
  blockNumber: number;
  timestamp: number;
  innerTxs: any; // TODO: Define type
  payload: string;
  sender: string;
  logs: string[];
  hash: string;
}
