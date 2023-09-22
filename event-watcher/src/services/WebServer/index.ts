import { createServer } from '../../builder/server';
import { env } from '../../config';
import { InfrastructureController } from '../../infrastructure/infrastructure.controller';
import { getLogger, WormholeLogger } from '../../utils/logger';

class WebServer {
  private server?: Awaited<ReturnType<typeof createServer>>;
  private logger: WormholeLogger;

  constructor(private infrastructureController: InfrastructureController) {
    this.logger = getLogger('WebServer');
    this.logger.info('Initializing...');
  }

  public async start() {
    this.logger.info(`Creating...`);

    const port = Number(env.PORT) || 3005;
    this.server = await createServer(port);

    this.server.get('/ready', { logLevel: 'silent' }, this.infrastructureController.ready);
    this.server.get('/health', { logLevel: 'silent' }, this.infrastructureController.health);

    try {
      const address = await this.server.listen({ host: '0.0.0.0', port });
      this.logger.info(`Listening ${address}`);
    } catch (err) {
      this.server.log.error(err);
      process.exit(1);
    }
  }

  public async stop() {
    this.logger.info('Stopping...');

    await this.server?.close();

    this.logger.info('Stopped');
  }
}

export default WebServer;
