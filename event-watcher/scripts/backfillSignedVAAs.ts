import * as dotenv from 'dotenv';
dotenv.config();
import { createReadStream } from 'fs';
import { createInterface } from 'readline';
import { assertEnvironmentVariable } from '@wormhole-foundation/wormhole-monitor-common/src/utils';
import { BigtableDatabase } from '../src/databases/BigtableDatabase';
import ora from 'ora';
import { makeSignedVAAsRowKey } from '../src/databases/utils';
import { ChainId } from '@certusone/wormhole-sdk';

// This script writes all VAAs from a csv file compatible with the guardian `sign-existing-vaas-csv` admin command to bigtable

const CHUNK_SIZE = 10000;

interface SignedVAAsRow {
  key: string;
  data: {
    info: {
      bytes: { value: Buffer; timestamp: '0' };
    };
  };
}

(async () => {
  try {
    const vaaCsvFilename = assertEnvironmentVariable('VAA_CSV_FILE');

    const bt = new BigtableDatabase();
    if (!bt.bigtable) {
      throw new Error('bigtable is undefined');
    }
    const vaaTableId = assertEnvironmentVariable('BIGTABLE_SIGNED_VAAS_TABLE_ID');
    const instance = bt.bigtable.instance(bt.instanceId);
    const vaaTable = instance.table(vaaTableId);

    const fileStream = createReadStream(vaaCsvFilename, { encoding: 'utf8' });

    const rl = createInterface({
      input: fileStream,
      crlfDelay: Infinity,
    });
    // Note: we use the crlfDelay option to recognize all instances of CR LF
    // ('\r\n') in input.txt as a single line break.

    let rows: SignedVAAsRow[] = [];
    let numWritten = 0;
    let log = ora('Writing VAAs to bigtable...').start();
    for await (const line of rl) {
      const split = line.split(',');
      const key = split[0];
      const vaa = split[1];
      const splitKey = key.split(':');
      const chain = Number(splitKey[0]);
      const emitter = splitKey[1];
      const sequence = splitKey[2];
      const rowKey = makeSignedVAAsRowKey(chain as ChainId, emitter, sequence);
      rows.push({
        key: rowKey,
        data: {
          info: {
            bytes: { value: Buffer.from(vaa, 'hex'), timestamp: '0' },
          },
        },
      });
      if (rows.length == CHUNK_SIZE) {
        await vaaTable.insert(rows);
        numWritten += rows.length;
        log.text = `Wrote ${numWritten} VAAs`;
        rows = [];
      }
    }
    if (rows.length > 0) {
      await vaaTable.insert(rows);
      numWritten += rows.length;
    }
    log.succeed(`Wrote ${numWritten} VAAs`);
  } catch (e) {
    console.error(e);
  }
})();
