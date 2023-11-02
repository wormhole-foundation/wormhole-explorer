import { CONTRACTS } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { describe, expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common/consts';
import { NETWORK_RPCS_BY_CHAIN } from '../../consts';
import { getNearProvider, getTransactionsByAccountId, NEAR_ARCHIVE_RPC } from '../../utils/near';
import { getMessagesFromBlockResults, NearWatcher } from '../NearWatcher';

jest.setTimeout(60000);

const INITIAL_NEAR_BLOCK = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.near ?? 0);

test('getFinalizedBlockNumber', async () => {
  const watcher = new NearWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(INITIAL_NEAR_BLOCK);
});

test('getMessagesForBlocks', async () => {
  // requests that are too old for rpc node should error, be caught, and return an empty object
  const watcher = new NearWatcher();
  const messages = await watcher.getMessagesForBlocks(INITIAL_NEAR_BLOCK, INITIAL_NEAR_BLOCK);
  expect(Object.keys(messages).length).toEqual(0);
});

describe('getNearProvider', () => {
  test('with normal RPC', async () => {
    const provider = await getNearProvider(NETWORK_RPCS_BY_CHAIN['near']!);
    // grab last block from core contract
    expect(await provider.block({ finality: 'final' })).toBeTruthy();
  });

  test('with archive RPC', async () => {
    const provider = await getNearProvider(NEAR_ARCHIVE_RPC);
    // grab first block with activity from core contract
    expect(
      await provider.block({ blockId: 'Asie8hpJFKaipvw8jh1wPfBwwbjP6JUfsQdCuQvwr3Sz' })
    ).toBeTruthy();
  });
});

test('getTransactionsByAccountId', async () => {
  let transactions = await getTransactionsByAccountId(
    CONTRACTS.MAINNET.near.core,
    10,
    '1669732480649090392'
  );
  expect(transactions.length).toEqual(10);
  expect(transactions[0].hash).toEqual('7jDrPnvErjbi3EHbQBcKT9wtiUPo77J9tpxXjE3KHcUp');

  // test custom timestamp, filtering out non function call actions, and querying last page
  transactions = await getTransactionsByAccountId(
    CONTRACTS.MAINNET.near.core,
    15,
    '1661429914932000000'
  );
  expect(transactions.length).toEqual(2);
  expect(transactions[0].hash).toEqual('3VivTHp1W5ErWgsASUQvW1qwoTCsxYeke4498apDJsss');
});

describe('getMessagesFromBlockResults', () => {
  test('with Provider', async () => {
    const watcher = new NearWatcher();
    const provider = await watcher.getProvider();
    const messages = getMessagesFromBlockResults(provider, [
      await provider.block({ finality: 'final' }),
    ]);
    expect(messages).toBeTruthy();
  });

  test.skip('with ArchiveProvider', async () => {
    const provider = await getNearProvider(NEAR_ARCHIVE_RPC);
    const messages = await getMessagesFromBlockResults(provider, [
      await provider.block({ blockId: 'Bzjemj99zxe1h8kVp8H2hwVifmbQL8HT34LyPHzEK5qp' }),
      await provider.block({ blockId: '4SHFxSo8DdP8DhMauS5iFqfmdLwLET3W3e8Lg9PFvBSn' }),
      await provider.block({ blockId: 'GtQYaYMhrDHgLJJTroUaUzSR24E29twewpkqyudrCyVN' }),
    ]);
    expect(messages).toMatchObject({
      '72777217/2022-08-25T18:42:26.121Z': [],
      '74616314/2022-09-21T18:48:05.392Z': [
        'SYRSkE8pBWWLPZWJtHEGN5Hk7SPZ7kHgf4D1Q4viRcz:15/148410499d3fcda4dcfd68a1ebfcdddda16ab28326448d4aae4d2f0465cdfcb7/233',
      ],
      '74714181/2022-09-23T05:15:53.722Z': [
        '2xh2rLR3ehjRRjU1BbuHEhU6FbXiKp5rZ88niyKC6MBs:15/148410499d3fcda4dcfd68a1ebfcdddda16ab28326448d4aae4d2f0465cdfcb7/237',
      ],
    });

    // validate keys
    const watcher = new NearWatcher();
    const blockKey = Object.keys(messages).at(-1)!;
    expect(watcher.isValidBlockKey(blockKey)).toBe(true);
    expect(watcher.isValidVaaKey(messages[blockKey][0])).toBe(true);
  });
});
