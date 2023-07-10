import { ChainId, Network } from "@certusone/wormhole-sdk";

export interface EventHandler<T> {
  name: string;
  getEventSignatureEvm(): string | null;
  handleEventEvm(chainId: ChainId, ...args: any): Promise<T | null>;
  shouldSupportChain(network: Network, chainId: ChainId): boolean;
  persistRecord(record: T): Promise<void>;
  getContractAddressEvm(network: Network, chainId: ChainId): string;
}

export function getEventListener(handler: EventHandler<any>, chainId: ChainId) {
  //@ts-ignore
  return (...args) => {
    // @ts-ignore
    return handler
      .handleEventEvm(chainId, ...args)
      .then((record) => {
        if (record) {
          handler.persistRecord(record);
        }
      })
      .catch((e) => {
        console.error(
          "Unexpected error processing the following event: ",
          chainId,
          ...args
        );
        console.error(e);
      });
  };
}
