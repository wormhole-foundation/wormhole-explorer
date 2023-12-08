import { beforeAll, describe, it, expect, afterAll } from "@jest/globals";
import pg from "pg";
import { PostgreSqlContainer, StartedPostgreSqlContainer } from "@testcontainers/postgresql";
import { PostgresMetadataRepository } from "../../../src/infrastructure/repositories/jobs/metadata/PostgresMetadataRepository";

describe("PostgresMigrator", () => {
  let postgres: StartedPostgreSqlContainer;

  beforeAll(async () => {
    postgres = await new PostgreSqlContainer().withDatabase("test").start();
  }, 15_000);

  afterAll(async () => {
    await postgres.stop();
  });

  it("should init its structure successfully", async () => {
    const aRepo = givenARepo(await givenAClient(postgres));
    const aDifferentRepo = givenARepo(await givenAClient(postgres));

    await Promise.all([aRepo.init(), aDifferentRepo.init()]);

    await cleanUp(aRepo);
    await cleanUp(aDifferentRepo);
  }, 10_000);

  it("should save and get", async () => {
    const repo = givenARepo(await givenAClient(postgres));
    await repo.init();

    const data = { latestBlocks: { 1: 2, 3: 4 } };

    await repo.save("job1", data);
    const response = await repo.get("job1");
    expect(response).toEqual(data);

    await cleanUp(repo);
  }, 10_000);
});

const givenARepo = (client: pg.Pool) => {
  return new PostgresMetadataRepository(client);
};

const givenAClient = async (postgres: StartedPostgreSqlContainer) => {
  return new pg.Pool({
    connectionString: postgres.getConnectionUri(),
    connectionTimeoutMillis: 10000,
    query_timeout: 10000,
    max: 5,
  });
};

const cleanUp = async (repo: PostgresMetadataRepository) => {
  await repo.close();
};
