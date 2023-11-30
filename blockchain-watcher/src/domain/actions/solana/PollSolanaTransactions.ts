import winston from "winston";
import { RunPollingJob } from "../RunPollingJob";
import { MetadataRepository, SolanaSlotRepository, StatRepository } from "../../repositories";
import { solana } from "../../entities";

export class PollSolanaTransactions extends RunPollingJob {
  private cfg: PollSolanaTransactionsConfig;
  private metadataRepo: MetadataRepository<PollSolanaTransactionsMetadata>;
  private slotRepository: SolanaSlotRepository;
  private statsRepo: StatRepository;
  private latestSlot?: number;
  private slotCursor?: number;
  private lastRange?: Range;
  protected logger: winston.Logger;

  constructor(
    metadataRepo: MetadataRepository<any>,
    slotRepo: SolanaSlotRepository,
    statsRepo: StatRepository,
    cfg: PollSolanaTransactionsConfig
  ) {
    super(1_000, cfg.id, statsRepo);

    this.metadataRepo = metadataRepo;
    this.slotRepository = slotRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollSolanaTransactions", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.slotCursor = metadata.lastSlot;
    }
  }

  async hasNext(): Promise<boolean> {
    if (this.cfg.toSlot && this.slotCursor && this.slotCursor >= this.cfg.toSlot) {
      this.logger.info(
        `Finished processing all slots from ${this.cfg.fromSlot} to ${this.cfg.toSlot}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<solana.Transaction[]> {
    this.report();
    this.latestSlot = await this.slotRepository.getLatestSlot(this.cfg.commitment);
    const range = this.getSlotRange(this.latestSlot);

    if (range.fromSlot > this.latestSlot) {
      this.logger.info(`Next range is after latest slot, waiting...`);
      return [];
    }

    let toBlock = await this.findValidBlock(range.toSlot, (slot) => slot - 1);
    let fromBlock = await this.findValidBlock(range.fromSlot, (slot) => slot + 1);

    if (!fromBlock.blockTime || !toBlock.blockTime || fromBlock.blockTime > toBlock.blockTime) {
      // TODO: validate if this is correct
      throw new Error(
        `Invalid slot range: fromSlotBlockTime=${fromBlock.blockTime} toSlotBlockTime=${toBlock.blockTime}`
      );
    }

    // signatures for address goes back from current sig
    const afterSignature = fromBlock.transactions[0]?.transaction.signatures[0];
    let beforeSignature: string | undefined =
      toBlock.transactions[toBlock.transactions.length - 1]?.transaction.signatures[0];
    if (!afterSignature || !beforeSignature) {
      throw new Error(
        `No signature presents in transactions. After: ${afterSignature}. Before: ${beforeSignature} [slots: ${range.fromSlot} - ${range.toSlot}]`
      );
    }

    let currentSignaturesCount = this.cfg.signaturesLimit;

    let results: solana.Transaction[] = [];
    while (currentSignaturesCount === this.cfg.signaturesLimit && beforeSignature != undefined) {
      const sigs: solana.ConfirmedSignatureInfo[] =
        await this.slotRepository.getSignaturesForAddress(
          this.cfg.programId,
          beforeSignature,
          afterSignature,
          this.cfg.signaturesLimit,
          this.cfg.commitment
        );
      this.logger.debug(
        `Got ${sigs.length} signatures for address ${this.cfg.programId} [slots: ${
          range.fromSlot
        } - ${range.toSlot}][sigs: ${afterSignature.substring(0, 5)} - ${beforeSignature.substring(
          0,
          5
        )}]]`
      );

      const txs = await this.slotRepository.getTransactions(sigs, this.cfg.commitment);
      results.push(...txs);
      currentSignaturesCount = sigs.length;
      beforeSignature = sigs.at(-1)?.signature;
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

  private getSlotRange(latestSlot: number): Range {
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
      throw new Error(
        `Invalid slot range: fromSlot=${fromSlot} toSlot=${toSlot}. Might be cause we are up to date.`
      );
    }

    return { fromSlot, toSlot };
  }

  private report(): void {
    const labels = {
      job: this.cfg.id,
      chain: "solana",
      commitment: this.cfg.commitment,
    };
    this.statsRepo.count("job_execution", labels);
    this.statsRepo.measure("block_height", BigInt(this.latestSlot ?? 0), labels);
    this.statsRepo.measure("block_cursor", BigInt(this.slotCursor ?? 0n), labels);
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
    const blockResult = await this.slotRepository.getBlock(slot, this.cfg.commitment);
    if (blockResult.isOk()) {
      return Promise.resolve(blockResult.getValue());
    }

    if (blockResult.getError().skippedSlot() || blockResult.getError().noBlockOrBlockTime()) {
      this.logger.warn(`No block found for slot ${slot}, trying next slot ${nextSlot(slot)}`);
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
  slotBatchSize: number = 1_000;
  signaturesLimit: number = 500;
  interval?: number = 5_000;

  constructor(id: string, programId: string, commitment?: string, slotBatchSize?: number) {
    this.id = id;
    this.commitment = commitment ?? "finalized";
    this.programId = programId;
    this.slotBatchSize = slotBatchSize ?? 10_000;
  }
}

export type PollSolanaTransactionsMetadata = {
  lastSlot: number;
};

type Range = {
  fromSlot: number;
  toSlot: number;
};
