import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { AptosTransaction } from "../../../../src/domain/entities/aptos";
import {
  PollAptosTransactionsMetadata,
  PollAptosTransactionsConfig,
  PollAptos,
} from "../../../../src/domain/actions/aptos/PollAptos";
import {
  MetadataRepository,
  AptosRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let getTransactionsByVersionSpy: jest.SpiedFunction<AptosRepository["getTransactionsByVersion"]>;
let getTransactionsSpy: jest.SpiedFunction<AptosRepository["getTransactions"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollAptosTransactionsMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: AptosTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>;
let aptosRepo: AptosRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: AptosTransaction[]) => Promise.resolve(),
  failing: (txs: AptosTransaction[]) => Promise.reject(),
};
let pollAptos: PollAptos;

let props = {
  blockBatchSize: 100,
  from: 0n,
  limit: 0n,
  environment: "testnet",
  commitment: "finalized",
  addresses: ["0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"],
  interval: 5000,
  topics: [],
  chainId: 22,
  filters: [
    {
      address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      type: "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::submit_vaa_and_register_entry",
    },
  ],
  chain: "aptos",
  id: "poll-log-message-published-aptos",
};

let cfg = new PollAptosTransactionsConfig(props);

describe("GetAptosTransactions", () => {
  afterEach(async () => {
    await pollAptos.stop();
  });

  it("should be not generate range (from and limit) and search the latest from plus from batch size cfg, and not process tx because is not a wormhole redeem", async () => {
    // Given
    const tx = {
      version: "487572390",
      hash: "0x487a4bfb6a7cda97090637ca5485afc3cb25e6eb21a873097ee3f0dcedc0b3b8",
      state_change_hash: "0x40f10f464ce9301249fa44104c186a776a22b56cdffc94b6e3c4787d5d600538",
      event_root_hash: "0x5017fa0a3016560a57eb8ed817ddbb0306d86c47fb4331ff993478c0acde30ca",
      state_checkpoint_hash: null,
      gas_used: "164",
      success: true,
      vm_status: "Executed successfully",
      accumulator_root_hash: "0x26b6defda012418727f34c701fd9e65ed9e5f8e2e59aac7ebeb0ca7038a8d647",
      changes: [],
      sender: "0x50bc83f01d48ab3b9c00048542332201ab9cbbea61bda5f48bf81dc506caa78f",
      sequence_number: "1494185",
      max_gas_amount: "100000",
      gas_unit_price: "100",
      expiration_timestamp_secs: "1709822634",
      payload: {
        function:
          "0xb1421c3d524a353411aa4e3cce0f0ce7f404a12da91a2889e1bc3cea6ffb17da::cancel_and_place::cancel_and_place",
        type_arguments: [
          "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::WETH",
          "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC",
        ],
        arguments: [
          "8",
          ["42645072269700382266349876", "42644869357063623703648693"],
          ["42645090718696431883245683", "42644906249285249691670581"],
          ["2517", "10041"],
          ["383120", "383216"],
          ["2539", "10008"],
          ["382928", "382832"],
          "0x50bc83f01d48ab3b9c00048542332201ab9cbbea61bda5f48bf81dc506caa78f",
          3,
          0,
        ],
        type: "entry_function_payload",
      },
      signature: {
        public_key: "0xc7756ecfa532b78c375a20e89910bf0120a9ec3431a02ed7e0e14999928d047d",
        signature:
          "0x09c4753e2efc67fd08e2060a97435e579f2157d95cc469568a0c9be804325cb4c8f1c1dc056763d866fa2eda5bf7b5b1ccbdce45aec1b77db79293bc855f5f02",
        type: "ed25519_signature",
      },
      events: [
        {
          guid: {
            creation_number: "10",
            account_address: "0x50bc83f01d48ab3b9c00048542332201ab9cbbea61bda5f48bf81dc506caa78f",
          },
          sequence_number: "2291689",
          type: "0xc0deb00c405f84c85dc13442e305df75d1288100cdd82675695f6148c7ece51c::user::CancelOrderEvent",
          data: {
            custodian_id: "0",
            market_id: "8",
            order_id: "42645072269700382266349876",
            reason: 3,
            user: "0x50bc83f01d48ab3b9c00048542332201ab9cbbea61bda5f48bf81dc506caa78f",
          },
        },
      ],
      timestamp: "1709822034948509",
      type: "user_transaction",
    };

    givenAptosBlockRepository(tx);
    givenMetadataRepository({ lastFrom: 423525334n });
    givenStatsRepository();
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith({ from: 423525334, limit: 100 })
    );
  });

  it("should be use from and batch size cfg, and process tx because is a wormhole redeem", async () => {
    // Given
    const tx = {
      version: "487581688",
      hash: "0x2853cb063b5351ea1b1ea46295bfadcd18117d20bc7b65de8db624284fd19061",
      state_change_hash: "0x1513a95994c2b8319cef0bb728e56dcf51519cc5982d494541dbb91e7ba9ee2e",
      event_root_hash: "0xadc25c39d0530da619a7620261194d6f3911aeed8c212dc3dfb699b2b6a07834",
      state_checkpoint_hash: null,
      gas_used: "753",
      success: true,
      vm_status: "Executed successfully",
      accumulator_root_hash: "0x9b6e4552555f3584e13910c3a998159fef6c31568a23a9934832661f5bde5a09",
      changes: [
        {
          address: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
          state_key_hash: "0xaf2393fef64599629efda83739a73fea2fc70c4d9bdff14e5681c396c51ab8f6",
          data: {
            type: "0x1::coin::CoinStore<0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T>",
            data: {
              coin: { value: "524897921" },
              deposit_events: {
                counter: "1291",
                guid: {
                  id: {
                    addr: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
                    creation_num: "4",
                  },
                },
              },
              frozen: false,
              withdraw_events: {
                counter: "750",
                guid: {
                  id: {
                    addr: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
                    creation_num: "5",
                  },
                },
              },
            },
          },
          type: "write_resource",
        },
        {
          address: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
          state_key_hash: "0x18a0f4ffd938393773095ff40524b113a48e1c088fef980f202096402be6bd7b",
          data: {
            type: "0x1::account::Account",
            data: {
              authentication_key:
                "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
              coin_register_events: {
                counter: "25",
                guid: {
                  id: {
                    addr: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
                    creation_num: "0",
                  },
                },
              },
              guid_creation_num: "52",
              key_rotation_events: {
                counter: "0",
                guid: {
                  id: {
                    addr: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
                    creation_num: "1",
                  },
                },
              },
              rotation_capability_offer: { for: { vec: [] } },
              sequence_number: "2050",
              signer_capability_offer: { for: { vec: [] } },
            },
          },
          type: "write_resource",
        },
      ],
      sender: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
      sequence_number: "2049",
      max_gas_amount: "904",
      gas_unit_price: "100",
      expiration_timestamp_secs: "1709822592",
      payload: {
        function:
          "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::submit_vaa_and_register_entry",
        type_arguments: [
          "0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T",
        ],
        arguments: [
          "0x01000000030d0160b44a3ef503cf5a38ea72be662214c57b1431818e96fce55581085dc926379461f7942c11676bb9e992bf3dfca07c2ed1fc6d6df2dc50f1efee226cc66efae301039451110dcf560d2a79b8ba9aca679bcd56eba1c4db1c8f6f44c12973e394f2126b5e8944472b513b3196323bdddd31de64c20294dda6858b3ca77d0c8de940730004ffd7ac756815bce35c1e2cb72247ab4ab4ead25f4b072dc1c74937101b1e09cf04475447fbff003adb686338816d50bc3fedf988e733f03f8c11c359895b5a05000642879078f0c4964ed01bd46e977a032acc8aa0bf5140df04349c2fa843ee152a570b03286040702c59a40b9ddc8e99cd3d46c5529ce9f2d5cc3a3c0b7aafcd960007f910ed72608e482fce0b2afcc243dcb3e7b0f2e450215c8f056288e631abc8966467f976eb39c992f20ef61b47b3587f0566fc934880357ad019da98ba29768c00088aa7b98f820581e011a5e554d347f5d714e16a74674d6117e3772ee77b20ca30157144b31ec203e18c8815f3c21feca7e549d5111e48c2f7d9362139abeb6bd1010989e8c573acb24be1915f0152f790d53c195ee8850eb7e41988e6e019e4a6decf390d19a16ae161b65ef265ff75a01b3ad23fe9e99d124e140a6d64fd8808d738010a9aa6b777d25f2bf4500f6ce574617350a136e217c2bfe8ba97a62a6cf3375eff158fb81149ae439f0f43ae7727189a8f6b5557318caa35829484ba1f5abb3eb8000c12fc58d0e3cfa020f126495e5d3f9401ac82dc48d6af1a22d8950b9cdb5fd8370181d4ca59ec41cc193bbb39618b1b645d557999ef97e89c919cec99aabe42c7010daa874ee1b0d6f3462c8f83d108953eb80d797d908d3c121c46e8a5514f50f2470822076d8f59e2934c4ee365520c53f9ea9ee3014d0ed8adc85c3653a0ac77000010e31607f6bfd58aa3d9c423b410c1081cbef7ee215a6a412c12ba60807867bcbd680d82a5ecf92ac99bb0682e670c38415304ed6eb9e7c95b0095866fdd2a67a100115e20281a08ebb632a9ff584f11924dd62edcaaceb89123814ce3598bd231b6d8678c13876a1f7c0b8c07acb7e0fad4d002f839db2f6f84b5f2ddfdc4ba3930f001126414cd214b9d5e0e0dbb2fad61ee9d567ffb79b7a575d84acd0bc8dbd28ea15f74feca54a47c5adae94b1718556b44f88a7993dc78602b9907738b0bd6427dcd0165e9d1e4400b01000010000000000000000000000000b1731c586ca89a23809861c6103f0b96b3f57d92000000000001219e0101000000000000000000000000000000000000000000000000000000001dcd6500000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480002a9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd200160000000000000000000000000000000000000000000000000000000000000000",
        ],
        type: "entry_function_payload",
      },
      signature: {
        public_key: "0x5d4d25db389615c6ba75d3920690a8dfbe2108e21835668ebdee9d2096b7e104",
        signature:
          "0xd849b7400247c747268ccc30a6879995d8792e83116cccbe53ba564bd0517f0c82299302014fd7bd089e6e00090b0d9ad46c8f2ecd1bd81ef2d6d0ae094a870f",
        type: "ed25519_signature",
      },
      events: [
        {
          guid: {
            creation_number: "4",
            account_address: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
          },
          sequence_number: "1289",
          type: "0x1::coin::DepositEvent",
          data: { amount: "500000000" },
        },
        {
          guid: {
            creation_number: "4",
            account_address: "0xa9e33cfc7bb1f0d8fc63dcfbfecaff4806facfee290c284cd83ae10763ca0bd2",
          },
          sequence_number: "1290",
          type: "0x1::coin::DepositEvent",
          data: { amount: "0" },
        },
        {
          guid: { creation_number: "0", account_address: "0x0" },
          sequence_number: "0",
          type: "0x1::transaction_fee::FeeStatement",
          data: {
            execution_gas_units: "118",
            io_gas_units: "8",
            storage_fee_octas: "62820",
            storage_fee_refund_octas: "0",
            total_charge_gas_units: "753",
          },
        },
      ],
      timestamp: "1709822474112433",
      type: "user_transaction",
    };
    givenAptosBlockRepository(tx);
    givenMetadataRepository();
    givenStatsRepository();
    // Se from for cfg
    props.from = 146040n;
    givenPollAptosTx(cfg);

    // Whem
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith({ from: 146040, limit: 100 }),
      () => expect(getTransactionsByVersionSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be return the same lastFrom and limit equal 1000", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousFrom: 146040n, lastFrom: 146040n });
    givenStatsRepository();
    // Se from for cfg
    props.from = 0n;
    givenPollAptosTx(cfg);

    // Whem
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith({ from: 146040, limit: 100 })
    );
  });

  it("should be if return the lastFrom and limit equal the block batch size", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousFrom: undefined, lastFrom: 146040n });
    givenStatsRepository();
    // Se from for cfg
    props.from = 0n;
    givenPollAptosTx(cfg);

    // Whem
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith({ from: 146040, limit: 100 })
    );
  });
});

