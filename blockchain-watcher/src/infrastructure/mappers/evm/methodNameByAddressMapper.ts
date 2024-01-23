import { EvmTransaction } from "../../../domain/entities";

const TESTNET_ENVIRONMENT = "testnet";

export const methodNameByAddressMapper = (
  chain: string,
  environment: string,
  transaction: EvmTransaction
): Protocol | undefined => {
  const address = transaction.to;
  const input = transaction.input;

  if (environment == TESTNET_ENVIRONMENT) {
    return methodsByAddressTestnet(chain, address, input);
  } else {
    return methodsByAddressMainnet(chain, address, input);
  }
};

const methodsByAddressTestnet = (
  chain: string,
  address: string,
  input: string
): Protocol | undefined => {
  const testnet: MethodsByAddress = {
    ethereum: [
      {
        [String("0xF890982f9310df57d00f659cf4fd87e65adEd8d7").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0x4cb69FaE7e7Af841e44E1A1c30Af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    "ethereum-sepolia": [
      {
        [String("0x4a8bc80Ed5a4067f1CCf107057b8270E0cC11A78").toLowerCase()]: ethBase,
      },
      {
        [String("0x2703483B1a5a7c577e8680de9Df8Be03c6f30e3c").toLowerCase()]: ccttp,
      },
    ],
    polygon: [
      {
        [String("0x377D55a7928c046E18eEbb61977e714d2a76472a").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0xc3D46e0266d95215589DE639cC4E93b79f88fc6C").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    bsc: [
      {
        [String("0x9dcF9D205C9De35334D646BeE44b2D2859712A09").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    fantom: [
      {
        [String("0x599CEa2204B4FaECd584Ab1F2b6aCA137a0afbE8").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    avalanche: [
      {
        [String("0x61E44E506Ca5659E6c0bba9b678586fA2d729756").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    oasis: [
      {
        [String("0x88d8004A9BdbfD9D28090A02010C19897a29605c").toLowerCase()]: ethBase,
      },
    ],
    moonbean: [
      {
        [String("0xbc976D4b9D57E57c3cA52e1Fd136C45FF7955A96").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    celo: [
      {
        [String("0x05ca6037eC51F8b712eD2E6Fa72219FEaE74E153").toLowerCase()]: ethBase,
      },
      {
        [String("0x9563a59C15842a6f322B10f69d1dD88b41f2E97B").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    arbitrum: [
      {
        [String("0x23908A62110e21C04F3A4e011d24F901F911744A").toLowerCase()]: ethBase,
      },
      {
        [String("0xe3e0511EEbD87F08FbaE4486419cb5dFB06e1343").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    "arbitrum-sepolia": [
      {
        [String("0xC7A204bDBFe983FCD8d8E61D02b475D4073fF97e").toLowerCase()]: ethBase,
      },
      {
        [String("0x6b9C8671cdDC8dEab9c719bB87cBd3e782bA6a35").toLowerCase()]: ethBase,
      },
      {
        [String("0x2703483B1a5a7c577e8680de9Df8Be03c6f30e3c").toLowerCase()]: ccttp,
      },
    ],
    optimism: [
      {
        [String("0xC7A204bDBFe983FCD8d8E61D02b475D4073fF97e").toLowerCase()]: ethBase,
      },
      {
        [String("0xc3D46e0266d95215589DE639cC4E93b79f88fc6C").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    "optimism-sepolia": [
      {
        [String("0x99737Ec4B815d816c49A385943baf0380e75c0Ac").toLowerCase()]: ethBase,
      },
      {
        [String("0x31377888146f3253211EFEf5c676D41ECe7D58Fe").toLowerCase()]: ethBase,
      },
      {
        [String("0x2703483B1a5a7c577e8680de9Df8Be03c6f30e3c").toLowerCase()]: ccttp,
      },
    ],
    base: [
      {
        [String("0xA31aa3FDb7aF7Db93d18DDA4e19F811342EDF780").toLowerCase()]: base,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    "base-sepolia": [
      {
        [String("0x79A1027a6A159502049F10906D333EC57E95F083").toLowerCase()]: ccttp,
      },
      {
        [String("0x2703483B1a5a7c577e8680de9Df8Be03c6f30e3c").toLowerCase()]: ccttp,
      },
    ],
  };

  return findMethodName(testnet, chain, address, input);
};

const methodsByAddressMainnet = (
  chain: string,
  address: string,
  input: string
): Protocol | undefined => {
  const mainnet: MethodsByAddress = {
    ethereum: [
      {
        [String("0x3ee18B2214AFF97000D974cf647E7C347E8fa585").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0xd8E1465908103eD5fd28e381920575fb09beb264").toLowerCase()]: receiveMessageAndSwap,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    polygon: [
      {
        [String("0x5a58505a96D1dbf8dF91cB21B54419FC36e93fdE").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0x09959798B95d00a3183d20FaC298E4594E599eab").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0xf6C5FD2C8Ecba25420859f61Be0331e68316Ba01").toLowerCase()]: receiveMessageAndSwap,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    bsc: [
      {
        [String("0xB6F6D86a8f9879A9c87f643768d9efc38c1Da6E7").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    fantom: [
      {
        [String("0x7C9Fc5741288cDFdD83CeB07f3ea7e22618D79D2").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    avalanche: [
      {
        [String("0x0e082F06FF657D94310cB8cE8B0D9a04541d8052").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    oasis: [
      {
        [String("0x5848C791e09901b40A9Ef749f2a6735b418d7564").toLowerCase()]: ethBase,
      },
    ],
    moonbean: [
      {
        [String("0xb1731c586ca89a23809861c6103f0b96b3f57d92").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    celo: [
      {
        [String("0x796Dff6D74F3E27060B71255Fe517BFb23C93eed").toLowerCase()]: ethBase,
      },
      {
        [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]:
          completeTransferWithRelay,
      },
    ],
    arbitrum: [
      {
        [String("0x0b2402144Bb366A632D14B83F244D2e0e21bD39c").toLowerCase()]: ethBase,
      },
      {
        [String("0x1293a54e160D1cd7075487898d65266081A15458").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0xf8497FE5B0C5373778BFa0a001d476A21e01f09b").toLowerCase()]: receiveMessageAndSwap,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    optimism: [
      {
        [String("0x1D68124e65faFC907325e3EDbF8c4d84499DAa8b").toLowerCase()]: ethBase,
      },
      {
        [String("0x1293a54e160D1cd7075487898d65266081A15458").toLowerCase()]: receiveTbtc,
      },
      {
        [String("0xcF205Fa51D33280D9B70321Ae6a3686FB2c178b2").toLowerCase()]: receiveMessageAndSwap,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
    base: [
      {
        [String("0x8d2de8d2f73F1F4cAB472AC9A881C9b123C79627").toLowerCase()]: base,
      },
      {
        [String("0x9816d7C448f79CdD4aF18c4Ae1726A14299E8C75").toLowerCase()]: receiveMessageAndSwap,
      },
      {
        [String("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").toLowerCase()]: ccttp,
      },
    ],
  };

  return findMethodName(mainnet, chain, address, input);
};

const findMethodName = (
  environment: MethodsByAddress,
  chain: string,
  address: string,
  input: string
): Protocol | undefined => {
  const first10Characters = input.slice(0, 10);
  let protocol: Protocol | undefined;

  environment[chain]?.find((addresses) => {
    const protocols = addresses[address];
    const foundProtocol = protocols?.get(first10Characters);
    protocol = foundProtocol;
    return foundProtocol;
  });

  return protocol;
};

export enum MethodID {
  // Method ids for wormhole token bridge contract
  MethodIDCompleteTransfer = "0xc6878519",
  MethodIDWrapAndTransfer = "0x9981509f",
  MethodIDTransferTokens = "0x0f5287b0",
  MethodIDAttestToken = "0xc48fa115",
  MethodIDCompleteAndUnwrapETH = "0xff200cde",
  MethodIDCreateWrapped = "0xe8059810",
  MethodIDUpdateWrapped = "0xf768441f",
  // Method id for wormhole connect wrapped contract.
  MethodCompleteTransferWithRelay = "0x2f25e25f",
  // Method id for wormhole tBTC gateway
  MethodIDReceiveTbtc = "0x5d21a596",
  // Method id for Portico contract
  MethodIDReceiveMessageAndSwap = "0x3d528f35",
  // Method id for CTTP contract
  MethodIDRedeemTokensCCTP = "0x0a55d735", // Automatic relayer process
  MethodIDReceiveMessageCCTP = "0x57ecfd28", // Manual process
}

const ethBase = new Map<string, Protocol>([
  [
    MethodID.MethodIDCompleteTransfer,
    { method: "MethodCompleteTransfer", name: "transfer-redeemed" },
  ],
  [
    MethodID.MethodIDCompleteAndUnwrapETH,
    { method: "MethodCompleteAndUnwrapETH", name: "transfer-redeemed" },
  ],
  [MethodID.MethodIDCreateWrapped, { method: "MethodCreateWrapped", name: "transfer-redeemed" }],
  [MethodID.MethodIDUpdateWrapped, { method: "MethodUpdateWrapped", name: "transfer-redeemed" }],
]);

const completeTransferWithRelay = new Map<string, Protocol>([
  [
    MethodID.MethodCompleteTransferWithRelay,
    { method: "MethodCompleteTransferWithRelay", name: "standard-relay-delivered" },
  ],
]);

const receiveMessageAndSwap = new Map<string, Protocol>([
  [MethodID.MethodIDReceiveMessageAndSwap, { method: "MethodReceiveMessageAndSwap", name: "" }], // TODO: When active this protocol set the name
]);

const receiveTbtc = new Map<string, Protocol>([
  [MethodID.MethodIDReceiveTbtc, { method: "MethodReceiveTbtc", name: "" }], // TODO: When active this protocol set the name
]);

const base = new Map<string, Protocol>([...ethBase, ...completeTransferWithRelay]);

const ccttp = new Map<string, Protocol>([
  [
    MethodID.MethodIDRedeemTokensCCTP,
    { method: "MethodRedeemTokensCCTP", name: "transfer-redeemed" },
  ],
  [
    MethodID.MethodIDReceiveMessageCCTP,
    { method: "MethodReceiveMessageCCTP", name: "transfer-redeemed" },
  ],
]);

type MethodsByAddress = {
  [chain: string]: {
    [address: string]: Map<string, Protocol>;
  }[];
};

type Protocol = {
  method: string;
  name: string;
};
