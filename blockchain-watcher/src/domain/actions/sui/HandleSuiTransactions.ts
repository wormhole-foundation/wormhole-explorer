import { TransactionFoundEvent } from "../../entities";
import { SuiTransactionBlockReceipt } from "../../entities/sui";

export class HandleSuiTransactions {
  constructor(
    private readonly cfg: HandleSuiTransactionsOptions,
    private readonly mapper: (tx: SuiTransactionBlockReceipt) => TransactionFoundEvent,
    private readonly target: (parsed: TransactionFoundEvent[]) => Promise<void>
  ) {}

  public async handle(txs: SuiTransactionBlockReceipt[]): Promise<TransactionFoundEvent[]> {
    const items = txs.filter(this.filterTransaction.bind(this)).map(this.mapper);

    await this.target(items);

    return items;
  }

  private filterTransaction({ events }: SuiTransactionBlockReceipt): boolean {
    return (
      !!events && !!this.cfg.eventTypes && events.some((e) => this.cfg.eventTypes!.includes(e.type))
    );
  }
}

export interface HandleSuiTransactionsOptions {
  eventTypes?: string[];
}
