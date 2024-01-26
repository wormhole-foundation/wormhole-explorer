import { SuiEvent, SuiTransactionBlock } from "@mysten/sui.js/client";

// from https://github.com/MystenLabs/sui/blob/9336315169d3a2912260bbe298995af96a239953/sdk/typescript/src/client/types/generated.ts

export interface SuiTransactionBlockReceipt {
  checkpoint: string;
  digest: string;
  errors?: string[];
  events: SuiEvent[];
  timestampMs: string;
  transaction: SuiTransactionBlock;
}

// export interface SuiEvent {
//   id: EventId;
//   packageId: string;
//   parsedJson: unknown;
//   sender: string;
//   timestampMs?: string | null;
//   transactionModule: string;
//   type: string;
// }

// export interface EventId {
//   eventSeq: string;
//   txDigest: string;
// }

// export interface SuiTransactionBlock {
// 	data: TransactionBlockData;
// 	txSignatures: string[];
// }

// export interface SuiTransactionBlockData {
// 	sender: string;
// 	transaction: SuiTransactionBlockKind;
// };

// export type SuiTransactionBlockKind =
// 	/** A system transaction that will update epoch information on-chain. */
// 	| {
// 			kind: 'ChangeEpoch';
// 	  } /** A system transaction used for initializing the initial state of the chain. */
// 	| {
// 			kind: 'Genesis';
// 	  } /** A system transaction marking the start of a series of transactions scheduled as part of a checkpoint */
// 	| {
// 			kind: 'ConsensusCommitPrologue';
// 	  } /** A series of transactions where the results of one transaction can be used in future transactions */
// 	| {
// 			/** Input objects or primitive values */
// 			inputs: SuiCallArg[];
// 			kind: 'ProgrammableTransaction';
// 			/**
// 			 * The transactions to be executed sequentially. A failure in any transaction will result in the
// 			 * failure of the entire programmable transaction block.
// 			 */
// 			transactions: SuiTransaction[];
// 	  } /** A transaction which updates global authenticator state */
// 	| {
// 			kind: 'AuthenticatorStateUpdate';
// 	  } /** The transaction which occurs only at the end of the epoch */
// 	| {
// 			kind: 'EndOfEpochTransaction';
// 	  };

// export type SuiCallArg =
//     | {
//         type: 'object';
//         digest: string;
//         objectId: string;
//         objectType: 'immOrOwnedObject';
//         version: string;
//       }
//     | {
//         type: 'object';
//         initialSharedVersion: string;
//         mutable: boolean;
//         objectId: string;
//         objectType: 'sharedObject';
//       }
//     | {
//         type: 'object';
//         digest: string;
//         objectId: string;
//         objectType: 'receiving';
//         version: string;
//       }
//     | {
//         type: 'pure';
//         value: unknown;
//         valueType?: string | null;
//       };

// export type SuiTransaction =
// 	/** A call to either an entry or a public Move function */
// 	| {
// 			MoveCall: MoveCallSuiTransaction;
// 	  } /**
// 	 * `(Vec<forall T:key+store. T>, address)` It sends n-objects to the specified address. These objects
// 	 * must have store (public transfer) and either the previous owner must be an address or the object
// 	 * must be newly created.
// 	 */
// 	| {
// 			TransferObjects: [SuiArgument[], SuiArgument];
// 	  } /**
// 	 * `(&mut Coin<T>, Vec<u64>)` -> `Vec<Coin<T>>` It splits off some amounts into a new coins with those
// 	 * amounts
// 	 */
// 	| {
// 			SplitCoins: [SuiArgument, SuiArgument[]];
// 	  } /** `(&mut Coin<T>, Vec<Coin<T>>)` It merges n-coins into the first coin */
// 	| {
// 			MergeCoins: [SuiArgument, SuiArgument[]];
// 	  } /**
// 	 * Publishes a Move package. It takes the package bytes and a list of the package's transitive
// 	 * dependencies to link against on-chain.
// 	 */
// 	| {
// 			Publish: string[];
// 	  } /** Upgrades a Move package */
// 	| {
// 			Upgrade: [string[], string, SuiArgument];
// 	  } /**
// 	 * `forall T: Vec<T> -> vector<T>` Given n-values of the same type, it constructs a vector. For non
// 	 * objects or an empty vector, the type tag must be specified.
// 	 */
// 	| {
// 			MakeMoveVec: [string | null, SuiArgument[]];
// 	  };