const givenAptosBlockRepository = (tx: any = {}) => {
  const events = [
    {
      version: "481740133",
      guid: {
        creation_number: "2",
        account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      },
      sequence_number: "148985",
      type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
      data: {
        consistency_level: 0,
        nonce: "41611",
        payload:
          "0x0100000000000000000000000000000000000000000000000000000000003b826000000000000000000000000082af49447d8a07e3bd95bd0d56f35241523fbab10017000000000000000000000000451febd0f01b9d6bda1a5b9d0b6ef88026e4a79100170000000000000000000000000000000000000000000000000000000000011c1e",
        sender: "1",
        sequence: "146040",
        timestamp: "1709585379",
      },
    },
  ];

  const txs = [tx];

  aptosRepo = {
    getEventsByEventHandle: () => Promise.resolve(events),
    getTransactionsByVersion: () => Promise.resolve([]),
    getTransactions: () => Promise.resolve(txs),
  };

  getTransactionsSpy = jest.spyOn(aptosRepo, "getTransactions");
  getTransactionsByVersionSpy = jest.spyOn(aptosRepo, "getTransactionsByVersion");
  handlerSpy = jest.spyOn(handlers, "working");
};

const givenMetadataRepository = (data?: PollAptosTransactionsMetadata) => {
  metadataRepo = {
    get: () => Promise.resolve(data),
    save: () => Promise.resolve(),
  };
  metadataSaveSpy = jest.spyOn(metadataRepo, "save");
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

const givenPollAptosTx = (cfg: PollAptosTransactionsConfig) => {
  pollAptos = new PollAptos(cfg, statsRepo, metadataRepo, aptosRepo, "GetAptosTransactions");
};

const whenPollAptosLogsStarts = async () => {
  pollAptos.run([handlers.working]);
};
