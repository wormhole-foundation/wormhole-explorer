export class Fallible<R, E> {
  private value?: R;
  private error?: E;

  constructor(value?: R, error?: E) {
    this.value = value;
    this.error = error;
  }

  public getValue(): R {
    if (!this.value) {
      throw this.error ? this.error : new Error("No value and no error present");
    }
    return this.value;
  }

  public getError(): E {
    if (!this.error) {
      throw new Error("No error present");
    }
    return this.error;
  }

  public isOk(): boolean {
    return !this.error && this.value !== undefined;
  }

  public static ok<R, E>(value: R): Fallible<R, E> {
    return new Fallible<R, E>(value);
  }

  public static error<R, E>(error: E): Fallible<R, E> {
    return new Fallible<R, E>(undefined, error);
  }
}
