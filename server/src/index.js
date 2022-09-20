const express = require("express");
const app = express();
const port = 4001;

const { MongoClient } = require("mongodb");
const mongoURI = process.env.MONGODB_URI;
if (!mongoURI) {
  console.error("You must set your 'MONGODB_URI' environmental variable.");
  process.exit(1);
}
const mongoClient = new MongoClient(mongoURI);

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

async function findAndSendMany(res, collectionName, reqForPagination, filter) {
  const database = mongoClient.db("wormhole");
  const collection = database.collection(collectionName);
  const cursor = await (reqForPagination
    ? paginatedFind(collection, reqForPagination, filter)
    : collection.find(filter));
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
}

async function findAndSendOne(res, collectionName, filter) {
  const database = mongoClient.db("wormhole");
  const collection = database.collection(collectionName);
  const result = await collection.findOne(filter);
  if (!result) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
}

app.get("/api/heartbeats", async (req, res) => {
  await findAndSendMany(res, "heartbeats");
});

app.get("/api/vaas", async (req, res) => {
  await findAndSendMany(res, "vaas", req);
});

app.get("/api/vaas/:chain", async (req, res) => {
  await findAndSendMany(res, "vaas", req, {
    _id: { $regex: `^${req.params.chain}/.*` },
  });
});

app.get("/api/vaas/:chain/:emitter", async (req, res) => {
  await findAndSendMany(res, "vaas", req, {
    _id: { $regex: `^${req.params.chain}/${req.params.emitter}/.*` },
  });
});

app.get("/api/vaas/:chain/:emitter/:sequence", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}`;
  await findAndSendOne(res, "vaas", { _id: id });
});

app.get("/api/observations", async (req, res) => {
  await findAndSendMany(res, "observations", req);
});

app.get("/api/observations/:chain", async (req, res) => {
  await findAndSendMany(res, "observations", req, {
    _id: { $regex: `^${req.params.chain}/.*` },
  });
});

app.get("/api/observations/:chain/:emitter", async (req, res) => {
  await findAndSendMany(res, "observations", req, {
    _id: { $regex: `^${req.params.chain}/${req.params.emitter}/.*` },
  });
});

app.get("/api/observations/:chain/:emitter/:sequence", async (req, res) => {
  await findAndSendMany(res, "observations", req, {
    _id: {
      $regex: `^${req.params.chain}/${req.params.emitter}/${req.params.sequence}/.*`,
    },
  });
});

app.get(
  "/api/observations/:chain/:emitter/:sequence/:signer/:hash",
  async (req, res) => {
    const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}/${req.params.signer}/${req.params.hash}`;
    await findAndSendOne(res, "observations", { _id: id });
  }
);

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`);
});
