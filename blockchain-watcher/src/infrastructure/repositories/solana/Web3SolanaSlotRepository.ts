import {
  Commitment,
  Finality,
  PublicKey,
  SolanaJSONRPCError,
  VersionedTransactionResponse,
} from "@solana/web3.js";
import { InstrumentedConnection, ProviderPool } from "@xlabs/rpc-pool";
import { solana } from "../../../domain/entities";
import { Fallible, SolanaFailure } from "../../../domain/errors";
import { SolanaSlotRepository } from "../../../domain/repositories";

export class Web3SolanaSlotRepository implements SolanaSlotRepository {
  constructor(private readonly pool: ProviderPool<InstrumentedConnection>) {}

  getLatestSlot(commitment: string): Promise<number> {
    return this.pool.get().getSlot(commitment as Commitment);
  }

  getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, SolanaFailure>> {
    return this.pool
      .get()
      .getBlock(slot, {
        maxSupportedTransactionVersion: 0,
        commitment: this.normalizeFinality(finality),
      })
      .then((block) => {
        if (block === null) {
          return Fallible.error<solana.Block, SolanaFailure>(
            new SolanaFailure(0, "Block not found")
          );
        }
        return Fallible.ok<solana.Block, SolanaFailure>({
          ...block,
          // TODO: the rpc method returns this field, but it is missing from the lib types
          // which probably needs a version bump
          blockHeight: (block as any).blockHeight,
          transactions: block.transactions.map((tx) => this.mapTx(tx, slot)),
        });
      })
      .catch((err) => {
        if (err instanceof SolanaJSONRPCError) {
          return Fallible.error(new SolanaFailure(err.code, err.message));
        }

        throw err;
      });
  }

  getSignaturesForAddress(
    address: string,
    beforeSig: string,
    afterSig: string,
    limit: number,
    finality?: string
  ): Promise<solana.ConfirmedSignatureInfo[]> {
    return this.pool.get().getSignaturesForAddress(
      new PublicKey(address),
      {
        limit: limit,
        before: beforeSig,
        until: afterSig,
      },
      this.normalizeFinality(finality)
    );
  }

  async getTransactions(
    sigs: solana.ConfirmedSignatureInfo[],
    finality?: string
  ): Promise<solana.Transaction[]> {
    const txs = await this.pool.get().getTransactions(
      sigs.map((sig) => sig.signature),
      { maxSupportedTransactionVersion: 0, commitment: this.normalizeFinality(finality) }
    );

    if (txs.length !== sigs.length) {
      throw new Error(`Expected ${sigs.length} transactions, but got ${txs.length} instead`);
    }

    return txs
      .filter((tx) => tx !== null)
      .map((tx, i) => {
        const message = tx?.transaction.message;
        const accountKeys =
          message?.version === "legacy"
            ? message.accountKeys.map((key) => key.toBase58())
            : message?.staticAccountKeys.map((key) => key.toBase58());

        return {
          ...tx,
          chain: "solana",
          chainId: 1,
          transaction: {
            ...tx?.transaction,
            message: {
              ...tx?.transaction.message,
              accountKeys,
              compiledInstructions: message?.compiledInstructions ?? [],
            },
          },
        } as solana.Transaction;
      });
  }

  private normalizeFinality(finality?: string): Finality | undefined {
    return finality === "finalized" || finality === "confirmed" ? finality : undefined;
  }

  private mapTx(tx: Partial<VersionedTransactionResponse>, slot?: number): solana.Transaction {
    const message = tx?.transaction?.message;
    const accountKeys =
      message?.version === "legacy"
        ? message.accountKeys.map((key) => key.toBase58())
        : message?.staticAccountKeys.map((key) => key.toBase58());

    return {
      ...tx,
      slot: tx.slot || slot,
      transaction: {
        ...tx.transaction,
        message: {
          ...tx?.transaction?.message,
          accountKeys,
          compiledInstructions: message?.compiledInstructions,
        },
      },
    } as solana.Transaction;
  }
}
