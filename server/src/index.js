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

async function paginatedFind(collection, req) {
  const limit =
    req.query?.limit && req.query.limit <= 100 ? req.query.limit : 20;
  const skip = req.query?.page ? req.query?.page * limit : undefined;
  const query = req.query?.before
    ? { createdAt: { $lt: new Date(req.query.before) } }
    : {};
  const cursor = await collection.find(query, {
    sort: { createdAt: -1 },
    skip,
    limit,
  });
  return cursor;
}

app.get("/api/heartbeats", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const heartbeats = database.collection("heartbeats");
  const cursor = heartbeats.find();
  const result = await cursor.toArray();
  res.send(result);
});

app.get("/api/vaas/:chain/:emitter/:sequence", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}`;
  const database = mongoClient.db("wormhole");
  const vaas = database.collection("vaas");
  const result = await vaas.findOne({ _id: id });
  if (!result) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/vaas", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const vaas = database.collection("vaas");
  const cursor = await paginatedFind(vaas, req);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/observations/:chain/:emitter/:sequence", async (req, res) => {
  const id = `${req.params.chain}/${req.params.emitter}/${req.params.sequence}`;
  const database = mongoClient.db("wormhole");
  const observations = database.collection("observations");
  const cursor = observations.find({ _id: { $regex: `^${id}/*` } });
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.get("/api/observations", async (req, res) => {
  const database = mongoClient.db("wormhole");
  const observations = database.collection("observations");
  const cursor = await paginatedFind(observations, req);
  const result = await cursor.toArray();
  if (result.length === 0) {
    res.sendStatus(404);
    return;
  }
  res.send(result);
});

app.listen(port, () => {
  console.log(`Example app listening on port ${port}`);
});
