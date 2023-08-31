import winston, { createLogger, format, Logger, LoggerOptions, transports } from 'winston';
import { env } from '../config';
import { toArray } from './array';

const { combine, errors, colorize } = format;
let logger: WormholeLogger | undefined = undefined;

export type WormholeLogger = Logger & { labels: string[] };

/**
 * Get a logger that is scoped to the given labels. If a parent logger is
 * provided, the parent's labels will be prepended to the given labels.
 * TODO: add support for custom log levels for scoped loggers
 *
 * Assuming `LOG_LEVEL=info`, the loggers below will output the following logs.
 * ```
 * getLogger().info(1); // base logger
 * const foo = getLogger('foo'); // implicitly uses base logger
 * foo.error(2)
 * getLogger('bar', foo).debug(3); // not logged because LOG_LEVEL=info
 * getLogger('bar', foo).warn(4);
 *
 * [2022-12-20 05:04:34.168 +0000] [info] [main] 1
 * [2022-12-20 05:04:34.170 +0000] [error] [foo] 2
 * [2022-12-20 05:04:34.170 +0000] [warn] [foo | bar] 4
 * ```
 * @param labels
 * @param parent
 * @returns
 */
export const getLogger = (
  labels: string | string[] = [],
  parent?: WormholeLogger,
): WormholeLogger => {
  // base logger is parent if unspecified
  if (!parent) parent = logger = logger ?? createBaseLogger();

  // no labels, return parent logger
  labels = toArray(labels);
  if (labels.length === 0) return parent;

  // create scoped logger
  const child: WormholeLogger = parent.child({
    labels: [...parent.labels, ...labels],
  }) as WormholeLogger;
  child.labels = labels;
  return child;
};

const createBaseLogger = (): WormholeLogger => {
  const { LOG_LEVEL, LOG_DIR } = env;
  const LOG_PATH = LOG_DIR ? `${LOG_DIR}/watcher.${new Date().toISOString()}.log` : null;
  console.log(`[Logger] Logging to ${LOG_PATH ?? 'the console'} at level ${LOG_LEVEL}`);

  const appendLoggerName = format((info) => {
    info.logger = 'wormhole-explorer-event-watcher';
    return info;
  });

  const appendTimestampISO = format((info) => {
    info.ts = new Date().toISOString();
    return info;
  });

  const loggerConfig: LoggerOptions = {
    level: LOG_LEVEL.toLowerCase() || 'debug',
    format: combine(
      appendLoggerName(),
      appendTimestampISO(),
      errors({ stack: true }),
      format.json(),
    ),
    transports: [
      LOG_PATH
        ? new transports.File({
            filename: LOG_PATH,
          })
        : new winston.transports.Console({
            format: combine(colorize({ all: true })),
          }),
    ],
    exitOnError: false,
  };

  const logger = createLogger(loggerConfig) as WormholeLogger;
  logger.labels = [];
  return logger;
};
