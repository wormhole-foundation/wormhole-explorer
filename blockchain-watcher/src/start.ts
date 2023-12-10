import { configuration } from "./infrastructure/config";
import { RepositoriesBuilder } from "./infrastructure/repositories/RepositoriesBuilder";
import log from "./infrastructure/log";
import { WebServer } from "./infrastructure/rpc/http/Server";
import { HealthController } from "./infrastructure/rpc/http/HealthController";
import { StartJobs } from "./domain/actions/index";

let repos: RepositoriesBuilder;
let server: WebServer;

async function run(): Promise<void> {
  log.info(`Starting: dryRunEnabled -> ${configuration.dryRun}`);

  repos = new RepositoriesBuilder(configuration);
  await repos.init();

  const startJobs = new StartJobs(repos.getJobsRepository(), repos.getJobExecutionRepository());

  await startServer(repos);
  await startJobs.run();

  log.info("Started");

  // Handle shutdown
  process.on("SIGINT", handleShutdown);
  process.on("SIGTERM", handleShutdown);
}

const startServer = async (repos: RepositoriesBuilder) => {
  server = new WebServer(configuration.port, new HealthController(repos.getStatsRepository()));
};

const handleShutdown = async () => {
  try {
    await Promise.allSettled([repos.close(), server.stop()]);

    process.exit();
  } catch (error: unknown) {
    process.exit(1);
  }
};

run().catch((e) => {
  log.error("Fatal error caused process to exit", e);
});
