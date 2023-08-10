import * as dotenv from 'dotenv';
dotenv.config();
import { BigtableDatabase } from '../src/databases/BigtableDatabase';
import ora from 'ora';
import { BigtableVAAsByTxHashRow } from '../src/databases/types';
import {
  makeSignedVAAsRowKey,
  makeVAAsByTxHashRowKey,
  parseMessageId,
} from '../src/databases/utils';
import { chunkArray } from '../src/common';

const CHUNK_SIZE = 10000;

(async () => {
  try {
    const bt = new BigtableDatabase();
    if (!bt.bigtable) {
      throw new Error('bigtable is undefined');
    }
    const instance = bt.bigtable.instance(bt.instanceId);
    const messageTable = instance.table(bt.msgTableId);
    const vaasByTxHashTable = instance.table(bt.vaasByTxHashTableId);

    let log = ora(`Reading rows from ${bt.msgTableId}...`).start();
    const observedMessages = await messageTable.getRows(); // TODO: pagination
    const vaasByTxHash: { [key: string]: string[] } = {};
    for (const msg of observedMessages[0]) {
      const txHash = msg.data.info.txHash[0].value;
      const { chain, emitter, sequence } = parseMessageId(msg.id);
      const txHashRowKey = makeVAAsByTxHashRowKey(txHash, chain);
      const vaaRowKey = makeSignedVAAsRowKey(chain, emitter, sequence.toString());
      vaasByTxHash[txHashRowKey] = [...(vaasByTxHash[txHashRowKey] || []), vaaRowKey];
    }
    const rowsToInsert = Object.entries(vaasByTxHash).map<BigtableVAAsByTxHashRow>(
      ([txHashRowKey, vaaRowKeys]) => ({
        key: txHashRowKey,
        data: {
          info: {
            vaaKeys: { value: JSON.stringify(vaaRowKeys), timestamp: '0' },
          },
        },
      })
    );
    const rowChunks = chunkArray(rowsToInsert, CHUNK_SIZE);
    let numWritten = 0;
    for (const rowChunk of rowChunks) {
      await vaasByTxHashTable.insert(rowChunk);
      numWritten += rowChunk.length;
      log.text = `Wrote ${numWritten}/${rowsToInsert.length} rows to ${bt.vaasByTxHashTableId}`;
    }
    log.succeed(`Wrote ${numWritten} rows to ${bt.vaasByTxHashTableId}`);
  } catch (e) {
    console.error(e);
  }
})();
