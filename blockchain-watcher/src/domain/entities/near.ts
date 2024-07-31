export interface NearTransaction {
  receiverId: string;
  timestamp: number;
  actions: {
    functionCall: {
      method: string;
      args: string;
    };
  }[];
  height: bigint;
  hash: string;
  logs: {
    outcome: {
      logs: string[];
    };
  }[];
}
