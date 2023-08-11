import * as dotenv from 'dotenv';
dotenv.config();
import { ChainId, CHAINS, coalesceChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { MAX_UINT_64, padUint16 } from '../src/common';
import { BigtableDatabase } from '../src/databases/BigtableDatabase';

// This script provides a summary of the message db

(async () => {
  const bt = new BigtableDatabase();
  if (!bt.bigtable) {
    throw new Error('bigtable is undefined');
  }
  const mainnetInstance = bt.bigtable.instance(bt.instanceId);
  const messageTable = mainnetInstance.table(bt.msgTableId);
  try {
    const chain: ChainId = 22;
    const prefix = `${padUint16(chain.toString())}/`;
    const observedMessages = await messageTable.getRows({ prefix, limit: 100 });
    console.log(
      coalesceChainName(chain).padEnd(12),
      observedMessages[0].length.toString().padStart(6)
    );
    if (observedMessages[0][0]) {
      console.log('   id           ', observedMessages[0][0]?.id);
      console.log('   chain        ', parseInt(observedMessages[0][0]?.id.split('/')[0]));
      console.log(
        '   block        ',
        BigInt(MAX_UINT_64) - BigInt(observedMessages[0][0]?.id.split('/')[1] || 0)
      );
      console.log('   emitter      ', observedMessages[0][0]?.id.split('/')[2]);
      console.log('   seq          ', parseInt(observedMessages[0][0]?.id.split('/')[3]));
      console.log('   timestamp    ', observedMessages[0][0]?.data.info.timestamp[0].value);
      console.log('   txHash       ', observedMessages[0][0]?.data.info.txHash[0].value);
      console.log('   hasSignedVaa ', observedMessages[0][0]?.data.info.hasSignedVaa[0].value);
    }
    if (observedMessages[0][1]) {
      console.log('   id           ', observedMessages[0][1]?.id);
      console.log('   chain        ', parseInt(observedMessages[0][1]?.id.split('/')[0]));
      console.log(
        '   block        ',
        BigInt(MAX_UINT_64) - BigInt(observedMessages[0][1]?.id.split('/')[1] || 0)
      );
      console.log('   emitter      ', observedMessages[0][1]?.id.split('/')[2]);
      console.log('   seq          ', parseInt(observedMessages[0][1]?.id.split('/')[3]));
      console.log('   timestamp    ', observedMessages[0][1]?.data.info.timestamp[0].value);
      console.log('   txHash       ', observedMessages[0][1]?.data.info.txHash[0].value);
      console.log('   hasSignedVaa ', observedMessages[0][1]?.data.info.hasSignedVaa[0].value);
    }
  } catch (e) {
    console.error(e);
  }
})();
