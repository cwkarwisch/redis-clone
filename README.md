This is a a Go implementation of a Redis clone following the stages laid out in the
["Build Your Own Redis" Challenge](https://codecrafters.io/challenges/redis) at codecrafters.

The Redis server accepts connections on port 6379. The server can handle concurrent connections and supports commands like `PING`, `SET`, `GET`, `ECHO` and `CONFIG`. The client can optionally specify an expiration in milliseconds when calling `SET` if the `PX` flag is included, which the server will respect on subsequent `GET` requests. The server expects requests to be sent in accordance with the [Redis serialization protocol specification](https://redis.io/docs/latest/develop/reference/protocol-spec/#resp-protocol-description) and responses will also be sent in accordance with the specification.

If an rdb file is present with the binary representation of the in-memory store, the server is capable of parsing the rdb file and responding to the `KEYS` command based on the contents of the persisted state.
