export interface AlgorandTransaction {
  applicationId: string;
  blockNumber: number;
  timestamp: number;
  payload: string;
  sender: string;
  logs: string[];
  hash: string;
}
