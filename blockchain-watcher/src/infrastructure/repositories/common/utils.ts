import { Range } from "../../../domain/actions/aptos/PollAptos";

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
