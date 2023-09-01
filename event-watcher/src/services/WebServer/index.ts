import { createServer } from '../../builder/server';
import { env } from '../../config';
import { InfrastructureController } from '../../infrastructure/infrastructure.controller';

class WebServer {
  private server?: Awaited<ReturnType<typeof createServer>>;

  constructor(private infrastructureController: InfrastructureController) {
    console.log('[Webserver]', 'Initializing...');
  }

  public async start() {
    console.log('[Webserver]', `Creating...`);

    const port = Number(env.PORT) || 3005;
    this.server = await createServer(port);

    this.server.get('/ready', { logLevel: 'silent' }, this.infrastructureController.ready);
    this.server.get('/health', { logLevel: 'silent' }, this.infrastructureController.health);

    try {
      const address = await this.server.listen({ host: '0.0.0.0', port });
      console.log('[Webserver]', `Listening ${address}`);
    } catch (err) {
      this.server.log.error(err);
      process.exit(1);
    }
  }

  public async stop() {
    console.log('[Webserver] Stopping...');
    await this.server?.close();
    console.log('[Webserver] Stopped');
  }
}

export default WebServer;
