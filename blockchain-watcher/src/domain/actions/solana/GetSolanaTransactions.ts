import { SolanaSlotRepository } from "../../repositories";
import { solana } from "../../entities";
import winston from "winston";

export class GetSolanaTransactions {
  private slotRepository: SolanaSlotRepository;
  protected logger: winston.Logger;

  constructor(slotRepo: SolanaSlotRepository) {
    this.slotRepository = slotRepo;
    this.logger = winston.child({ module: "GetSolanaTransactions" });
  }

  async execute(
    programId: string,
    range: Range,
    opts: GetSolanaTxsOpts
  ): Promise<solana.Transaction[]> {
    if (
      !range.fromBlock.blockTime ||
      !range.toBlock.blockTime ||
      range.fromBlock.blockTime > range.toBlock.blockTime
    ) {
      throw new Error(
        `Invalid slot range: fromSlotBlockTime=${range.fromBlock.blockTime} toSlotBlockTime=${range.toBlock.blockTime}`
      );
    }

    // signatures for address goes back from current sig
    const afterSignature = range.fromBlock.transactions[0]?.transaction.signatures[0];
    let beforeSignature: string | undefined =
      range.toBlock.transactions[range.toBlock.transactions.length - 1]?.transaction.signatures[0];
    if (!afterSignature || !beforeSignature) {
      throw new Error(
        `No signature presents in transactions. After: ${afterSignature}. Before: ${beforeSignature} [slots: ${range.fromBlock.blockTime} - ${range.toBlock.blockTime}]`
      );
    }

    let currentSignaturesCount = opts.signaturesLimit;

    let results: solana.Transaction[] = [];
    while (currentSignaturesCount === opts.signaturesLimit && beforeSignature != undefined) {
      const sigs: solana.ConfirmedSignatureInfo[] =
        await this.slotRepository.getSignaturesForAddress(
          programId,
          beforeSignature,
          afterSignature,
          opts.signaturesLimit,
          opts.commitment
        );

      this.logger.info(
        `Got ${sigs.length} signatures for address ${programId} [blocks: ${
          range.fromBlock.blockTime
        } - ${range.toBlock.blockTime}][sigs: ${afterSignature.substring(
          0,
          5
        )} - ${beforeSignature.substring(0, 5)}]]`
      );

      const txs = await this.slotRepository.getTransactions(sigs, opts.commitment);

      const populatedTxs = txs.map((tx) => {
        tx.chainId = opts.chainId;
        tx.chain = opts.chain;
        return tx;
      });

      results.push(...populatedTxs);
      currentSignaturesCount = sigs.length;
      beforeSignature = sigs.at(-1)?.signature;
    }

    return results;
  }
}

export type GetSolanaTxsOpts = {
  commitment: string;
  signaturesLimit: number;
  chainId: number;
  chain: string;
};

type Range = {
  fromBlock: solana.Block;
  toBlock: solana.Block;
};
