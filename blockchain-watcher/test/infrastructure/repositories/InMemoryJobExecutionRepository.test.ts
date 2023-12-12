import { describe, beforeEach, it, expect } from "@jest/globals";
import { InMemoryJobExecutionRepository } from "../../../src/infrastructure/repositories/jobs/execution/InMemoryJobExecutionRepository";
import { JobDefinition, JobExecution } from "../../../src/domain/entities";

describe("InMemoryJobExecutionRepository", () => {
  let repository: InMemoryJobExecutionRepository;

  beforeEach(() => {
    repository = new InMemoryJobExecutionRepository();
  });

  describe("start", () => {
    it("should start a job execution and return the execution object", async () => {
      const job = {
        id: "job1",
        name: "job1",
      } as JobDefinition;

      const execution: JobExecution = await repository.start(job);

      expect(execution).toBeDefined();
      expect(execution.job).toEqual(job);
    });
  });

  describe("stop", () => {
    it("should stop a job execution and return the updated execution object", async () => {
      const job = {
        id: "job1",
        name: "job1",
      } as JobDefinition;
      const execution: JobExecution = await repository.start(job);

      const updatedExecution: JobExecution = await repository.stop(execution);

      expect(updatedExecution).toBeDefined();
      expect(updatedExecution.status).toEqual("stopped");
    });

    it("should stop a job execution with an error and return the updated execution object", async () => {
      const job = {
        id: "job1",
        name: "job1",
      } as JobDefinition;
      const execution: JobExecution = await repository.start(job);
      const error: Error = new Error("Something went wrong");

      const updatedExecution: JobExecution = await repository.stop(execution, error);

      expect(updatedExecution).toBeDefined();
      expect(updatedExecution.status).toEqual("stopped");
      expect(updatedExecution.error).toEqual(error);
    });
  });
});
