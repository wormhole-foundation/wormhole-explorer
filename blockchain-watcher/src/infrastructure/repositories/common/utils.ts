export function divideIntoBatches(set: Set<string | bigint>, batchSize = 10) {
  const batches = [];
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
