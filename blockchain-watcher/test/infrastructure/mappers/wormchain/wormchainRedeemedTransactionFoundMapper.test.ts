import { wormchainRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/wormchain/wormchainRedeemedTransactionFoundMapper";
import { describe, it, expect } from "@jest/globals";
import { CosmosRedeem } from "../../../../src/domain/entities/wormchain";

describe("wormchainRedeemedTransactionFoundMapper", () => {
  it("should be able to map log to wormchainRedeemedTransactionFoundMapper", async () => {
    // When
    const result = wormchainRedeemedTransactionFoundMapper(cosmosRedeem) as any;

    // Then
    expect(result.name).toBe("transfer-redeemed");
    expect(result.chainId).toBe(20);
    expect(result.txHash).toBe(
      "0xC196E9E445748AB4BE26E980F685F8F1FD02E8F327F9F1929CE5C426C936BF74"
    );
    expect(result.address).toBe("osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv");
    expect(result.attributes.emitterAddress).toBe(
      "ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5"
    );
    expect(result.attributes.sequence).toBe(128751);
    expect(result.attributes.protocol).toBe("Wormhole Gateway");
    expect(result.attributes.emitterChain).toBe(21);
  });
});

const cosmosRedeem: CosmosRedeem = {
  blockTimestamp: 1715867714747,
  timestamp: "1717077314747225611",
  chainId: 20,
  events: [
    {
      type: "coin_spent",
      attributes: [
        {
          key: "spender",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv",
          index: true,
        },
      ],
    },
    {
      type: "coin_received",
      attributes: [
        {
          key: "receiver",
          value: "osmo17xpfvakm2amg962yls6f84z3kell8c5lczssa0",
          index: true,
        },
        {
          key: "amount",
          value: "1904uosmo",
          index: true,
        },
      ],
    },
    {
      type: "transfer",
      attributes: [
        {
          key: "recipient",
          value: "osmo17xpfvakm2amg962yls6f84z3kell8c5lczssa0",
          index: true,
        },
        {
          key: "sender",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv",
          index: true,
        },
        {
          key: "amount",
          value: "1904uosmo",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "sender",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv",
          index: true,
        },
      ],
    },
    {
      type: "tx",
      attributes: [
        {
          key: "fee",
          value: "1904uosmo",
          index: true,
        },
      ],
    },
    {
      type: "tx",
      attributes: [
        {
          key: "acc_seq",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv/214524",
          index: true,
        },
      ],
    },
    {
      type: "tx",
      attributes: [
        {
          key: "signature",
          value:
            "ruDMLvHL7KUcfm7XZchYJ9rLg9YI2iAn0yVsbTOmU10oOhL6h18lzFMm1hlfaUyV4nSeZP7BiBY6vIKVP27j1w==",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "action",
          value: "/ibc.core.client.v1.MsgUpdateClient",
          index: true,
        },
        {
          key: "sender",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv",
          index: true,
        },
        {
          key: "msg_index",
          value: "0",
          index: true,
        },
      ],
    },
    {
      type: "update_client",
      attributes: [
        {
          key: "client_id",
          value: "07-tendermint-2927",
          index: true,
        },
        {
          key: "client_type",
          value: "07-tendermint",
          index: true,
        },
        {
          key: "consensus_height",
          value: "0-8439326",
          index: true,
        },
        {
          key: "consensus_heights",
          value: "0-8439326",
          index: true,
        },
        {
          key: "header",
          value:
            "0a262f6962632e6c69676874636c69656e74732e74656e6465726d696e742e76312e48656164657212e7220aa30f0a92030a02080b1209776f726d636861696e189e8c8304220c08c8a098b20610cf9aa2d4012a480a208c6271da18687b3935221270cd3584987edd71f3b6df823e89706f47409653c3122408011220fa150ff46f6868214a21d77de985a9547d9128b44131a1d9a526d69fd357a8903220f9d4145e332228cbb79db79067f5a59d6d711aae9b9540ebe13d36d5c860111c3a20a58d77faec0fd48e86ce061174e2b0c58665b2d432203a18ade4f7a805bafde64220209171fbbd08db5878d988e4c319ea87f2a46c502adfc9e2335a82c8e69911b14a20209171fbbd08db5878d988e4c319ea87f2a46c502adfc9e2335a82c8e69911b15220048091bc7ddc283f77bfbf91d73c44da58c3df8a9cbc867405d8b7f3daada22f5a2068294e4144e84cdf17eab8ad8cbe91bd22c905509f7bab61e4affebb9809b29d6220e0b9af07d57a3885593a9dc814d580e4530aeb9569836699cf084513c65e81546a20e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855721468254e0607098224e14c0788c203c525857ed972128b0c089e8c83041a480a20fa3b6a09d5bb29a6dd75ed63ff4d41f12e4b0c91d1223bcec0df96830c91dd94122408011220baa8d54ca5e810d8528663021f1b20b189b8e5ddf047e74d099f29b27ee2c88a22680802121400af7814014de4dc7b46effd53d68cbee4012d741a0c08c9a098b20610cf9aa2d40122407bb3cb15b2b59990967e7788f552831e48d0436c50e963506f95ac5e5f6a22cce308db9a626156f9bbfb5ca179fcc3e2f31475205e83b5976a7fe385b0c4630d220f08011a0b088092b8c398feffffff0122670802121425ad17ef6dd326e501f41711c2252cb40efc08831a0b08cea098b20610b589a73722404166437d86bc0a856e85714275a31867e712cee42a93a983cfc7bbe827fe32b77cf308e3f84b741b9a1f7a7a250ab82a2a4b7c1eef40b730907107b3f128c20b2267080212142cadd23c1d339b0c13000f47c66e84c9eed583c31a0b08cea098b20610e8bc8c642240af471e18937e9b3a7c0fcb0d17cf39a10c7286b4e370e154404c3fa523b877a8fac78f93a29065d66e0bf83fabe05863a1cd6f2fed73261aff591d732db0220c2267080212142d509655b46c445783d92c4e0bfb82fa447865081a0b08cea098b20610dce8aa42224041a51549bb98ea1507b6b6da210917760418d0b7fe85fbe063bc66055e0600c8041f51558ea4c5561d1d04422f9f41185c23107a116006a26e776a764f00010e2267080212143f3339b051c3cddfe708f40520989775732b74fa1a0b08cea098b20610bfd0a84a224033b638b95294824bea53d6ff3ea27883024b828e3c9d78a4724c58aed25fe202c2b2c9ff3e25251350d6fe7a2bcdfa93531a7d166c0a5e4166e43389e6c5a602220f08011a0b088092b8c398feffffff0122670802121468254e0607098224e14c0788c203c525857ed9721a0b08cea098b20610e3f7ba4b2240bc6a6bb7306d3ecc54aa2033fd54d6f29ccc31ab6aa1f166a33b72e67bf6e0c72e5a798ec08bede3c7403ae034dd7a3af4659936c97770ffaa99c0f635bd0605220f08011a0b088092b8c398feffffff01226708021214856295f68baf4a47c0a77a1da2142819db17aaa21a0b08cea098b20610ef88b740224004bc5924ff724e20f4abe67cac635d485dbfbd77ab842453076c3517cc8498f06fe8d4dbe5270dfcd100a2b7e81e19d7eb7e17c25652d51094a701a5068b91042267080212149637f8eb34a01a190901656ee7a02dabdf2295d21a0b08cea098b20610e4b2b43822400b2f7283af706338572493df47bc967e0b0e21c19d36e097ecf18aa0d3279c9fc0f169531ea86549acacc15db854bee149361e32bebed6e9b4362893cbd1060b2267080212149a220b553072cb391d5ff240433cb4f1d88dfaea1a0b08cea098b20610f8f1863c224008e08eb7fe6824414e0ad003be0c84c6462296e9fb41625a4c15f835a4c23cd9b838ad99eb403ab1185aa1340c8afa61c0a111d39780c55dd82da209318a740c2267080212149fe41ccef2d418cd9ff0cf3aee95c83c26e108a01a0b08cea098b20610fbb3a64a2240724798a8a495d0868b78056a522f1404ccb62d0044f6f9046e2d79398e86652eb6ea11918bb15bd3e9fd691085380f8996e8076a20e89d5113565c81ac5ef50e226708021214bb3529a2fe75854a6e6935d8d5247d63e0d891b81a0b08cea098b206109184943d2240fde174048823db3534b25c51adb073a3fcdd2d63ab4965c9b8922ec77618d54634e254a5fe2120c26d7bcc1cffcabf658e2d8f7254b993e7241f1afd08c36b0f220f08011a0b088092b8c398feffffff01220f08011a0b088092b8c398feffffff01220f08011a0b088092b8c398feffffff01226708021214edb1862a7401d4d782fe6c0803d4a7e4a3d3354b1a0b08cea098b20610e497db5f22407da484e22e4f1b2a0c01adf34a0bcfb37cbc6a33907fd0eac57218318fb430b82a21f80243bd84638fb23cbbb974c00056d30d204986de79989345914366cc0a226708021214fcfd813e8a9f2da8258aa3af958ff96b822fb4831a0b08cea098b20610c7bafe3e2240aa97073465f2039666e3a1c610790fab4964b842e14d5dddc350cc68b0c09f075b9b87bedad9af5fc2e86282c517ce5cec0bb56bf9ba306bab46d7f2ec6bc10e12da090a3c0a1400af7814014de4dc7b46effd53d68cbee4012d7412220a20c4208f8a990fa83716ff961f9051d9d6933c087d4415d412d398b00a4458836618010a3c0a140aecee628c90c9ee2aef51d0809ab0c3c9baf69b12220a20c73fb6f78eb8654752cc67b92a7009bf37e849ad030fd30377488ae3455214ec18010a3c0a1425ad17ef6dd326e501f41711c2252cb40efc088312220a20e8380be9001c813a98ab3801cf0f4c8a5915e34c0d75c17492fdcc763597522618010a3c0a142cadd23c1d339b0c13000f47c66e84c9eed583c312220a200438e655501a35adf0bb736c30a3b602dbb20beaac2c8bbe7cc3f76601e2fd2918010a3c0a142d509655b46c445783d92c4e0bfb82fa4478650812220a20a6cf0e5892917b59115a2e1e069d8827ce161f025a172b765ca1c03292d22b0718010a3c0a143f3339b051c3cddfe708f40520989775732b74fa12220a202502b6f88a2cf754e6bf53d95bc68ae19f4cb00107e0ab9192888fa46866451618010a3c0a144cb3cc41bbbaaafe9363fb00ef7bc1915aa1473c12220a20678feccb0f0e7444d1aac2bdb018bc47746390bc4065d4d55f58e5b4703a908218010a3c0a1468254e0607098224e14c0788c203c525857ed97212220a20eec66e58221355bfd41ae9cc9383588963d2d5ca35f62b60cf06fe1f4aed71c218010a3c0a1479e1e01605ff3f06e31022214e393090f1c86ffe12220a20a110317340f3a0d77918c6130b247d5b8b25c3d3f0622929947132082d66d38118010a3c0a14856295f68baf4a47c0a77a1da2142819db17aaa212220a2026aa1c03bec40759fa5dabeb0a296618af09dd0b0c34286324868dbf5224005618010a3c0a149637f8eb34a01a190901656ee7a02dabdf2295d212220a207cabb403e5ecf1abb094f7fbe4ce7ecceea4e09a4f5c01cfeac9775748d8b29a18010a3c0a149a220b553072cb391d5ff240433cb4f1d88dfaea12220a2011d5f5c6786bd0034a3df910ec7dc56506600b6bab3970ca0975513e17cbb80618010a3c0a149fe41ccef2d418cd9ff0cf3aee95c83c26e108a012220a20856795689c1959451a8addd9e6878055da07c68ecb6837c837952a4e2d6d0f8618010a3c0a14bb3529a2fe75854a6e6935d8d5247d63e0d891b812220a204fe86c557e7611aaebb0bfa63b0bf798bd1bc966b611cb6c1ba5e68a450c162418010a3c0a14d1ac9e6718aa5200d161ec15c2561bf6940a585612220a20e5ad4c79988d0a4a524b974be32c26671ce7114a574d400073e7cc043935437518010a3c0a14d7c8cf17e0a5f1f4bdcea6c045dd14a2b777548912220a209491b95055367dbb34e320544a967762c7743a371142cf2d3a9b954d10f9a1c618010a3c0a14e637616ce859141c82d1b21c805f10402e613db212220a202b437a098ffa56f2ea86698ab01e2dae03b2f62f707baac620422384f128224e18010a3c0a14edb1862a7401d4d782fe6c0803d4a7e4a3d3354b12220a2024da14bf90e7735f7b4dd4be8ebec67b16a8a8a0c232f093831a6e0f07bde30618010a3c0a14fcfd813e8a9f2da8258aa3af958ff96b822fb48312220a20f2d057752995e07e388476ba79e07d8a13a5ea9513a0f7c7c6ffe04753adadc11801123c0a1468254e0607098224e14c0788c203c525857ed97212220a20eec66e58221355bfd41ae9cc9383588963d2d5ca35f62b60cf06fe1f4aed71c2180118131a0510be89830422da090a3c0a1400af7814014de4dc7b46effd53d68cbee4012d7412220a20c4208f8a990fa83716ff961f9051d9d6933c087d4415d412d398b00a4458836618010a3c0a140aecee628c90c9ee2aef51d0809ab0c3c9baf69b12220a20c73fb6f78eb8654752cc67b92a7009bf37e849ad030fd30377488ae3455214ec18010a3c0a1425ad17ef6dd326e501f41711c2252cb40efc088312220a20e8380be9001c813a98ab3801cf0f4c8a5915e34c0d75c17492fdcc763597522618010a3c0a142cadd23c1d339b0c13000f47c66e84c9eed583c312220a200438e655501a35adf0bb736c30a3b602dbb20beaac2c8bbe7cc3f76601e2fd2918010a3c0a142d509655b46c445783d92c4e0bfb82fa4478650812220a20a6cf0e5892917b59115a2e1e069d8827ce161f025a172b765ca1c03292d22b0718010a3c0a143f3339b051c3cddfe708f40520989775732b74fa12220a202502b6f88a2cf754e6bf53d95bc68ae19f4cb00107e0ab9192888fa46866451618010a3c0a144cb3cc41bbbaaafe9363fb00ef7bc1915aa1473c12220a20678feccb0f0e7444d1aac2bdb018bc47746390bc4065d4d55f58e5b4703a908218010a3c0a1468254e0607098224e14c0788c203c525857ed97212220a20eec66e58221355bfd41ae9cc9383588963d2d5ca35f62b60cf06fe1f4aed71c218010a3c0a1479e1e01605ff3f06e31022214e393090f1c86ffe12220a20a110317340f3a0d77918c6130b247d5b8b25c3d3f0622929947132082d66d38118010a3c0a14856295f68baf4a47c0a77a1da2142819db17aaa212220a2026aa1c03bec40759fa5dabeb0a296618af09dd0b0c34286324868dbf5224005618010a3c0a149637f8eb34a01a190901656ee7a02dabdf2295d212220a207cabb403e5ecf1abb094f7fbe4ce7ecceea4e09a4f5c01cfeac9775748d8b29a18010a3c0a149a220b553072cb391d5ff240433cb4f1d88dfaea12220a2011d5f5c6786bd0034a3df910ec7dc56506600b6bab3970ca0975513e17cbb80618010a3c0a149fe41ccef2d418cd9ff0cf3aee95c83c26e108a012220a20856795689c1959451a8addd9e6878055da07c68ecb6837c837952a4e2d6d0f8618010a3c0a14bb3529a2fe75854a6e6935d8d5247d63e0d891b812220a204fe86c557e7611aaebb0bfa63b0bf798bd1bc966b611cb6c1ba5e68a450c162418010a3c0a14d1ac9e6718aa5200d161ec15c2561bf6940a585612220a20e5ad4c79988d0a4a524b974be32c26671ce7114a574d400073e7cc043935437518010a3c0a14d7c8cf17e0a5f1f4bdcea6c045dd14a2b777548912220a209491b95055367dbb34e320544a967762c7743a371142cf2d3a9b954d10f9a1c618010a3c0a14e637616ce859141c82d1b21c805f10402e613db212220a202b437a098ffa56f2ea86698ab01e2dae03b2f62f707baac620422384f128224e18010a3c0a14edb1862a7401d4d782fe6c0803d4a7e4a3d3354b12220a2024da14bf90e7735f7b4dd4be8ebec67b16a8a8a0c232f093831a6e0f07bde30618010a3c0a14fcfd813e8a9f2da8258aa3af958ff96b822fb48312220a20f2d057752995e07e388476ba79e07d8a13a5ea9513a0f7c7c6ffe04753adadc11801123c0a14d7c8cf17e0a5f1f4bdcea6c045dd14a2b777548912220a209491b95055367dbb34e320544a967762c7743a371142cf2d3a9b954d10f9a1c618011813",
          index: true,
        },
        {
          key: "msg_index",
          value: "0",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "module",
          value: "ibc_client",
          index: true,
        },
        {
          key: "msg_index",
          value: "0",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "action",
          value: "/ibc.core.channel.v1.MsgRecvPacket",
          index: true,
        },
        {
          key: "sender",
          value: "osmo1hhzf9u376mg8zcuvx3jsls7t805kzcrsfsaydv",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "recv_packet",
      attributes: [
        {
          key: "packet_data",
          value:
            '{"amount":"60000000","denom":"factory/wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx/GGh9Ufn1SeDGrhzEkMyRKt5568VbbxZK2yvWNsd6PbXt","receiver":"osmo1hm2395wftql4ceh2dkzda07lwutpgkleg6qdgh","sender":"wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx"}',
          index: true,
        },
        {
          key: "packet_data_hex",
          value:
            "7b22616d6f756e74223a223630303030303030222c2264656e6f6d223a22666163746f72792f776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a782f4747683955666e315365444772687a456b4d79524b7435353638566262785a4b327976574e73643650625874222c227265636569766572223a226f736d6f31686d32333935776674716c3463656832646b7a646130376c77757470676b6c65673671646768222c2273656e646572223a22776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a78227d",
          index: true,
        },
        {
          key: "packet_timeout_height",
          value: "0-0",
          index: true,
        },
        {
          key: "packet_timeout_timestamp",
          value: "1717077314747225611",
          index: true,
        },
        {
          key: "packet_sequence",
          value: "37163",
          index: true,
        },
        {
          key: "packet_src_port",
          value: "transfer",
          index: true,
        },
        {
          key: "packet_src_channel",
          value: "channel-3",
          index: true,
        },
        {
          key: "packet_dst_port",
          value: "transfer",
          index: true,
        },
        {
          key: "packet_dst_channel",
          value: "channel-2186",
          index: true,
        },
        {
          key: "packet_channel_ordering",
          value: "ORDER_UNORDERED",
          index: true,
        },
        {
          key: "packet_connection",
          value: "connection-2424",
          index: true,
        },
        {
          key: "connection_id",
          value: "connection-2424",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "module",
          value: "ibc_channel",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "sudo",
      attributes: [
        {
          key: "_contract_address",
          value: "osmo17r7qdw2zk6jyw62cvwm6flmhtj9q7zd26r8zc6sqyf0pnaq46cfss8hgxg",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "wasm",
      attributes: [
        {
          key: "_contract_address",
          value: "osmo17r7qdw2zk6jyw62cvwm6flmhtj9q7zd26r8zc6sqyf0pnaq46cfss8hgxg",
          index: true,
        },
        {
          key: "method",
          value: "try_transfer",
          index: true,
        },
        {
          key: "channel_id",
          value: "channel-2186",
          index: true,
        },
        {
          key: "denom",
          value: "ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "quota",
          value: "none",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "denomination_trace",
      attributes: [
        {
          key: "trace_hash",
          value: "6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "denom",
          value: "ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "coin_received",
      attributes: [
        {
          key: "receiver",
          value: "osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3",
          index: true,
        },
        {
          key: "amount",
          value: "60000000ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "coinbase",
      attributes: [
        {
          key: "minter",
          value: "osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3",
          index: true,
        },
        {
          key: "amount",
          value: "60000000ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "coin_spent",
      attributes: [
        {
          key: "spender",
          value: "osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3",
          index: true,
        },
        {
          key: "amount",
          value: "60000000ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "coin_received",
      attributes: [
        {
          key: "receiver",
          value: "osmo1hm2395wftql4ceh2dkzda07lwutpgkleg6qdgh",
          index: true,
        },
        {
          key: "amount",
          value: "60000000ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "transfer",
      attributes: [
        {
          key: "recipient",
          value: "osmo1hm2395wftql4ceh2dkzda07lwutpgkleg6qdgh",
          index: true,
        },
        {
          key: "sender",
          value: "osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3",
          index: true,
        },
        {
          key: "amount",
          value: "60000000ibc/6B99DB46AA9FF47162148C1726866919E44A6A5E0274B90912FD17E19A337695",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "sender",
          value: "osmo1yl6hdjhmkf37639730gffanpzndzdpmhxy9ep3",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "fungible_token_packet",
      attributes: [
        {
          key: "module",
          value: "transfer",
          index: true,
        },
        {
          key: "sender",
          value: "wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx",
          index: true,
        },
        {
          key: "receiver",
          value: "osmo1hm2395wftql4ceh2dkzda07lwutpgkleg6qdgh",
          index: true,
        },
        {
          key: "denom",
          value:
            "factory/wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx/GGh9Ufn1SeDGrhzEkMyRKt5568VbbxZK2yvWNsd6PbXt",
          index: true,
        },
        {
          key: "amount",
          value: "60000000",
          index: true,
        },
        {
          key: "memo",
          value: "",
          index: true,
        },
        {
          key: "success",
          value: "true",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "write_acknowledgement",
      attributes: [
        {
          key: "packet_data",
          value:
            '{"amount":"60000000","denom":"factory/wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx/GGh9Ufn1SeDGrhzEkMyRKt5568VbbxZK2yvWNsd6PbXt","receiver":"osmo1hm2395wftql4ceh2dkzda07lwutpgkleg6qdgh","sender":"wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx"}',
          index: true,
        },
        {
          key: "packet_data_hex",
          value:
            "7b22616d6f756e74223a223630303030303030222c2264656e6f6d223a22666163746f72792f776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a782f4747683955666e315365444772687a456b4d79524b7435353638566262785a4b327976574e73643650625874222c227265636569766572223a226f736d6f31686d32333935776674716c3463656832646b7a646130376c77757470676b6c65673671646768222c2273656e646572223a22776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a78227d",
          index: true,
        },
        {
          key: "packet_timeout_height",
          value: "0-0",
          index: true,
        },
        {
          key: "packet_timeout_timestamp",
          value: "1717077314747225611",
          index: true,
        },
        {
          key: "packet_sequence",
          value: "37163",
          index: true,
        },
        {
          key: "packet_src_port",
          value: "transfer",
          index: true,
        },
        {
          key: "packet_src_channel",
          value: "channel-3",
          index: true,
        },
        {
          key: "packet_dst_port",
          value: "transfer",
          index: true,
        },
        {
          key: "packet_dst_channel",
          value: "channel-2186",
          index: true,
        },
        {
          key: "packet_ack",
          value: '{"result":"AQ=="}',
          index: true,
        },
        {
          key: "packet_ack_hex",
          value: "7b22726573756c74223a2241513d3d227d",
          index: true,
        },
        {
          key: "packet_connection",
          value: "connection-2424",
          index: true,
        },
        {
          key: "connection_id",
          value: "connection-2424",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
    {
      type: "message",
      attributes: [
        {
          key: "module",
          value: "ibc_channel",
          index: true,
        },
        {
          key: "msg_index",
          value: "1",
          index: true,
        },
      ],
    },
  ],
  height: "15778340",
  data: "Ei0KKy9pYmMuY29yZS5jbGllbnQudjEuTXNnVXBkYXRlQ2xpZW50UmVzcG9uc2USMAoqL2liYy5jb3JlLmNoYW5uZWwudjEuTXNnUmVjdlBhY2tldFJlc3BvbnNlEgIIAg==",
  hash: "C196E9E445748AB4BE26E980F685F8F1FD02E8F327F9F1929CE5C426C936BF74",
  tx: Buffer.from([
    10, 245, 13, 10, 242, 13, 10, 36, 47, 99, 111, 115, 109, 119, 97, 115, 109, 46, 119, 97, 115,
    109, 46, 118, 49, 46, 77, 115, 103, 69, 120, 101, 99, 117, 116, 101, 67, 111, 110, 116, 114, 97,
    99, 116, 18, 201, 13, 10, 47, 119, 111, 114, 109, 104, 111, 108, 101, 49, 106, 121, 120, 99, 50,
    56, 102, 56, 109, 48, 104, 55, 117, 122, 53, 122, 119, 113, 120, 99, 48, 115, 54, 112, 53, 122,
    97, 55, 101, 117, 51, 48, 120, 110, 114, 109, 113, 56, 18, 67, 119, 111, 114, 109, 104, 111,
    108, 101, 49, 52, 101, 106, 113, 106, 121, 113, 56, 117, 109, 52, 112, 51, 120, 102, 113, 106,
    55, 52, 121, 108, 100, 53, 119, 97, 113, 108, 106, 102, 56, 56, 102, 122, 50, 53, 121, 120, 110,
    109, 97, 48, 99, 110, 103, 115, 112, 120, 101, 51, 108, 101, 115, 48, 48, 102, 112, 106, 120,
    26, 208, 12, 123, 34, 99, 111, 109, 112, 108, 101, 116, 101, 95, 116, 114, 97, 110, 115, 102,
    101, 114, 95, 97, 110, 100, 95, 99, 111, 110, 118, 101, 114, 116, 34, 58, 123, 34, 118, 97, 97,
    34, 58, 34, 65, 81, 65, 65, 65, 65, 81, 78, 65, 77, 104, 53, 77, 114, 106, 68, 47, 119, 84, 120,
    77, 84, 104, 48, 82, 54, 77, 103, 115, 81, 71, 120, 102, 83, 101, 50, 70, 120, 86, 71, 110, 82,
    82, 108, 80, 50, 89, 80, 116, 82, 120, 97, 85, 77, 68, 53, 85, 120, 67, 110, 97, 99, 109, 75,
    82, 114, 80, 110, 54, 98, 103, 100, 104, 69, 104, 109, 87, 108, 53, 79, 84, 83, 48, 89, 55, 121,
    119, 110, 122, 49, 70, 113, 117, 104, 99, 65, 65, 81, 118, 82, 69, 89, 57, 53, 116, 112, 67, 89,
    110, 55, 77, 74, 97, 51, 97, 69, 86, 68, 112, 88, 110, 100, 97, 81, 89, 80, 78, 52, 57, 114, 73,
    115, 105, 57, 56, 113, 56, 109, 84, 118, 70, 78, 67, 75, 104, 73, 118, 72, 84, 114, 113, 112,
    67, 119, 76, 103, 47, 48, 71, 104, 103, 68, 111, 72, 113, 66, 87, 114, 85, 66, 80, 89, 73, 101,
    74, 79, 69, 99, 120, 116, 77, 72, 119, 66, 65, 117, 114, 120, 49, 48, 54, 57, 106, 47, 119, 82,
    85, 81, 70, 77, 98, 99, 101, 115, 90, 47, 105, 74, 48, 67, 88, 89, 86, 79, 104, 82, 118, 74, 66,
    89, 102, 87, 66, 75, 90, 79, 87, 119, 83, 101, 88, 116, 108, 50, 43, 113, 104, 113, 120, 122,
    89, 103, 75, 103, 65, 110, 53, 73, 84, 114, 79, 50, 102, 75, 77, 107, 65, 57, 120, 66, 108, 104,
    52, 86, 106, 88, 84, 57, 71, 67, 69, 65, 65, 48, 116, 47, 119, 67, 118, 70, 115, 67, 74, 53, 79,
    49, 110, 48, 48, 102, 76, 98, 115, 122, 80, 81, 81, 107, 118, 70, 112, 87, 89, 83, 100, 56, 81,
    67, 99, 102, 114, 70, 109, 116, 106, 101, 89, 86, 99, 50, 99, 73, 90, 82, 84, 77, 116, 55, 43,
    111, 66, 75, 49, 114, 65, 121, 109, 121, 97, 48, 82, 118, 112, 79, 114, 122, 77, 83, 51, 119,
    66, 50, 52, 86, 88, 118, 55, 117, 73, 66, 66, 80, 118, 90, 55, 65, 108, 98, 114, 78, 78, 68,
    111, 72, 43, 114, 78, 102, 115, 114, 121, 90, 71, 49, 73, 74, 48, 103, 118, 68, 65, 122, 82, 72,
    109, 51, 78, 101, 48, 48, 107, 47, 74, 85, 80, 111, 83, 101, 78, 76, 111, 48, 81, 114, 72, 122,
    103, 87, 69, 68, 66, 50, 67, 109, 78, 51, 79, 52, 47, 71, 106, 103, 83, 69, 102, 79, 101, 107,
    108, 117, 56, 53, 84, 53, 66, 115, 77, 66, 66, 113, 109, 122, 105, 69, 56, 98, 97, 89, 74, 113,
    114, 89, 119, 122, 115, 99, 79, 119, 113, 81, 67, 105, 120, 89, 68, 85, 106, 110, 81, 54, 89,
    122, 103, 104, 57, 101, 69, 120, 83, 104, 103, 81, 74, 114, 52, 71, 65, 51, 108, 56, 73, 54,
    121, 68, 84, 98, 119, 101, 72, 68, 119, 65, 78, 69, 114, 66, 55, 56, 90, 43, 89, 113, 90, 87,
    89, 89, 82, 50, 90, 100, 89, 100, 117, 80, 103, 65, 67, 103, 103, 55, 70, 78, 84, 87, 116, 99,
    75, 90, 54, 50, 116, 90, 74, 103, 79, 71, 121, 114, 105, 101, 107, 97, 75, 98, 47, 119, 112, 48,
    43, 72, 108, 43, 57, 56, 117, 79, 105, 119, 68, 66, 88, 53, 104, 78, 115, 120, 97, 89, 103, 85,
    117, 81, 71, 79, 118, 110, 121, 111, 87, 43, 49, 80, 81, 104, 53, 121, 65, 65, 101, 74, 51, 120,
    47, 77, 48, 102, 86, 89, 81, 55, 74, 121, 85, 66, 68, 79, 51, 117, 113, 114, 71, 121, 52, 81,
    76, 68, 114, 87, 114, 104, 53, 119, 85, 54, 54, 85, 119, 98, 77, 43, 50, 97, 105, 106, 54, 69,
    76, 106, 48, 65, 51, 69, 67, 115, 86, 100, 73, 117, 82, 70, 110, 112, 56, 122, 122, 112, 55, 86,
    80, 54, 99, 57, 116, 98, 122, 106, 105, 118, 68, 43, 83, 87, 88, 100, 97, 116, 115, 68, 104, 51,
    101, 75, 87, 55, 47, 90, 83, 115, 118, 53, 56, 66, 68, 112, 87, 55, 109, 100, 114, 82, 109, 105,
    69, 116, 49, 113, 98, 97, 112, 52, 113, 106, 67, 77, 107, 101, 79, 110, 51, 70, 88, 84, 110, 70,
    117, 73, 112, 49, 84, 55, 104, 68, 109, 121, 79, 119, 88, 55, 73, 52, 80, 83, 117, 102, 100,
    109, 67, 50, 99, 71, 112, 114, 85, 111, 101, 110, 119, 66, 114, 54, 86, 122, 120, 118, 86, 89,
    113, 106, 112, 109, 53, 65, 101, 102, 104, 74, 88, 78, 111, 65, 68, 120, 108, 107, 56, 120, 47,
    77, 87, 48, 99, 71, 68, 100, 98, 121, 81, 101, 85, 110, 89, 102, 117, 117, 87, 118, 110, 48,
    111, 55, 119, 78, 51, 73, 110, 54, 121, 98, 89, 70, 43, 117, 77, 66, 82, 105, 78, 71, 53, 65,
    112, 81, 82, 105, 105, 78, 51, 72, 90, 74, 56, 54, 70, 122, 84, 116, 98, 85, 98, 77, 74, 111,
    75, 107, 51, 111, 100, 68, 76, 51, 43, 76, 73, 102, 56, 48, 111, 65, 69, 76, 107, 50, 71, 105,
    89, 115, 68, 72, 86, 90, 102, 48, 82, 103, 70, 98, 109, 87, 56, 78, 118, 74, 56, 89, 50, 71, 82,
    103, 67, 87, 105, 99, 101, 90, 72, 89, 82, 69, 99, 107, 86, 82, 82, 52, 107, 70, 122, 107, 65,
    56, 101, 122, 98, 120, 48, 98, 100, 73, 97, 87, 87, 69, 79, 69, 118, 101, 73, 105, 110, 76, 120,
    57, 70, 85, 105, 114, 53, 76, 98, 116, 100, 51, 56, 67, 65, 66, 69, 83, 116, 113, 117, 54, 100,
    77, 115, 106, 71, 82, 67, 87, 68, 122, 87, 79, 56, 102, 54, 99, 43, 90, 101, 109, 99, 108, 99,
    55, 79, 65, 71, 83, 121, 117, 67, 43, 101, 70, 99, 72, 79, 100, 101, 121, 90, 106, 68, 118, 78,
    103, 98, 76, 99, 106, 106, 97, 111, 50, 49, 53, 98, 106, 49, 77, 105, 89, 88, 87, 72, 47, 80,
    120, 90, 87, 104, 103, 103, 74, 102, 83, 57, 86, 69, 73, 73, 65, 69, 107, 84, 113, 119, 76, 82,
    73, 121, 103, 43, 52, 117, 51, 106, 115, 47, 47, 79, 67, 49, 77, 117, 110, 100, 105, 74, 103,
    54, 90, 79, 67, 86, 43, 97, 48, 53, 77, 105, 51, 77, 54, 81, 89, 101, 47, 48, 83, 86, 115, 103,
    50, 85, 107, 51, 113, 83, 68, 78, 113, 105, 78, 79, 56, 78, 86, 112, 107, 106, 88, 74, 54, 86,
    54, 76, 119, 119, 56, 84, 70, 49, 66, 112, 51, 104, 72, 107, 66, 90, 107, 89, 81, 80, 65, 65,
    66, 72, 107, 89, 65, 70, 99, 122, 79, 54, 121, 107, 48, 106, 51, 71, 57, 48, 105, 47, 43, 57,
    68, 111, 113, 71, 99, 72, 49, 116, 101, 70, 56, 88, 77, 112, 85, 69, 86, 75, 82, 73, 66, 103,
    109, 99, 113, 51, 108, 65, 65, 65, 65, 65, 65, 65, 66, 57, 117, 56, 65, 65, 119, 65, 65, 65, 65,
    65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65,
    65, 65, 65, 65, 65, 65, 65, 65, 65, 68, 107, 52, 99, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65,
    65, 65, 65, 65, 65, 65, 111, 76, 104, 112, 107, 99, 89, 104, 105, 122, 98, 66, 48, 90, 49, 75,
    76, 112, 54, 119, 122, 106, 89, 71, 54, 48, 103, 65, 65, 113, 53, 107, 67, 82, 65, 72, 53, 117,
    111, 89, 109, 83, 67, 88, 113, 107, 43, 50, 106, 117, 103, 47, 74, 74, 122, 112, 69, 113, 104,
    68, 84, 51, 49, 43, 74, 111, 103, 69, 50, 89, 47, 122, 68, 67, 68, 119, 70, 89, 72, 82, 87, 113,
    120, 53, 75, 105, 55, 57, 48, 118, 74, 98, 102, 117, 117, 120, 70, 104, 48, 110, 90, 67, 113,
    111, 97, 90, 107, 79, 100, 106, 57, 87, 67, 73, 57, 121, 84, 110, 115, 105, 90, 50, 70, 48, 90,
    88, 100, 104, 101, 86, 57, 48, 99, 109, 70, 117, 99, 50, 90, 108, 99, 105, 73, 54, 101, 121, 74,
    106, 97, 71, 70, 112, 98, 105, 73, 54, 77, 106, 65, 115, 73, 109, 53, 118, 98, 109, 78, 108, 73,
    106, 111, 53, 78, 84, 77, 120, 76, 67, 74, 121, 90, 87, 78, 112, 99, 71, 108, 108, 98, 110, 81,
    105, 79, 105, 74, 105, 77, 48, 53, 48, 89, 110, 112, 71, 98, 50, 74, 85, 83, 88, 112, 80, 86,
    70, 89, 122, 87, 109, 53, 83, 101, 71, 74, 69, 85, 109, 112, 97, 86, 50, 100, 53, 87, 107, 100,
    48, 78, 108, 112, 72, 82, 88, 100, 79, 77, 110, 103, 122, 90, 70, 104, 83, 100, 49, 111, 121,
    100, 72, 78, 97, 86, 50, 77, 121, 89, 49, 100, 83, 98, 109, 70, 66, 80, 84, 48, 105, 76, 67, 74,
    109, 90, 87, 85, 105, 79, 105, 73, 119, 73, 110, 49, 57, 34, 125, 125, 18, 90, 10, 82, 10, 70,
    10, 31, 47, 99, 111, 115, 109, 111, 115, 46, 99, 114, 121, 112, 116, 111, 46, 115, 101, 99, 112,
    50, 53, 54, 107, 49, 46, 80, 117, 98, 75, 101, 121, 18, 35, 10, 33, 3, 223, 240, 89, 118, 55,
    33, 98, 82, 238, 45, 84, 70, 47, 145, 108, 189, 14, 216, 222, 76, 243, 221, 121, 34, 206, 59,
    86, 139, 56, 105, 156, 101, 18, 4, 10, 2, 8, 1, 24, 219, 167, 2, 18, 4, 16, 128, 137, 122, 26,
    64, 75, 65, 114, 145, 18, 98, 172, 128, 115, 22, 187, 66, 35, 103, 34, 10, 76, 132, 196, 147,
    101, 230, 218, 112, 169, 58, 154, 248, 144, 72, 236, 235, 93, 69, 50, 163, 47, 134, 128, 57, 69,
    143, 127, 181, 116, 243, 72, 88, 181, 67, 43, 217, 141, 119, 157, 208, 136, 157, 142, 183, 69,
    28, 242, 54,
  ]),
};
