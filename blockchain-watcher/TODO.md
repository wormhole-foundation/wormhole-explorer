Each blockchain has three watchers,

- A websocket watcher for low latency
- A querying watcher
- And a sequence gap watcher

These three watchers all invoke the same handler callback.

The handler callback is responsible for:

- Providing the watchers with the necessary query filter information
- parsing the event into a persistence object
- invoking the persistence manager

The persistence manager is responsible for:

- Inserting records into the database in a safe manner, which takes into account that items will be seen multiple times.
- Last write wins should be the approach taken here.
