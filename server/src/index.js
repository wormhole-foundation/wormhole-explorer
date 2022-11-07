const express = require("express");
const app = express();
const port = 4000;

const { MongoClient } = require("mongodb");
const mongoURI = process.env.MONGODB_URI;
if (!mongoURI) {
  console.error("You must set your 'MONGODB_URI' environmental variable.");
  process.exit(1);
}
const mongoClient = new MongoClient(mongoURI);

/*
 *  Helpers
 */

async function paginatedFind(collection, req, filter) {
  const limit =
    req.query?.limit && req.query.limit <= 100 ? req.query.limit : 20;
  const skip = req.query?.page ? req.query?.page * limit : undefined;
  const query = req.query?.before
    ? { ...filter, createdAt: { $lt: new Date(req.query.before) } }
    : filter;
  const cursor = await collection.find(query, {
    sort: { createdAt: -1 },
    skip,
    limit,
  });
  return cursor;
}

async function findAndSendMany(
  db,
  res,
  collectionName,
  reqForPagination,
  filter,
  project
) {
  const database = mongoClient.db(db);
  const collection = database.collection(collectionName);
  const cursor = await (reqForPagination
    ? paginatedFind(collection, reqForPagination, filter)
    : collection.find(filter).project(project));
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
}

async function findAndSendOne(db, res, collectionName, filter, project) {
  const database = mongoClient.db(db);
  const collection = database.collection(collectionName);
  const result = await collection.findOne(filter, project);
  if (!result) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
}

/*
 *  Heartbeats
 */

app.get("/api/heartbeats", async (req, res) => {
  await findAndSendMany("wormhole", res, "heartbeats");
});

/*
 *  VAAs
 */

app.get("/api/vaas", async (req, res) => {
  await findAndSendMany("wormhole", res, "vaas", req);
});

app.get("/api/vaas/:chain", async (req, res) => {
  await findAndSendMany("wormhole", res, "vaas", req, {
    _id: { $regex: `^${req.params.chain}/.*` },
  });
});

app.get("/api/vaas/:chain/:emitter", async (req, res) => {
  await findAndSendMany("wormhole", res, "vaas", req, {
    _id: { $regex: `^${req.params.chain}/${req.params.emitter}/.*` },
  });
});

app.get("/api/vaas/:chain/:emitter/:sequence", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}`;
  await findAndSendOne("wormhole", res, "vaas", { _id: id });
});

app.get("/api/vaas-sans-pythnet", async (req, res) => {
  await findAndSendMany("wormhole", res, "vaas", req, {
    _id: { $not: { $regex: `^26/.*` } },
  });
});

app.get("/api/vaa-counts", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("vaas");
  const cursor = await collection.aggregate([
    {
      $bucket: {
        groupBy: "$_id",
        boundaries: [
          "1/",
          "10/",
          "11/",
          "12/",
          "13/",
          "14/",
          "15/",
          "16/",
          "18/",
          "2/",
          "26/",
          "3/",
          "4/",
          "5/",
          "6/",
          "7/",
          "8/",
          "9/",
        ],
        default: "unknown",
        output: {
          count: { $sum: 1 },
        },
      },
    },
  ]);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

/*
 *  Observations
 */

app.get("/api/observations", async (req, res) => {
  await findAndSendMany("wormhole", res, "observations", req);
});

app.get("/api/observations/:chain", async (req, res) => {
  await findAndSendMany("wormhole", res, "observations", req, {
    _id: { $regex: `^${req.params.chain}/.*` },
  });
});

app.get("/api/observations/:chain/:emitter", async (req, res) => {
  await findAndSendMany("wormhole", res, "observations", req, {
    _id: { $regex: `^${req.params.chain}/${req.params.emitter}/.*` },
  });
});

app.get("/api/observations/:chain/:emitter/:sequence", async (req, res) => {
  await findAndSendMany("wormhole", res, "observations", req, {
    _id: {
      $regex: `^${req.params.chain}/${req.params.emitter}/${req.params.sequence}/.*`,
    },
  });
});

app.get(
  "/api/observations/:chain/:emitter/:sequence/:signer/:hash",
  async (req, res) => {
    const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}/${req.params.signer}/${req.params.hash}`;
    await findAndSendOne("wormhole", res, "observations", { _id: id });
  }
);

/*
 *  GovernorConfig
 */
