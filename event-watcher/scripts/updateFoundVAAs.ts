import * as dotenv from 'dotenv';
dotenv.config();
import { BigtableDatabase } from '../src/databases/BigtableDatabase';

// This script takes the output of fetchMissingVAAs and writes the found records back to the VAA big table

(async () => {
  const found: { [id: string]: string } = require('../found.json');
  const bt = new BigtableDatabase();
  if (!bt.bigtable) {
    throw new Error('bigtable is undefined');
  }
  try {
    bt.storeSignedVAAs(
      Object.entries(found).map(([id, vaaBytes]) => {
        const vaa = Buffer.from(vaaBytes, 'hex');
        return {
          key: id,
          data: {
            info: {
              bytes: { value: vaa, timestamp: '0' },
            },
          },
        };
      })
    );
  } catch (e) {
    console.error(e);
  }
})();
