import { solana } from "../../../domain/entities";

export const methodNameByInstructionMapper = (
  instruction: solana.MessageCompiledInstruction,
  programIdIndex: number
): Status | undefined => {
  const data = instruction.data;

  if (!programIdIndex || instruction.programIdIndex != Number(programIdIndex) || data.length == 0) {
    return {
      id: MethodID.unknownInstructionID,
      method: Method.unknownInstruction.toString(),
    };
  }

  const methodId = data[0];
  const selectedMethod = methodsMapping[methodId] || Method.unknownInstruction;

  return {
    id: methodId,
    method: selectedMethod?.method?.toString(),
  };
};

type Status = {
  id: number;
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

const methodsMapping: { [key: number]: { method: Method } } = {
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
