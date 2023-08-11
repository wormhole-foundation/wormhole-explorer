import { FastifyReply, FastifyRequest } from "fastify";

export class InfrastructureController {

    ready = async (_: FastifyRequest, reply: FastifyReply) => {
        return reply.code(200).send({ status: "OK" })
    }

    health = async (_: FastifyRequest, reply: FastifyReply) => {
        return reply.code(200).send({ status: "OK" })
    }

}