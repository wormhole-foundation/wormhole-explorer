import { FastifyReply, FastifyRequest } from 'fastify';
import { DBOptionTypes } from '../databases/types';

export class InfrastructureController {
  constructor(private db: DBOptionTypes) {}

  ready = async (_: FastifyRequest, reply: FastifyReply) => {
    return reply.code(200).send({ status: 'OK' });
  };

  health = async (_: FastifyRequest, reply: FastifyReply) => {
    const isConnected = await this.db.isConnected();
    if (isConnected) {
      return reply.code(200).send({ status: 'OK' });
    } else {
      return reply.code(500).send({ status: 'NOT OK' });
    }
  };
}
