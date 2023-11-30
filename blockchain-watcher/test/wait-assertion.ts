import { setTimeout } from "timers/promises";

export const thenWaitForAssertion = async (...assertions: (() => void)[]) => {
  for (let index = 1; index < 5; index++) {
    try {
      for (const assertion of assertions) {
        assertion();
      }
      break;
    } catch (error) {
      if (index === 4) {
        throw error;
      }
      await setTimeout(10, undefined, { ref: false });
    }
  }
};
