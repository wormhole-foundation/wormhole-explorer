import { InstrumentedConnectionWrapper } from "../../rpc/http/InstrumentedConnectionWrapper";
import { Fallible, SolanaFailure } from "../../../domain/errors";
import { SolanaSlotRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import { solana } from "../../../domain/entities";
import winston from "../../log";
import {
  VersionedTransactionResponse,
  SolanaJSONRPCError,
  Commitment,
  PublicKey,
  Finality,
  VersionedBlockResponse,
} from "@solana/web3.js";

export class Web3SolanaSlotRepository implements SolanaSlotRepository {
  protected readonly logger;

  constructor(private readonly pool: ProviderPool<InstrumentedConnectionWrapper>) {
    this.logger = winston.child({ module: "Web3SolanaSlotRepository" });
  }

  async getLatestSlot(commitment: string): Promise<number> {
    const provider = this.pool.get();
    return this.withProvider<number>(
      provider,
      (provider) => provider.getSlot(commitment as Commitment),
      "getSlot"
    );
  }

  async getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, SolanaFailure>> {
    const provider = this.pool.get();
    return this.withProvider<VersionedBlockResponse | null>(
      provider,
      (provider) =>
        provider.getBlock(slot, {
          maxSupportedTransactionVersion: 0,
          commitment: this.normalizeFinality(finality),
        }),
      "getBlock"
    )
      .then((block) => {
        if (block === null) {
          // In this case we throw and error and we retry the request
          throw new Error("Unable to parse result of getBlock");
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
          // We skip the block if it is not available (e.g Slot N was skipped - Error code: -32007, -32009)
          return Fallible.error(new SolanaFailure(err.code, err.message));
        }

        this.logger.error(
          `[solana][getBlock] Cannot process this slot: ${slot}}, error ${JSON.stringify(
            err
          )} on ${provider.getUrl()}`
        );
        provider.setProviderOffline();
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
    const provider = this.pool.get();
    return this.withProvider<solana.ConfirmedSignatureInfo[]>(
      provider,
      (provider) =>
        provider.getSignaturesForAddress(
          new PublicKey(address),
          {
            limit: limit,
            before: beforeSig,
            until: afterSig,
          },
          this.normalizeFinality(finality)
        ),
      "getSignaturesForAddress"
    );
  }

  async getTransactions(
    sigs: solana.ConfirmedSignatureInfo[],
    finality?: string
  ): Promise<solana.Transaction[]> {
    const provider = this.pool.get();
    const txs = await this.withProvider<(VersionedTransactionResponse | null)[]>(
      provider,
      (provider) =>
        provider.getTransactions(
          sigs.map((sig) => sig.signature),
          { maxSupportedTransactionVersion: 0, commitment: this.normalizeFinality(finality) }
        ),
      "getTransactions"
    );

    if (txs.length !== sigs.length) {
      this.logger.error(
        `[solana][getTransactions] Expected ${sigs.length} transactions, but got ${
          txs.length
        } instead on ${provider.getUrl()}`
      );
      provider.setProviderOffline();
      throw new Error("Unable to parse result of getTransactions");
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

  private async withProvider<T>(
    provider: InstrumentedConnectionWrapper,
    fn: (provider: InstrumentedConnectionWrapper) => Promise<T>,
    method: string
  ) {
    try {
      return await fn(provider);
    } catch (e) {
      this.logger.error(`[solana][${method}] Error getting result on ${provider.getUrl()}`);
      provider.setProviderOffline();
      throw e;
    }
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
