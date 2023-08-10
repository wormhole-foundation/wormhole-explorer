import axios from 'axios';
import { AXIOS_CONFIG_JSON, GUARDIAN_RPC_HOSTS } from '../consts';

export const getSignedVAA = async (
  chain: number,
  emitter: string,
  sequence: string
): Promise<Buffer | null> => {
  for (const host of GUARDIAN_RPC_HOSTS) {
    try {
      const result = await axios.get(
        `${host}/v1/signed_vaa/${chain}/${emitter}/${sequence.toString()}`,
        AXIOS_CONFIG_JSON
      );
      if (result.data.vaaBytes) {
        return Buffer.from(result.data.vaaBytes, 'base64');
      }
    } catch (e) {}
  }
  return null;
};