app.get("/api/governorConfig", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorCfgs");
  const cursor = await collection.find({}).project({
    createdAt: 1,
    updatedAt: 1,
    nodename: "$parsedConfig.nodename", //<-- rename fields to flatten
    counter: "$parsedConfig.counter",
    chains: "$parsedConfig.chains",
    tokens: "$parsedConfig.tokens",
  });
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/governorConfig/:guardianaddr", async (req, res) => {
  const id = `${req.params.guardianaddr}`;
  await findAndSendOne(
    "wormhole",
    res,
    "governorCfgs",
    {
      _id: id,
    },
    {
      projection: {
        createdAt: 1,
        updatedAt: 1,
        nodename: "$parsedConfig.nodename", //<-- rename fields to flatten
        counter: "$parsedConfig.counter",
        chains: "$parsedConfig.chains",
        tokens: "$parsedConfig.tokens",
      },
    }
  );
});

app.get("/api/governorLimits", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorCfgs");
  const cursor = await collection.aggregate([
    {
      $lookup: {
        from: "governorStatus",
        localField: "_id",
        foreignField: "_id",
        as: "status",
      },
    },
    {
      $unwind: "$status",
    },
    {
      $project: {
        configChains: "$parsedConfig.chains",
        statusChains: "$status.parsedStatus.chains",
      },
    },
    {
      $unwind: "$configChains",
    },
    {
      $unwind: "$statusChains",
    },
    {
      $match: {
        $expr: { $eq: ["$configChains.chainid", "$statusChains.chainid"] },
      },
    },
    {
      $sort: {
        "configChains.chainid": 1,
      },
    },
    {
      $group: {
        _id: "$configChains.chainid",
        notionalLimits: {
          $push: {
            notionalLimit: "$configChains.notionallimit",
            maxTransactionSize: "$configChains.bigtransactionsize",
            availableNotional: "$statusChains.remainingavailablenotional",
          },
        },
      },
    },
    {
      $project: {
        chainId: "$_id",
        notionalLimits: 1,
      },
    },
  ]);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  const minGuardianNum = 13;
  var agg = [];
  result.forEach((chain) => {
    const sortedAvailableNotionals = chain.notionalLimits.sort(function (a, b) {
      return parseInt(b.availableNotional) - parseInt(a.availableNotional);
    });
    const sortedNotionalLimits = chain.notionalLimits.sort(function (a, b) {
      return parseInt(b.notionalLimit) - parseInt(a.notionalLimit);
    });

    const sortedMaxTransactionSize = chain.notionalLimits.sort(function (a, b) {
      return parseInt(b.maxTransactionSize) - parseInt(a.maxTransactionSize);
    });
    agg.push({
      chainId: chain.chainId,
      availableNotional:
        sortedAvailableNotionals[minGuardianNum - 1]?.availableNotional || null,
      notionalLimit:
        sortedNotionalLimits[minGuardianNum - 1]?.notionalLimit || null,
      maxTransactionSize:
        sortedMaxTransactionSize[minGuardianNum - 1]?.maxTransactionSize ||
        null,
    });
  });
  res.send(
    agg.sort(function (a, b) {
      return parseInt(a.chainId) - parseInt(b.chainId);
    })
  );
});

app.get("/api/notionalLimits", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorCfgs");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        chains: "$parsedConfig.chains",
      },
    },
    {
      $unwind: "$chains",
    },
    {
      $sort: {
        "chains.chainid": 1,
        "chains.notionallimit": -1,
        "chains.bigtransactionsize": -1,
      },
    },
    {
      $group: {
        _id: "$chains.chainid",
        notionalLimits: {
          $push: {
            notionalLimit: "$chains.notionallimit",
            maxTransactionSize: "$chains.bigtransactionsize",
          },
        },
      },
    },
    {
      $project: {
        chainId: "$_id",
        notionalLimits: 1,
      },
    },
  ]);
  const result = await cursor.toArray();
  console.log(result);
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  const minGuardianNum = 13;
  var agg = [];
  result.forEach((chain) => {
    agg.push({
      chainId: chain.chainId,
      notionalLimit:
        chain.notionalLimits[minGuardianNum - 1]?.notionalLimit || null,
      maxTransactionSize:
        chain.notionalLimits[minGuardianNum - 1]?.maxTransactionSize || null,
    });
  });
  res.send(
    agg.sort(function (a, b) {
      return parseInt(a.chainId) - parseInt(b.chainId);
    })
  );
});

