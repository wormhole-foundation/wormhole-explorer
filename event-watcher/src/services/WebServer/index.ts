import { createServer } from '../../builder/server';
import { env } from '../../config';
import { InfrastructureController } from '../../infrastructure/infrastructure.controller';

class WebServer {
  constructor(private infrastructureController: InfrastructureController) {
    console.log('[Webserver]', 'Initializing...');
  }

  public async start() {
    console.log('[Webserver]', `Creating...`);

    const port = Number(env.PORT) || 3005;
    const server = await createServer(port);

    server.get('/ready', { logLevel: 'silent' }, this.infrastructureController.ready);
    server.get('/health', { logLevel: 'silent' }, this.infrastructureController.health);

    try {
      const address = await server.listen({ host: '0.0.0.0', port });
      console.log('[Webserver]', `Listening ${address}`);
    } catch (err) {
      server.log.error(err);
      process.exit(1);
    }
  }

  public stop() {
    console.log('WebServer stopping...');
  }
}

export default WebServer;
