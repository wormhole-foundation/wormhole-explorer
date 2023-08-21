import { env } from '../config';

let loggingEnv: LoggingEnvironment | undefined = undefined;

export type LoggingEnvironment = {
  logLevel: string;
  logDir?: string;
};

export const getEnvironment = () => {
  if (loggingEnv) {
    return loggingEnv;
  } else {
    loggingEnv = {
      logLevel: env.LOG_LEVEL,
      logDir: env.LOG_DIR,
    };
    return loggingEnv;
  }
};
