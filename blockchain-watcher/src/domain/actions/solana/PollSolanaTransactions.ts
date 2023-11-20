import winston from "winston";
import { RunPollingJob } from "../RunPollingJob";
import { MetadataRepository, SolanaSlotRepository } from "../../repositories";
import { solana } from "../../entities";

export class PollSolanaTransactions extends RunPollingJob {
  private cfg: PollSolanaTransactionsConfig;
  private metadataRepo: MetadataRepository<PollSolanaTransactionsMetadata>;
  private slotRepository: SolanaSlotRepository;

  private latestSlot?: number;
  private slotCursor?: number;
  private lastRange?: { fromSlot: number; toSlot: number };
  protected logger: winston.Logger;

  constructor(
    metadataRepo: MetadataRepository<any>,
    slotRepo: SolanaSlotRepository,
    cfg: PollSolanaTransactionsConfig
  ) {
    super(1_000);

    this.metadataRepo = metadataRepo;
    this.slotRepository = slotRepo;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollSolanaTransactions", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.slotCursor = metadata.lastSlot;
    }
  }

  protected async hasNext(): Promise<boolean> {
    if (this.cfg.toSlot && this.slotCursor && this.slotCursor > this.cfg.toSlot) {
      this.logger.info(
        `Finished processing all slots from ${this.cfg.fromSlot} to ${this.cfg.toSlot}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<solana.Transaction[]> {
    // TODO: report stats
    this.latestSlot = await this.slotRepository.getLatestSlot(this.cfg.commitment);
    const range = this.getSlotRange(this.latestSlot);

    if (range.fromSlot > this.latestSlot) {
      this.logger.info(`Next range is after latest slot, waiting...`);
      return [];
    }

    let toBlock = await this.findValidBlock(range.toSlot, (slot) => slot - 1);
    let fromBlock = await this.findValidBlock(range.fromSlot, (slot) => slot + 1);

    if (fromBlock.blockTime > toBlock.blockTime) {
      // TODO: validate if this is correct
      throw new Error(
        `Invalid slot range: fromSlotBlockTime=${fromBlock.blockTime} toSlotBlockTime=${toBlock.blockTime}`
      );
    }

    // signatures for address goes back from current sig
    const afterSignature = fromBlock.transactions[0].transaction.signatures[0];
    let beforeSignature = toBlock.transactions[0].transaction.signatures[0];
    let currentSignaturesCount = this.cfg.signaturesLimit;

    let results: solana.Transaction[] = [];
    while (currentSignaturesCount === this.cfg.signaturesLimit) {
      const sigs = await this.slotRepository.getSignaturesForAddress(
        this.cfg.programId,
        beforeSignature,
        afterSignature,
        this.cfg.signaturesLimit
      );
      this.logger.debug(
        `Got ${sigs.length} signatures for address ${this.cfg.programId} between ${beforeSignature} and ${afterSignature}`
      );

      const txs = await this.slotRepository.getTransactions(sigs);
      results.push(...txs);
      currentSignaturesCount = sigs.length;
    }

    this.lastRange = range;
    return results;
  }

  protected async persist(): Promise<void> {
    this.slotCursor = this.lastRange?.toSlot;
    if (this.slotCursor) {
      await this.metadataRepo.save(this.cfg.id, { lastSlot: this.slotCursor });
    }
  }

  private getSlotRange(latestSlot: number): { fromSlot: number; toSlot: number } {
    let fromSlot = this.slotCursor ? this.slotCursor + 1 : this.cfg.fromSlot ?? latestSlot;
    // cfg.fromSlot is present and is greater than current slot height, then we allow to skip slots.
    if (this.slotCursor && this.cfg.fromSlot && this.cfg.fromSlot > this.slotCursor) {
      fromSlot = this.cfg.fromSlot;
    }

    let toSlot = Math.min(fromSlot + this.cfg.slotBatchSize, latestSlot);
    // limit toSlot to configured toSlot
    if (this.cfg.toSlot && toSlot > this.cfg.toSlot) {
      toSlot = this.cfg.toSlot;
    }

    if (fromSlot > toSlot) {
      throw new Error(`Invalid slot range: fromSlot=${fromSlot} toSlot=${toSlot}`);
    }

    return { fromSlot, toSlot };
  }

  /**
   * Recursively find a valid block for the given slot.
   * @param slot - the starting slot
   * @param nextSlot - the function to get the next slot (to go up or down for example)
   * @returns a block if found, otherwise throws an error
   */
  private async findValidBlock(
    slot: number,
    nextSlot: (slot: number) => number
  ): Promise<solana.Block> {
    const blockResult = await this.slotRepository.getBlock(slot);
    if (blockResult.isOk()) {
      return Promise.resolve(blockResult.getValue());
    }

    if (blockResult.getError().skippedSlot() || blockResult.getError().noBlockOrBlockTime()) {
      return this.findValidBlock(nextSlot(slot), nextSlot);
    }

    throw blockResult.getError();
  }
}

export class PollSolanaTransactionsConfig {
  id: string;
  commitment: string;
  programId: string;
  fromSlot?: number;
  toSlot?: number;
  slotBatchSize: number = 10_000;
  signaturesLimit: number = 1_000;
  interval?: number = 5_000;

  constructor(id: string, programId: string, commitment?: string, slotBatchSize?: number) {
    this.id = id;
    this.commitment = commitment ?? "confirmed";
    this.programId = programId;
    this.slotBatchSize = slotBatchSize ?? 10_000;
  }
}

export type PollSolanaTransactionsMetadata = {
  lastSlot: number;
};
