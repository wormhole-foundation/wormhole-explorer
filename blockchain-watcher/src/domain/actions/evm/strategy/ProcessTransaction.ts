export interface ProcessTransaction<T> {
  apply(): boolean;
  execute(): T[];
}
