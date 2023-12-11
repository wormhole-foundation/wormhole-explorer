import { beforeAll, describe, it, afterAll, expect } from "@jest/globals";
import pg from "pg";
import { PostgreSqlContainer, StartedPostgreSqlContainer } from "@testcontainers/postgresql";
import { PostgresJobExecutionRepository } from "../../../src/infrastructure/repositories/jobs/execution/PostgresJobExecutionRepository";

describe("PostgresJobExecutionRepository", () => {
  let postgres: StartedPostgreSqlContainer;

  beforeAll(async () => {
    postgres = await new PostgreSqlContainer("postgres:15").withDatabase("test-exec").start();
  }, 15_000);

  afterAll(async () => {
    await postgres.stop();
  });

  it("should return an execution if no other present", async () => {
    const aRepo = await givenARepo(await givenAClient(postgres), true);
    const jobExec = await aRepo.start(givenJobDefinition());

    expect(jobExec.status).toEqual("running");
    expect(jobExec.startedAt).toBeDefined();

    await cleanUp(aRepo);
  }, 10_000);

  it("should not return an execution if other present", async () => {
    const job = givenJobDefinition();
    const aRepo = await givenARepo(await givenAClient(postgres), true);
    const anotherRepo = await givenARepo(await givenAClient(postgres), true);

    const firstStart = await aRepo.start(job);
    expect(firstStart.status).toEqual("running");
    expect(firstStart.startedAt).toBeDefined();

    await expect(anotherRepo.start(job)).rejects.toThrow();

    await aRepo.stop(firstStart);
    await cleanUp(aRepo, anotherRepo);
  }, 60_000);
});

const givenJobDefinition = () => {
  return {
    id: "job1",
    name: "job1",
    source: { action: "action", config: {} },
    chain: "ethereum",
    network: "mainnet",
    handlers: [],
  };
};

const givenARepo = async (client: pg.Pool, init: boolean = false) => {
  return new PostgresJobExecutionRepository(client);
};

const givenAClient = async (postgres: StartedPostgreSqlContainer) => {
  const client = new pg.Pool({
    connectionString: postgres.getConnectionUri(),
    connectionTimeoutMillis: 10000,
    query_timeout: 10000,
    max: 1,
    min: 0,
  });
  return client;
};

const cleanUp = async (...repos: PostgresJobExecutionRepository[]) => {
  await Promise.all(repos.map((repo) => repo.close()));
};
