import { MetadataRepository, StatRepository } from "../../src/domain/repositories";

export const mockStatsRepository = (): StatRepository => {
  return {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

export const mockMetadataRepository = <T>(metadata?: T): MetadataRepository<T> => {
  return {
    get: () => Promise.resolve(metadata),
    save: () => Promise.resolve(),
  };
};
