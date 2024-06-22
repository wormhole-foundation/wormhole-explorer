export interface AlgorandTransaction {
  type: string;
  applicationTransaction: any;
  id: string;
  sender: string;
  blockNumber: number;
  applicationArgs: any;
  timestamp: number;
  innerTxs: any; // TODO: Define type
  logs: string[];
}
