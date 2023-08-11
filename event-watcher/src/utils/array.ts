/**
 * Wrap input in array if it isn't already an array.
 * @param input
 */
export function toArray<T>(input: T | T[]): T[] {
  if (input == null) return []; // Catch undefined and null values
  return input instanceof Array ? input : [input];
}
