import { expect, jest, test } from '@jest/globals';
import { INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN } from '../../common/consts';
import { SolanaWatcher } from '../SolanaWatcher';

jest.setTimeout(60000);

const INITIAL_SOLANA_BLOCK = Number(INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN.solana ?? 0);

test('getFinalizedBlockNumber', async () => {
  const watcher = new SolanaWatcher();
  const blockNumber = await watcher.getFinalizedBlockNumber();
  expect(blockNumber).toBeGreaterThan(INITIAL_SOLANA_BLOCK);
});

test('getMessagesForBlocks - single block', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(170799004, 170799004);
  expect(Object.keys(messages).length).toBe(1);
  expect(messages).toMatchObject({
    '170799004/2023-01-04T16:43:43.000Z': [
      '3zWJevhFB5XqUCdDmqoRLQUMgiNBmFZLaE5rZpSexH47Mx2268eimrj2FY23Z1mq1WXsRRkyhmMcsguXcSw7Rnh1:1/ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5/262100',
    ],
  });

  // validate keys
  expect(watcher.isValidBlockKey(Object.keys(messages)[0])).toBe(true);
  expect(watcher.isValidVaaKey(Object.values(messages).flat()[0])).toBe(true);
});

// temporary skip due to SolanaJSONRPCError: failed to get confirmed block: Block 171774030 cleaned up, does not exist on node. First available block: 176896202
test('getMessagesForBlocks - fromSlot is skipped slot', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(171774030, 171774032); // 171774024 - 171774031 are skipped
  expect(Object.keys(messages).length).toBe(1);
  expect(messages).toMatchObject({ '171774032/2023-01-10T13:36:38.000Z': [] });
});

test('getMessagesForBlocks - toSlot is skipped slot', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(171774023, 171774025);
  expect(messages).toMatchObject({ '171774023/2023-01-10T13:36:34.000Z': [] });
});

test('getMessagesForBlocks - empty block', async () => {
  // Even if there are no messages, last block should still be returned
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(170979766, 170979766);
  expect(Object.keys(messages).length).toBe(1);
  expect(messages).toMatchObject({ '170979766/2023-01-05T18:40:24.000Z': [] });
});

// temporary skip due to SolanaJSONRPCError: failed to get confirmed block: Block 174108865 cleaned up, does not exist on node. First available block: 176892532
test('getMessagesForBlocks - block with no transactions', async () => {
  const watcher = new SolanaWatcher();
  expect(watcher.getMessagesForBlocks(174108861, 174108861)).rejects.toThrowError(
    'solana: invalid block range'
  );

  let messages = await watcher.getMessagesForBlocks(174108661, 174108861);
  expect(Object.keys(messages).length).toBe(1);
  expect(Object.values(messages).flat().length).toBe(0);

  messages = await watcher.getMessagesForBlocks(174108863, 174109061);
  expect(Object.keys(messages).length).toBe(1);
  expect(Object.values(messages).flat().length).toBe(0);
});

test('getMessagesForBlocks - multiple blocks', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(171050470, 171050474);
  expect(Object.keys(messages).length).toBe(2);
  expect(Object.values(messages).flat().length).toBe(2);
});

test('getMessagesForBlocks - multiple blocks, last block empty', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(170823000, 170825000);
  expect(Object.keys(messages).length).toBe(3);
  expect(Object.values(messages).flat().length).toBe(2); // 2 messages, last block has no message
});

test('getMessagesForBlocks - multiple blocks containing more than `getSignaturesLimit` WH transactions', async () => {
  const watcher = new SolanaWatcher();
  watcher.getSignaturesLimit = 10;
  const messages = await watcher.getMessagesForBlocks(171582367, 171583452);
  expect(Object.keys(messages).length).toBe(3);
  expect(Object.values(messages).flat().length).toBe(3);
});

test('getMessagesForBlocks - multiple calls', async () => {
  const watcher = new SolanaWatcher();
  const messages1 = await watcher.getMessagesForBlocks(171773021, 171773211);
  const messages2 = await watcher.getMessagesForBlocks(171773212, 171773250);
  const messages3 = await watcher.getMessagesForBlocks(171773251, 171773500);
  const allMessageKeys = [
    ...Object.keys(messages1),
    ...Object.keys(messages2),
    ...Object.keys(messages3),
  ];
  const uniqueMessageKeys = [...new Set(allMessageKeys)];
  expect(allMessageKeys.length).toBe(uniqueMessageKeys.length); // assert no duplicate keys
});

test('getMessagesForBlocks - handle failed transactions', async () => {
  const watcher = new SolanaWatcher();
  const messages = await watcher.getMessagesForBlocks(94401321, 94501321);
  expect(Object.keys(messages).length).toBe(6);
  expect(Object.values(messages).flat().length).toBe(5);
  expect(
    Object.values(messages)
      .flat()
      .map((m) => m.split('/')[2])
      .join(',')
  ).toBe('4,3,2,1,0');
});
