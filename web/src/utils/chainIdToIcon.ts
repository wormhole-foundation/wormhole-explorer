import {
  CHAIN_ID_ACALA,
  CHAIN_ID_ALGORAND,
  CHAIN_ID_AURORA,
  CHAIN_ID_AVAX,
  CHAIN_ID_BSC,
  CHAIN_ID_CELO,
  CHAIN_ID_ETH,
  CHAIN_ID_FANTOM,
  CHAIN_ID_KARURA,
  CHAIN_ID_KLAYTN,
  CHAIN_ID_MOONBEAM,
  CHAIN_ID_NEAR,
  CHAIN_ID_POLYGON,
  CHAIN_ID_SOLANA,
  CHAIN_ID_TERRA,
  CHAIN_ID_TERRA2,
} from "@certusone/wormhole-sdk";
import acalaIcon from "../icons/acala.svg";
import algorandIcon from "../icons/algorand.svg";
import auroraIcon from "../icons/aurora.svg";
import avaxIcon from "../icons/avax.svg";
import bscIcon from "../icons/bsc.svg";
import celoIcon from "../icons/celo.svg";
import ethIcon from "../icons/eth.svg";
import fantomIcon from "../icons/fantom.svg";
import karuraIcon from "../icons/karura.svg";
import klaytnIcon from "../icons/klaytn.svg";
import moonbeamIcon from "../icons/moonbeam.svg";
import nearIcon from "../icons/near.svg";
import polygonIcon from "../icons/polygon.svg";
import solanaIcon from "../icons/solana.svg";
import terraIcon from "../icons/terra.svg";
import terra2Icon from "../icons/terra2.svg";

const chainIdToIconMap: { [id: number]: string } = {
  [CHAIN_ID_SOLANA]: solanaIcon,
  [CHAIN_ID_ETH]: ethIcon,
  [CHAIN_ID_TERRA]: terraIcon,
  [CHAIN_ID_BSC]: bscIcon,
  [CHAIN_ID_ACALA]: acalaIcon,
  [CHAIN_ID_ALGORAND]: algorandIcon,
  [CHAIN_ID_AURORA]: auroraIcon,
  [CHAIN_ID_AVAX]: avaxIcon,
  [CHAIN_ID_CELO]: celoIcon,
  [CHAIN_ID_FANTOM]: fantomIcon,
  [CHAIN_ID_TERRA2]: terra2Icon,
  [CHAIN_ID_KARURA]: karuraIcon,
  [CHAIN_ID_KLAYTN]: klaytnIcon,
  [CHAIN_ID_MOONBEAM]: moonbeamIcon,
  [CHAIN_ID_NEAR]: nearIcon,
  [CHAIN_ID_POLYGON]: polygonIcon,
};
const chainIdToIcon = (chainId: number) => chainIdToIconMap[chainId] || "";
export default chainIdToIcon;
