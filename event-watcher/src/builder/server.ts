import Fastify, { FastifyInstance } from 'fastify';

export const createServer = async (port: number): Promise<FastifyInstance> => {
  const server = Fastify({ logger: true });

  await server.register(require('@fastify/swagger'), {
    swagger: {
      info: {
        title: 'VAA Payload Parser',
        description:
          'API allows the parsing of VAA with a custom parser depending on the application that originated the VAA',
        version: '0.0.1',
      },
      externalDocs: {
        url: 'https://swagger.io',
      },
      host: `localhost:${port}`,
      schemes: ['http'],
      consumes: ['application/json'],
      produces: ['application/json'],
    },
  });

  return server;
};
