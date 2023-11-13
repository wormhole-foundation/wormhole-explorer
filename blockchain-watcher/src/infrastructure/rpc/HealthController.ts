import { StatRepository } from "../../domain/repositories";

export class HealthController {
  private readonly statsRepo: StatRepository;

  constructor(statsRepo: StatRepository) {
    this.statsRepo = statsRepo;
  }

  metrics = async () => {
    return this.statsRepo.report();
  };
}
