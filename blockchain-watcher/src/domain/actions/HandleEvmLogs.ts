import { ethers } from "ethers";
import { EvmLog, EvmLogFilter } from "../entities";

export class HandleEvmLogs<T> {
  cfg: HandleEvmLogsConfig;
  mapper: (args: ReadonlyArray<any>) => T;
  target: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleEvmLogsConfig,
    mapper: (args: ReadonlyArray<any>) => T,
    target: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = cfg;
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(logs: EvmLog[]): Promise<T[]> {
    const mappedItems = logs
      .filter(
        (log) =>
          this.cfg.filter.addresses.includes(log.address) &&
          this.cfg.filter.topics.includes(log.topics[0])
      )
      .map((log) => {
        const iface = new ethers.utils.Interface([this.cfg.abi]);
        const parsedLog = iface.parseLog(log);
        return this.mapper(parsedLog.args);
      });

    await this.target(mappedItems);
    // TODO: return a result specifying failures if any
    return mappedItems;
  }
}

export type HandleEvmLogsConfig = {
  filter: EvmLogFilter;
  abi: string;
};
