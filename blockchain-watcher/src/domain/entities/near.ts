export interface NearTransaction {
  blockHeight: bigint;
  receiverId: string;
  signerId: string;
  timestamp: number;
  chainId: number;
  hash: string;
  actions: {
    functionCall: {
      method: string;
      args: string;
    };
  }[];
  logs: {
    outcome: {
      logs: string[];
    };
  }[];
}
