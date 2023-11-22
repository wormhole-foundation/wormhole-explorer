import {
  Commitment,
  Connection,
  PublicKey,
  VersionedTransactionResponse,
  SolanaJSONRPCError,
  Finality,
} from "@solana/web3.js";

import { Fallible, solana } from "../../domain/entities";
import { SolanaSlotRepository } from "../../domain/repositories";

export class Web3SolanaSlotRepository implements SolanaSlotRepository {
  connection: Connection;

  constructor(connection: Connection) {
    this.connection = connection;
  }

  getLatestSlot(commitment: string): Promise<number> {
    return this.connection.getSlot(commitment as Commitment);
  }

  getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, solana.Failure>> {
    return this.connection
      .getBlock(slot, {
        maxSupportedTransactionVersion: 0,
        commitment: finality === "finalized" || finality === "confirmed" ? finality : undefined,
      })
      .then((block) => {
        if (block === null) {
          return Fallible.error<solana.Block, solana.Failure>(
            new solana.Failure(0, "Block not found")
          );
        }
        return Fallible.ok<solana.Block, solana.Failure>({
          ...block,
          transactions: block.transactions.map((tx) => this.mapTx(tx, slot)),
        });
      })
      .catch((err) => {
        if (err instanceof SolanaJSONRPCError) {
          return Fallible.error(new solana.Failure(err.code, err.message));
        }

        return Fallible.error(new solana.Failure(0, err.message));
      });
  }

  getSignaturesForAddress(
    address: string,
    beforeSig: string,
    afterSig: string,
    limit: number
  ): Promise<solana.ConfirmedSignatureInfo[]> {
    return this.connection.getSignaturesForAddress(new PublicKey(address), {
      limit: limit,
      before: beforeSig,
      until: afterSig,
    });
  }

  async getTransactions(sigs: solana.ConfirmedSignatureInfo[]): Promise<solana.Transaction[]> {
    const txs = await this.connection.getTransactions(
      sigs.map((sig) => sig.signature),
      { maxSupportedTransactionVersion: 0 }
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
