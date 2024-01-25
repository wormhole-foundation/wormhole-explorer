import { solana } from "../../../domain/entities";

export const methodNameByInstructionMapper = (
  instruction: solana.MessageCompiledInstruction,
  programIdIndex: number
): Status => {
  const data = instruction.data;

  if (!programIdIndex || instruction.programIdIndex != Number(programIdIndex) || data.length == 0) {
    return {
      id: MethodID.unknownInstructionID,
      method: Method.unknownInstruction.toString(),
    };
  }

  const methodId = data[0];
  const selectedMethod = methodsMapping[methodId].method || Method.unknownInstruction;

  return {
    id: methodId,
    method: selectedMethod,
  };
};

type Status = {
  id: number;
  method: string;
};

enum MethodID {
  completeWrappedInstructionID = 3,
  completeNativeInstructionID = 2,
  unknownInstructionID = 0,
}

enum Method {
  completeWrappedInstruction = "completeWrappedInstruction",
  completeNativeInstruction = "completeNativeInstruction",
  unknownInstruction = "unknownInstruction",
}

const methodsMapping: { [key: number]: { method: string } } = {
  [MethodID.completeWrappedInstructionID]: {
    method: Method.completeWrappedInstruction,
  },
  [MethodID.completeNativeInstructionID]: {
    method: Method.completeNativeInstruction,
  },
  [MethodID.unknownInstructionID]: {
    method: Method.unknownInstruction,
  },
};
