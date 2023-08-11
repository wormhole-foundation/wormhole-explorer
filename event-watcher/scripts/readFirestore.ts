import * as dotenv from 'dotenv';
dotenv.config();
import { BigtableDatabase } from '../src/databases/BigtableDatabase';

// This script provides a summary of the latest block db

(async () => {
  const bt = new BigtableDatabase();
  try {
    const collectionRef = bt.firestoreDb.collection(bt.latestCollectionName);
    const snapshot = await collectionRef.get();
    snapshot.docs
      .sort((a, b) => Number(a.id) - Number(b.id))
      .forEach((doc) => {
        const [block, timestamp] = doc.data().lastBlockKey.split('/');
        console.log(doc.id.padEnd(2), '=>', timestamp, block.padStart(10));
      });
  } catch (e) {
    console.error(e);
  }
})();
