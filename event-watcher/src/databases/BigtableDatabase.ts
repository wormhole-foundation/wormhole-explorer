import { ChainName, coalesceChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { parseVaa } from '@certusone/wormhole-sdk/lib/cjs/vaa/wormhole';
import { Bigtable } from '@google-cloud/bigtable';
import {
  assertEnvironmentVariable,
  chunkArray,
  sleep,
} from '../common';
import { cert, initializeApp } from 'firebase-admin/app';
import { getFirestore } from 'firebase-admin/firestore';
import { Database } from './Database';
import {
  BigtableMessagesResultRow,
  BigtableMessagesRow,
  BigtableSignedVAAsResultRow,
  BigtableSignedVAAsRow,
  BigtableVAAsByTxHashRow,
  VaasByBlock,
} from './types';
import {
  makeMessageId,
  makeVAAsByTxHashRowKey,
  makeSignedVAAsRowKey,
  parseMessageId,
} from './utils';
import { getSignedVAA } from '../utils/getSignedVAA';
import { PubSub } from '@google-cloud/pubsub';

const WATCH_MISSING_TIMEOUT = 5 * 60 * 1000;

export class BigtableDatabase extends Database {
  msgTableId: string;
  signedVAAsTableId: string;
  vaasByTxHashTableId: string;
  instanceId: string;
  bigtable: Bigtable;
  firestoreDb: FirebaseFirestore.Firestore;
  latestCollectionName: string;
  pubsubSignedVAATopic: string;
  pubsub: PubSub;
  constructor() {
    super();
    this.msgTableId = assertEnvironmentVariable('BIGTABLE_TABLE_ID');
    this.signedVAAsTableId = assertEnvironmentVariable('BIGTABLE_SIGNED_VAAS_TABLE_ID');
    this.vaasByTxHashTableId = assertEnvironmentVariable('BIGTABLE_VAAS_BY_TX_HASH_TABLE_ID');
    this.instanceId = assertEnvironmentVariable('BIGTABLE_INSTANCE_ID');
    this.latestCollectionName = assertEnvironmentVariable('FIRESTORE_LATEST_COLLECTION');
    this.pubsubSignedVAATopic = assertEnvironmentVariable('PUBSUB_SIGNED_VAA_TOPIC');
    try {
      this.bigtable = new Bigtable();
      const serviceAccount = require(assertEnvironmentVariable('FIRESTORE_ACCOUNT_KEY_PATH'));
      initializeApp({
        credential: cert(serviceAccount),
      });
      this.firestoreDb = getFirestore();
      this.pubsub = new PubSub();
    } catch (e) {
      throw new Error('Could not load bigtable db');
    }
  }

  async getLastBlockByChain(chain: ChainName): Promise<string | null> {
    const chainId = coalesceChainId(chain);
    const lastObservedBlock = this.firestoreDb
      .collection(this.latestCollectionName)
      .doc(chainId.toString());
    const lastObservedBlockByChain = await lastObservedBlock.get();
    const blockKeyData = lastObservedBlockByChain.data();
    const lastBlockKey = blockKeyData?.lastBlockKey;
    if (lastBlockKey) {
      this.logger.info(`for chain=${chain}, found most recent firestore block=${lastBlockKey}`);
      const tokens = lastBlockKey.split('/');
      return chain === 'aptos' ? tokens.at(-1) : tokens[0];
    }
    return null;
  }

  async storeLatestBlock(chain: ChainName, lastBlockKey: string): Promise<void> {
    if (this.firestoreDb === undefined) {
      this.logger.error('no firestore db set');
      return;
    }
    const chainId = coalesceChainId(chain);
    this.logger.info(`storing last block=${lastBlockKey} for chain=${chainId}`);
    const lastObservedBlock = this.firestoreDb
      .collection(this.latestCollectionName)
      .doc(`${chainId.toString()}`);
    await lastObservedBlock.set({ lastBlockKey });
  }

  async storeVaasByBlock(
    chain: ChainName,
    vaasByBlock: VaasByBlock,
    updateLatestBlock: boolean = true
  ): Promise<void> {
    if (this.bigtable === undefined) {
      this.logger.warn('no bigtable instance set');
      return;
    }
    const chainId = coalesceChainId(chain);
    const filteredBlocks = BigtableDatabase.filterEmptyBlocks(vaasByBlock);
    const instance = this.bigtable.instance(this.instanceId);
    const table = instance.table(this.msgTableId);
    const vaasByTxHashTable = instance.table(this.vaasByTxHashTableId);
    const rowsToInsert: BigtableMessagesRow[] = [];
    const vaasByTxHash: { [key: string]: string[] } = {};
    Object.keys(filteredBlocks).forEach((blockKey) => {
      const [block, timestamp] = blockKey.split('/');
      filteredBlocks[blockKey].forEach((msgKey) => {
        const [txHash, vaaKey] = msgKey.split(':');
        const [, emitter, seq] = vaaKey.split('/');
        rowsToInsert.push({
          key: makeMessageId(chainId, block, emitter, seq),
          data: {
            info: {
              timestamp: {
                value: timestamp,
                // write 0 timestamp to only keep 1 cell each
                // https://cloud.google.com/bigtable/docs/gc-latest-value
                timestamp: '0',
              },
              txHash: {
                value: txHash,
                timestamp: '0',
              },
              hasSignedVaa: {
                value: 0,
                timestamp: '0',
              },
            },
          },
        });
        const txHashRowKey = makeVAAsByTxHashRowKey(txHash, chainId);
        const vaaRowKey = makeSignedVAAsRowKey(chainId, emitter, seq);
        vaasByTxHash[txHashRowKey] = [...(vaasByTxHash[txHashRowKey] || []), vaaRowKey];
      });
    });
    const txHashRowsToInsert = Object.entries(vaasByTxHash).map<BigtableVAAsByTxHashRow>(
      ([txHashRowKey, vaaRowKeys]) => ({
        key: txHashRowKey,
        data: {
          info: {
            vaaKeys: { value: JSON.stringify(vaaRowKeys), timestamp: '0' },
          },
        },
      })
    );
    await Promise.all([table.insert(rowsToInsert), vaasByTxHashTable.insert(txHashRowsToInsert)]);

    if (updateLatestBlock) {
      // store latest vaasByBlock to firestore
      const blockKeys = Object.keys(vaasByBlock).sort(
        (bk1, bk2) => Number(bk1.split('/')[0]) - Number(bk2.split('/')[0])
      );
      if (blockKeys.length) {
        const lastBlockKey = blockKeys[blockKeys.length - 1];
        this.logger.info(`for chain=${chain}, storing last bigtable block=${lastBlockKey}`);
        await this.storeLatestBlock(chain, lastBlockKey);
      }
    }
  }

  async updateMessageStatuses(messageKeys: string[], value: number = 1): Promise<void> {
    const instance = this.bigtable.instance(this.instanceId);
    const table = instance.table(this.msgTableId);
    const chunkedMessageKeys = chunkArray(messageKeys, 1000);
    for (const chunk of chunkedMessageKeys) {
      const rowsToInsert: BigtableMessagesRow[] = chunk.map((id) => ({
        key: id,
        data: {
          info: {
            hasSignedVaa: {
              value: value,
              timestamp: '0',
            },
          },
        },
      }));
      // console.log(rowsToInsert[0].data.info)
      await table.insert(rowsToInsert);
    }
  }

  async fetchMissingVaaMessages(): Promise<BigtableMessagesResultRow[]> {
    const instance = this.bigtable.instance(this.instanceId);
    const messageTable = instance.table(this.msgTableId);
    // TODO: how to filter to only messages with hasSignedVaa === 0
    const observedMessages = (await messageTable.getRows())[0] as BigtableMessagesResultRow[];
    const missingVaaMessages = observedMessages.filter(
      (x) => x.data.info.hasSignedVaa?.[0].value === 0
    );
    return missingVaaMessages;
  }

  async watchMissing(): Promise<void> {
    const instance = this.bigtable.instance(this.instanceId);
    const signedVAAsTable = instance.table(this.signedVAAsTableId);
    while (true) {
      try {
        // this array first stores all of the messages which are missing VAAs
        // messages which we find VAAs for are then pruned from the array
        // lastly we try to fetch VAAs for the messages in the pruned array from the guardians
        const missingVaaMessages = await this.fetchMissingVaaMessages();
        const total = missingVaaMessages.length;
        this.logger.info(`locating ${total} messages with hasSignedVAA === 0`);
        let found = 0;
        const chunkedVAAIds = chunkArray(
          missingVaaMessages.map((observedMessage) => {
            const { chain, emitter, sequence } = parseMessageId(observedMessage.id);
            return makeSignedVAAsRowKey(chain, emitter, sequence.toString());
          }),
          1000
        );
        let chunkNum = 0;
        const foundKeys: string[] = [];
        for (const chunk of chunkedVAAIds) {
          this.logger.info(`processing chunk ${++chunkNum} of ${chunkedVAAIds.length}`);
          const vaaRows = (
            await signedVAAsTable.getRows({
              keys: chunk,
              decode: false,
            })
          )[0] as BigtableSignedVAAsResultRow[];
          for (const row of vaaRows) {
            try {
              const vaaBytes = row.data.info.bytes[0].value;
              const parsed = parseVaa(vaaBytes);
              const matchingIndex = missingVaaMessages.findIndex((observedMessage) => {
                const { chain, emitter, sequence } = parseMessageId(observedMessage.id);
                if (
                  parsed.emitterChain === chain &&
                  parsed.emitterAddress.toString('hex') === emitter &&
                  parsed.sequence === sequence
                ) {
                  return true;
                }
              });
              if (matchingIndex !== -1) {
                found++;
                // remove matches to keep array lean
                // messages with missing VAAs will be kept in the array
                const [matching] = missingVaaMessages.splice(matchingIndex, 1);
                foundKeys.push(matching.id);
              }
            } catch (e) {}
          }
        }
        this.logger.info(`processed ${total} messages, found ${found}, missing ${total - found}`);
        this.updateMessageStatuses(foundKeys);
        // attempt to fetch VAAs missing from messages from the guardians and store them
        // this is useful for cases where the VAA doesn't exist in the `signedVAAsTable` (perhaps due to an outage) but is available
        const missingSignedVAARows: BigtableSignedVAAsRow[] = [];
        for (const msg of missingVaaMessages) {
          const { chain, emitter, sequence } = parseMessageId(msg.id);
          const seq = sequence.toString();
          const vaaBytes = await getSignedVAA(chain, emitter, seq);
          if (vaaBytes) {
            const key = makeSignedVAAsRowKey(chain, emitter, seq);
            missingSignedVAARows.push({
              key,
              data: {
                info: {
                  bytes: { value: vaaBytes, timestamp: '0' },
                },
              },
            });
          }
        }
        this.storeSignedVAAs(missingSignedVAARows);
        this.publishSignedVAAs(missingSignedVAARows.map((r) => r.key));
        // TODO: add slack message alerts
      } catch (e) {
        this.logger.error(e);
      }
      await sleep(WATCH_MISSING_TIMEOUT);
    }
  }

  async storeSignedVAAs(rows: BigtableSignedVAAsRow[]): Promise<void> {
    const instance = this.bigtable.instance(this.instanceId);
    const table = instance.table(this.signedVAAsTableId);
    const chunks = chunkArray(rows, 1000);
    for (const chunk of chunks) {
      await table.insert(chunk);
      this.logger.info(`wrote ${chunk.length} signed VAAs to the ${this.signedVAAsTableId} table`);
    }
  }

  async publishSignedVAAs(keys: string[]): Promise<void> {
    if (keys.length === 0) {
      return;
    }
    try {
      const topic = this.pubsub.topic(this.pubsubSignedVAATopic);
      if (!(await topic.exists())) {
        this.logger.error(`pubsub topic doesn't exist: ${this.publishSignedVAAs}`);
        return;
      }
      for (const key of keys) {
        await topic.publishMessage({ data: Buffer.from(key) });
      }
      this.logger.info(
        `published ${keys.length} signed VAAs to pubsub topic: ${this.pubsubSignedVAATopic}`
      );
    } catch (e) {
      this.logger.error(`pubsub error - ${e}`);
    }
  }
}
