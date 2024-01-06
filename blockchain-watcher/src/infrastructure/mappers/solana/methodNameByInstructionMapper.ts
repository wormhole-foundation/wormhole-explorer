import { solana } from "../../../domain/entities";

const TRANSFER_REDEEMED_NAME = "transfer-redeemed";

export const methodNameByInstructionMapper = (
  instruction: solana.MessageCompiledInstruction,
  programIdIndex: number
): Status | undefined => {
  const data = instruction.data;

  if (!programIdIndex || instruction.programIdIndex != Number(programIdIndex) || data.length == 0) {
    return {
      id: MethodID.unknownInstructionID,
      name: TRANSFER_REDEEMED_NAME,
      method: Method.unknownInstruction.toString(),
    };
  }

  const methodId = data[0];
  const selectedMethod = methodsMapping[methodId] || Method.unknownInstruction;

  return {
    id: methodId,
    name: selectedMethod.name,
    method: selectedMethod.method.toString(),
  };
};

type Status = {
  id: number;
  name: string;
  method: string;
};

enum MethodID {
  completeWrappedInstructionID = 0x3,
  completeNativeInstructionID = 0x2,
  unknownInstructionID = 0x0,
}

enum Method {
  completeWrappedInstruction,
  completeNativeInstruction,
  unknownInstruction,
}

const methodsMapping: { [key: number]: { method: Method; name: string } } = {
  [MethodID.completeWrappedInstructionID]: {
    method: Method.completeWrappedInstruction,
    name: TRANSFER_REDEEMED_NAME,
  },
  [MethodID.completeNativeInstructionID]: {
    method: Method.completeNativeInstruction,
    name: TRANSFER_REDEEMED_NAME,
  },
  [MethodID.unknownInstructionID]: {
    method: Method.unknownInstruction,
    name: TRANSFER_REDEEMED_NAME,
  },
};
