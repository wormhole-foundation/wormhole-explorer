import { StatRepository } from "../../domain/repositories";

export class UnhandledError {
  private readonly statRepo: StatRepository;
  private readonly error: any;
  private readonly id: string;

  constructor(statRepo: StatRepository, error: any, id: string) {
    this.statRepo = statRepo;
    this.error = error;
    this.id = id;
  }

  private errorTypes = [
    {
      type: "noHealthyProviders",
      messages: ["No healthy providers"],
      errorMessage: "No healthy providers",
    },
    { type: "rateLimit", messages: ["Ratelimited", "rate limited"], errorMessage: "Rate limited" },
  ];

  public validateError(): never | void {
    for (const { type, messages, errorMessage } of this.errorTypes) {
      if (messages.some((msg) => this.error.toString().includes(msg))) {
        this.throwError(type, errorMessage);
      }
    }
  }

  private throwError(type: string, message: string): never {
    this.statRepo.count(`job_unhandled_errors_total`, { id: this.id, status: "error" });
    throw new Error(`[run][${type}] ${message}, job: ${this.id}`);
  }
}