app.get("/api/notionalLimits/:chainNum", async (req, res) => {
  const id = `${req.params.chainNum}`;
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorCfgs");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: "$parsedConfig.nodename",
        "parsedConfig.chains": {
          $filter: {
            input: "$parsedConfig.chains",
            as: "chain",
            cond: { $eq: [`$$chain.chainid`, parseInt(id)] },
          },
        },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        notionalLimits: { $arrayElemAt: ["$parsedConfig.chains", 0] },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        chainId: "$notionalLimits.chainid",
        notionalLimit: "$notionalLimits.notionallimit",
        maxTransactionSize: "$notionalLimits.bigtransactionsize",
      },
    },
  ]);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

/*
 *  GovernorStatus
 */
app.get("/api/governorStatus", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.find({}).project({
    createdAt: 1,
    updatedAt: 1,
    nodename: "$parsedStatus.nodename", //<-- rename fields to flatten
    chains: "$parsedStatus.chains",
  });
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/governorStatus/:guardianaddr", async (req, res) => {
  const id = `${req.params.guardianaddr}`;
  await findAndSendOne(
    "wormhole",
    res,
    "governorStatus",
    {
      _id: id,
    },
    {
      projection: {
        createdAt: 1,
        updatedAt: 1,
        nodename: "$parsedStatus.nodename",
        chains: "$parsedStatus.chains",
      },
    }
  );
});

app.get("/api/availableNotional", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        chains: "$parsedStatus.chains",
      },
    },
    {
      $unwind: "$chains",
    },
    {
      $sort: {
        "chains.chainid": 1,
        "chains.remainingavailablenotional": -1,
      },
    },
    {
      $group: {
        _id: "$chains.chainid",
        availableNotionals: {
          $push: {
            availableNotional: "$chains.remainingavailablenotional",
          },
        },
      },
    },
    {
      $project: {
        chainId: "$_id",
        availableNotionals: 1,
      },
    },
  ]);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  const minGuardianNum = 13;
  var agg = [];
  result.forEach((chain) => {
    agg.push({
      chainId: chain.chainId,
      availableNotional:
        chain.availableNotionals[minGuardianNum - 1]?.availableNotional || null,
    });
  });
  res.send(
    agg.sort(function (a, b) {
      return parseInt(a.chainId) - parseInt(b.chainId);
    })
  );
});

app.get("/api/availableNotional/:chainNum", async (req, res) => {
  const id = `${req.params.chainNum}`;
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: "$parsedStatus.nodename",
        "parsedStatus.chains": {
          $filter: {
            input: "$parsedStatus.chains",
            as: "chain",
            cond: { $eq: [`$$chain.chainid`, parseInt(id)] },
          },
        },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        availableNotional: { $arrayElemAt: ["$parsedStatus.chains", 0] },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        chainId: "$availableNotional.chainid",
        availableNotional: "$availableNotional.remainingavailablenotional",
      },
    },
  ]);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/maxAvailableNotional/:chainNum", async (req, res) => {
  const id = `${req.params.chainNum}`;
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: "$parsedStatus.nodename",
        "parsedStatus.chains": {
          $filter: {
            input: "$parsedStatus.chains",
            as: "chain",
            cond: { $eq: [`$$chain.chainid`, parseInt(id)] },
          },
        },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        availableNotional: { $arrayElemAt: ["$parsedStatus.chains", 0] },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        chainId: "$availableNotional.chainid",
        availableNotional: "$availableNotional.remainingavailablenotional",
        emitters: "$availableNotional.emitters",
      },
    },
  ]);
  const result = await cursor.toArray();
  const sortedResult = result.sort(function (b, a) {
    return parseInt(a.availableNotional) - parseInt(b.availableNotional);
  });
  if (sortedResult.length === 0) {
    res.sendStatus(404);
    return;
  }
  const minGuardianNum = 13;
  res.send(sortedResult[minGuardianNum - 1]);
});

