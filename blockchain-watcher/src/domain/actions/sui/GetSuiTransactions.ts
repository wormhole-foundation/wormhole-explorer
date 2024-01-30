import winston from "winston";
import { Range } from "../../entities";
import { SuiRepository } from "../../repositories";
import { SuiTransactionBlockReceipt } from "../../entities/sui";

export class GetSuiTransactions {
  private readonly logger: winston.Logger;

  constructor(private readonly repo: SuiRepository) {
    this.logger = winston.child({ module: "GetSuiTransactions" });
  }

  async execute(range: Range): Promise<SuiTransactionBlockReceipt[]> {
    if (range.from > range.to) {
      this.logger.info(`[sui][exec] Invalid range [from: ${range.from} - to: ${range.to}]`);
      return [];
    }

    let checkpoints = await this.repo.getCheckpoints(range);

    return this.repo.getTransactionBlockReceipts(checkpoints.flatMap((c) => c.transactions));
  }
}
