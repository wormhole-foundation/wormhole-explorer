import * as dotenv from 'dotenv';
dotenv.config();

import { CHAIN_ID_SOLANA, coalesceChainId, coalesceChainName } from '@certusone/wormhole-sdk';
import { padUint16 } from '../src/common';
import { BigtableDatabase } from '../src/databases/BigtableDatabase';

// Script to delete all messages for the chain given by the CHAIN variable below

const CHAIN = CHAIN_ID_SOLANA;

(async () => {
  const bt = new BigtableDatabase();
  if (!bt.bigtable) {
    throw new Error('bigtable is undefined');
  }

  const instance = bt.bigtable.instance(bt.instanceId);
  const messageTable = instance.table(bt.msgTableId);
  await messageTable.deleteRows(`${padUint16(coalesceChainId(CHAIN).toString())}/`);
  console.log('Deleted all rows starting with', coalesceChainName(CHAIN));
})();
