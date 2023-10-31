import { ChainId, Network } from "@certusone/wormhole-sdk";

export default abstract class EventHandler<T> {
  public name: string;

  constructor(name: string) {
    this.name = name;
  }

  public abstract getEventSignatureEvm(): string | null;
  public abstract getEventAbiEvm(): string[] | null;
  public abstract handleEventEvm(
    chainId: ChainId,
    ...args: any
  ): Promise<T | null>;
  public abstract shouldSupportChain(
    network: Network,
    chainId: ChainId
  ): boolean;
  public abstract persistRecord(record: T): Promise<void>;
  public abstract getContractAddressEvm(
    network: Network,
    chainId: ChainId
  ): string;

  public getEventListener(handler: EventHandler<any>, chainId: ChainId) {
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
}
