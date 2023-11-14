import { PollEvmLogs, PollEvmLogsConfig, HandleEvmLogs } from "./domain/actions";
import { LogFoundEvent } from "./domain/entities";
import { configuration } from "./infrastructure/config";
import { evmLogMessagePublishedMapper } from "./infrastructure/mappers/evmLogMessagePublishedMapper";
import { RepositoriesBuilder } from "./infrastructure/RepositoriesBuilder";
import log from "./infrastructure/log";
import { WebServer } from "./infrastructure/rpc/Server";
import { HealthController } from "./infrastructure/rpc/HealthController";

let repos: RepositoriesBuilder;
let server: WebServer;

async function run(): Promise<void> {
  log.info(`Starting: dryRunEnabled -> ${configuration.dryRun}`);

  repos = new RepositoriesBuilder(configuration);

  await startServer(repos);
  await startJobs(repos);

  // Just keep this running until killed
  setInterval(() => {
    log.info("Still running");
  }, 20_000);

  log.info("Started");

  // Handle shutdown
  process.on("SIGINT", handleShutdown);
  process.on("SIGTERM", handleShutdown);
}

const startServer = async (repos: RepositoriesBuilder) => {
  server = new WebServer(configuration.port, new HealthController(repos.getStatsRepository()));
};

const startJobs = async (repos: RepositoriesBuilder) => {
  /** Job definition is hardcoded, but should be loaded from cfg or a data store soon enough */
  const jobs = [
    {
      id: "poll-log-message-published-ethereum",
      chain: "ethereum",
      source: {
        action: "PollEvmLogs",
        config: {
          fromBlock: 10012499n,
          blockBatchSize: 100,
          commitment: "latest",
          interval: 15_000,
          addresses: ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
          chain: "ethereum",
          topics: [],
        },
      },
      handlers: [
        {
          action: "HandleEvmLogs",
          target: "sns",
          mapper: "evmLogMessagePublishedMapper",
          config: {
            abi: "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)",
            filter: {
              addresses: ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
              topics: ["0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2"],
            },
          },
        },
      ],
    },
  ];

  const pollEvmLogs = new PollEvmLogs(
    repos.getEvmBlockRepository("ethereum"),
    repos.getMetadataRepository(),
    repos.getStatsRepository(),
    new PollEvmLogsConfig({ ...jobs[0].source.config, id: jobs[0].id })
  );

  const snsTarget = async (events: LogFoundEvent<any>[]) => {
    const result = await repos.getSnsEventRepository().publish(events);
    if (result.status === "error") {
      log.error(`Error publishing events to SNS: ${result.reason ?? result.reasons}`);
      throw new Error(`Error publishing events to SNS: ${result.reason}`);
    }
    log.info(`Published ${events.length} events to SNS`);
  };

  const handleEvmLogs = new HandleEvmLogs<LogFoundEvent<any>>(
    jobs[0].handlers[0].config,
    evmLogMessagePublishedMapper,
    configuration.dryRun
      ? async (events) => {
          log.info(`Got ${events.length} events`);
        }
      : snsTarget
  );

  pollEvmLogs.start([handleEvmLogs.handle.bind(handleEvmLogs)]);
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
