import {
  Commitment,
  Connection,
  Finality,
  PublicKey,
  VersionedTransactionResponse,
  SolanaJSONRPCError,
} from "@solana/web3.js";

import { solana } from "../../../domain/entities";
import { SolanaSlotRepository } from "../../../domain/repositories";
import { Fallible, SolanaFailure } from "../../../domain/errors";
import winston from "../../../infrastructure/log";

export class Web3SolanaSlotRepository implements SolanaSlotRepository {
  private connection: Connection;
  private logger: winston.Logger = winston.child({ module: "Web3SolanaSlotRepository" });

  constructor(connection: Connection) {
    this.connection = connection;
    this.logger.info(`Using RPC node ${new URL(connection.rpcEndpoint).hostname}`);
  }

  getLatestSlot(commitment: string): Promise<number> {
    return this.connection.getSlot(commitment as Commitment);
  }

  getBlock(slot: number, finality?: string): Promise<Fallible<solana.Block, SolanaFailure>> {
    return this.connection
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
    return this.connection.getSignaturesForAddress(new PublicKey(address), {
      limit: limit,
      before: beforeSig,
      until: afterSig,
    },
    this.normalizeFinality(finality)
    );
  }

  async getTransactions(sigs: solana.ConfirmedSignatureInfo[], finality?: string): Promise<solana.Transaction[]> {
    const txs = await this.connection.getTransactions(
      sigs.map((sig) => sig.signature),
      { maxSupportedTransactionVersion: 0,
        commitment: this.normalizeFinality(finality)
      }
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
