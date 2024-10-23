import { UnhandledError } from "../../infrastructure/errors/UnhandledError";
import { StatRepository } from "../repositories";
import winston from "winston";

export abstract class RunRegistry {
  private statRepo?: StatRepository;

  protected abstract logger: winston.Logger;
  protected abstract execute(): Promise<void>;
  protected abstract report(): void;

  constructor(statsRepo: StatRepository) {
    this.statRepo = statsRepo;
  }

  public async run(): Promise<void> {
    console.log("RunRegistry");
  }
}
