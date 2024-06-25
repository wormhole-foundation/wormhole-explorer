export interface AlgorandTransaction {
  applicationId: string;
  blockNumber: number;
  timestamp: number;
  innerTxs?: {
    applicationId?: string;
    payload?: string;
    sender: string;
    method?: string;
    logs?: string[];
  }[];
  payload: string;
  method: string;
  sender: string;
  hash: string;
}
