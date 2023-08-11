import { padUint16, sleep } from '../src/common';
import * as dotenv from 'dotenv';
dotenv.config();
import { BigtableDatabase } from '../src/databases/BigtableDatabase';
import { parseMessageId } from '../src/databases/utils';

// This script updates the getSignedVaa value for a given list of vaa rowkeys

export function parseVaaId(vaaRowKey: string) {
  let [chain, emitter, sequence] = vaaRowKey.split(':');
  // chain: number, emitter: string, sequence: bigint
  return [chain, emitter, sequence];
}

(async () => {
  const bt = new BigtableDatabase();
  if (!bt.bigtable) {
    throw new Error('bigtable is undefined');
  }
  const instance = bt.bigtable.instance(bt.instanceId);
  const messageTable = instance.table(bt.msgTableId);

  const rowKeysToUpdate: string[] = [
    '5:0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde:0000000000006840',
    '7:00000000000000000000000004952d522ff217f40b5ef3cbf659eca7b952a6c1:0000000000000002',
    '7:0000000000000000000000005848c791e09901b40a9ef749f2a6735b418d7564:0000000000006971',
    '15:148410499d3fcda4dcfd68a1ebfcdddda16ab28326448d4aae4d2f0465cdfcb7:0000000000000001',
  ];

  try {
    // STEP 1

    console.log(`processing ${rowKeysToUpdate.length} rowKeys`);

    for (const rowKey of rowKeysToUpdate) {
      let [chain, targetEmitter, targetSequence] = parseVaaId(rowKey);
      const formattedChainId = padUint16(chain);
      const [rowsByChain] = await messageTable.getRows({ prefix: formattedChainId });
      let messageRowKey = '';
      //filter to find sequence numbers:
      rowsByChain.forEach((row) => {
        const { chain, block, emitter, sequence } = parseMessageId(row.id);
        if (targetEmitter === emitter && BigInt(targetSequence) === sequence) {
          console.log(`found ${row.id} for rowKey=${rowKey}`);

          //update rowKey
          messageRowKey = row.id;
        }
      });
      if (messageRowKey !== '') {
        console.log(`updating ${messageRowKey} to value=${2}`);
        await bt.updateMessageStatuses([messageRowKey], 2);
      }
    }
  } catch (e) {
    console.error(e);
  }
})();