app.get("/api/enqueuedVaas", async (req, res) => {
  const id = `${req.params.chainNum}`;
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        chains: "$parsedStatus.chains",
      },
    },
    {
      $unwind: "$chains",
    },
    {
      $project: {
        _id: 1,
        chainId: "$chains.chainid",
        emitters: "$chains.emitters",
      },
    },
    {
      $group: {
        _id: "$chainId",
        emitters: {
          $push: {
            emitterAddress: { $arrayElemAt: ["$emitters.emitteraddress", 0] },
            enqueuedVaas: { $arrayElemAt: ["$emitters.enqueuedvaas", 0] },
          },
        },
      },
    },
  ]);
  const result = await cursor.toArray();
  var filteredResult = [];
  var keys = [];
  result.forEach((res) => {
    const chainId = res._id;
    const emitters = res.emitters;
    emitters.forEach((emitter) => {
      const emitterAddress = emitter.emitterAddress;
      const enqueuedVaas = emitter.enqueuedVaas;
      if (enqueuedVaas != null) {
        enqueuedVaas.forEach((vaa) => {
          //add to dictionary
          const key = `${emitterAddress}/${vaa.sequence}/${vaa.txhash}`;
          if (!keys.includes(key)) {
            filteredResult.push({
              chainId: chainId,
              emitterAddress: emitterAddress,
              sequence: vaa.sequence,
              notionalValue: vaa.notionalvalue,
              txHash: vaa.txhash,
            });
            keys.push(key);
          }
        });
      }
    });
  });

  if (filteredResult.length === 0) {
    res.sendStatus(404);
    return;
  }

  const groups = filteredResult.reduce((groups, item) => {
    const group = groups[item.chainId] || [];
    group.push(item);
    groups[item.chainId] = group;
    return groups;
  }, {});
  const modifiedResult = [];
  for (const [key, value] of Object.entries(groups)) {
    modifiedResult.push({ chainId: key, enqueuedVaas: value });
  }
  res.send(modifiedResult);
});

app.get("/api/enqueuedVaas/:chainNum", async (req, res) => {
  // returns unique enqueued vaas for chainNum
  const id = `${req.params.chainNum}`;
  const database = mongoClient.db("wormhole");
  const collection = database.collection("governorStatus");
  const cursor = await collection.aggregate([
    {
      $match: {},
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: "$parsedStatus.nodename",
        "parsedStatus.chains": {
          $filter: {
            input: "$parsedStatus.chains",
            as: "chain",
            cond: { $eq: [`$$chain.chainid`, parseInt(id)] },
          },
        },
      },
    },
    {
      $project: {
        _id: 1,
        createdAt: 1,
        updatedAt: 1,
        nodeName: 1,
        emitters: "$parsedStatus.chains.emitters",
      },
    },
    {
      $unwind: "$emitters",
    },
    {
      $group: {
        _id: { $arrayElemAt: ["$emitters.emitteraddress", 0] },
        enqueuedVaas: {
          $push: {
            enqueuedVaa: "$emitters.enqueuedvaas",
          },
        },
      },
    },
  ]);
  const result = await cursor.toArray();
  var filteredResult = [];
  var keys = [];
  result.forEach((res) => {
    const emitterAddress = res._id;
    const enqueuedVaas = res.enqueuedVaas;
    enqueuedVaas.forEach((vaa) => {
      const enqueuedVaa = vaa.enqueuedVaa;
      enqueuedVaa.forEach((eV) => {
        if (eV != null) {
          eV.forEach((ev) => {
            if (ev != null) {
              //add to dictionary
              const key = `${emitterAddress}/${ev.sequence}/${ev.txhash}`;
              if (!keys.includes(key)) {
                filteredResult.push({
                  chainId: id,
                  emitterAddress: emitterAddress,
                  sequence: ev.sequence,
                  notionalValue: ev.notionalvalue,
                  txHash: ev.txhash,
                  releaseTime: ev.releasetime,
                });

                keys.push(key);
              }
            }
          });
        }
      });
    });
  });
  if (filteredResult.length === 0) {
    res.sendStatus(404);
    return;
  }

  const sortedResult = filteredResult.sort(function (a, b) {
    return parseInt(a.sequence) - parseInt(b.sequence);
  });
  res.send(sortedResult);
});

/*
 *  Custody
 */

app.get("/api/custody", async (req, res) => {
  await findAndSendMany("onchain_data", res, "custody", req);
});

app.get("/api/custody/:chain/:emitter", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}`;
  await findAndSendOne(
    "onchain_data",
    res,
    "custody",
    {
      _id: id,
    },
    {}
  );
});

app.get("/api/custody/tokens", async (req, res) => {
  await findAndSendMany("onchain_data", res, "custody", req, {}, { tokens: 1 });
});

app.get("/api/custody/tokens/:chain/:emitter", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}`;
  await findAndSendOne(
    "onchain_data",
    res,
    "custody",
    {
      _id: id,
    },
    { projection: { tokens: 1 } }
  );
});

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`);
});
