import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '@wormhole-foundation/wormhole-monitor-common/dist/consts';
import { SuiWatcher } from '../SuiWatcher';

jest.setTimeout(60000);

const INITAL_SEQUENCE_NUMBER = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.sui ?? 1581000);

test('getFinalizedSequenceNumber', async () => {
  const watcher = new SuiWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  console.log('Received blockNumber:', blockNumber);
  expect(blockNumber).toBeGreaterThan(INITAL_SEQUENCE_NUMBER);
});

// This test will fail as time goes on because getMessagesForBlocks() grabs the latest and
// works backwards.  This will cause a 429 until we clear that up.
test.skip('getMessagesForBlocks', async () => {
  const watcher = new SuiWatcher();
  const messages = await watcher.getMessagesForBlocks(1581997, 1581997);
  console.log(messages);
  const entries = Object.entries(messages);
  expect(entries.length).toEqual(46);
  expect(entries.filter(([block, vaas]) => vaas.length === 0).length).toEqual(1);
  expect(entries.filter(([block, vaas]) => vaas.length === 1).length).toEqual(40);
  expect(entries.filter(([block, vaas]) => vaas.length === 2).length).toEqual(5);
  expect(messages['1584976/2023-05-03T17:15:00.000Z']).toBeDefined();
  expect(messages['1584976/2023-05-03T17:15:00.000Z'].length).toEqual(1);
  expect(messages['1584976/2023-05-03T17:15:00.000Z'][0]).toEqual(
    'HydDe4yNBBu98ak46fPdw7qCZ4x7h8DsYdMfeWEBf5ge:21/ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5/187'
  );
});
