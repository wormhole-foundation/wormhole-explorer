export interface Source {
  get(): Promise<any[]>;
  hasNext(): Promise<boolean>;
}
