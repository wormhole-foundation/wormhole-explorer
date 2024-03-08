import { Block } from "../../../domain/actions/aptos/PollAptos";

export function divideIntoBatches<T>(set: Set<T>, batchSize = 10): Set<T>[] {
  const batches: Set<T>[] = [];
  let batch: any[] = [];

  set.forEach((item) => {
    batch.push(item);
    if (batch.length === batchSize) {
      batches.push(new Set(batch));
      batch = [];
    }
  });

  if (batch.length > 0) {
    batches.push(new Set(batch));
  }
  return batches;
}

export function createBatches(range: Block | undefined): number[] {
  let batchSize = 100;
  let total = 1;

  if (range && range.toBlock) {
    batchSize = range.toBlock < batchSize ? range.toBlock : batchSize;
    total = range.toBlock ?? total;
  }

  const numBatches = Math.ceil(total / batchSize);
  const batches: number[] = [];

  for (let i = 0; i < numBatches; i++) {
    batches.push(batchSize);
  }

  return batches;
}
