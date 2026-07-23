# Redis Commands — To Be Implemented

Reference: [Redis official command docs](https://redis.io/docs/latest/commands/)

Deprecated commands (GETSET, SETEX, SETNX, PSETEX, HMSET, BRPOPLPUSH, RPOPLPUSH, ZREVRANGE, ZRANGEBYSCORE, etc.) are
excluded.

---

## String

| Command       | Description                                               | Done |
|---------------|-----------------------------------------------------------|------|
| `APPEND`      | Appends a string to the value of a key                    | [X]  |
| `STRLEN`      | Returns the length of a string value                      | [X]  |
| `GETRANGE`    | Returns a substring of the string stored at a key         | [X]  |
| `SETRANGE`    | Overwrites part of a string value at a given offset       | [ ]  |
| `INCRBY`      | Increments the integer value of a key by a number         | [ ]  |
| `INCRBYFLOAT` | Increments the floating-point value of a key by a number  | [ ]  |
| `MGET`        | Returns the string values of multiple keys                | [ ]  |
| `MSET`        | Sets the string values of multiple keys                   | [ ]  |
| `MSETNX`      | Sets multiple keys only when none of them exist           | [ ]  |
| `SETBIT`      | Sets or clears the bit at an offset in a string value     | [ ]  |
| `GETBIT`      | Returns the bit value at an offset in a string value      | [ ]  |
| `BITCOUNT`    | Counts the number of set bits in a string                 | [ ]  |
| `BITPOS`      | Finds the first set or clear bit in a string              | [ ]  |
| `BITOP`       | Performs bitwise AND, OR, XOR, NOT on multiple strings    | [ ]  |
| `BITFIELD`    | Performs arbitrary bitfield integer operations on strings | [ ]  |
| `BITFIELD_RO` | Read-only variant of BITFIELD                             | [ ]  |

---

## Generic / Key

| Command            | Description                                                             | Done |
|--------------------|-------------------------------------------------------------------------|------|
| `EXPIRE`           | Sets the expiration time of a key in seconds                            | [ ]  |
| `EXPIREAT`         | Sets the expiration time of a key to a Unix timestamp (seconds)         | [ ]  |
| `EXPIRETIME`       | Returns the expiration time of a key as a Unix timestamp (seconds)      | [ ]  |
| `PEXPIRE`          | Sets the expiration time of a key in milliseconds                       | [ ]  |
| `PEXPIREAT`        | Sets the expiration time of a key to a Unix timestamp (milliseconds)    | [ ]  |
| `PEXPIRETIME`      | Returns the expiration time of a key as a Unix timestamp (milliseconds) | [ ]  |
| `PERSIST`          | Removes the expiration time of a key                                    | [ ]  |
| `TYPE`             | Returns the data type of a key's value                                  | [ ]  |
| `KEYS`             | Returns all key names matching a pattern                                | [ ]  |
| `SCAN`             | Iterates over the key names in the database                             | [ ]  |
| `RENAME`           | Renames a key, overwriting the destination if it exists                 | [ ]  |
| `RENAMENX`         | Renames a key only when the destination key doesn't exist               | [ ]  |
| `COPY`             | Copies the value of a key to a new key                                  | [ ]  |
| `UNLINK`           | Asynchronously deletes one or more keys                                 | [ ]  |
| `RANDOMKEY`        | Returns a random key from the database                                  | [ ]  |
| `TOUCH`            | Returns the number of existing keys and updates their last access time  | [ ]  |
| `SORT` / `SORT_RO` | Sorts the elements in a list, set, or sorted set                        | [ ]  |

---

## List

| Command   | Description                                                                                             | Done |
|-----------|---------------------------------------------------------------------------------------------------------|------|
| `BLMOVE`  | Blocking version of LMOVE — pops from one list, pushes to another, blocks until an element is available | [ ]  |
| `BLMPOP`  | Blocking pop from the first non-empty list among multiple keys; blocks until an element is available    | [ ]  |
| `LINSERT` | Inserts an element before or after a pivot element in a list                                            | [ ]  |
| `LMPOP`   | Pops multiple elements from the first non-empty list among multiple keys                                | [ ]  |
| `LPOS`    | Returns the index (or indices) of elements matching a value in a list                                   | [ ]  |
| `LPUSHX`  | Prepends one or more elements to a list only if the key already exists                                  | [ ]  |
| `RPUSHX`  | Appends one or more elements to a list only if the key already exists                                   | [ ]  |
| `LREM`    | Removes N occurrences of a value from a list                                                            | [ ]  |
| `LSET`    | Sets the value of an element at a given index                                                           | [ ]  |
| `LTRIM`   | Trims a list to a specified index range, removing all elements outside it                               | [ ]  |

---

## Hash

| Command        | Description                                                                | Done |
|----------------|----------------------------------------------------------------------------|------|
| `HKEYS`        | Returns all field names in a hash                                          | [ ]  |
| `HVALS`        | Returns all values in a hash                                               | [ ]  |
| `HLEN`         | Returns the number of fields in a hash                                     | [ ]  |
| `HSETNX`       | Sets the value of a field only when the field doesn't exist                | [ ]  |
| `HINCRBYFLOAT` | Increments the floating-point value of a hash field                        | [ ]  |
| `HSTRLEN`      | Returns the length of the string value of a hash field                     | [ ]  |
| `HRANDFIELD`   | Returns one or more random fields from a hash                              | [ ]  |
| `HSCAN`        | Iterates over fields and values of a hash                                  | [ ]  |
| `HGETDEL`      | Returns the value of a field and deletes it                                | [ ]  |
| `HGETEX`       | Gets a field value and optionally sets its expiration                      | [ ]  |
| `HSETEX`       | Sets a field value and optionally sets its expiration                      | [ ]  |
| `HTTL`         | Returns the TTL in seconds of a hash field                                 | [ ]  |
| `HPTTL`        | Returns the TTL in milliseconds of a hash field                            | [ ]  |
| `HEXPIRE`      | Sets the expiration of hash fields (relative, seconds)                     | [ ]  |
| `HEXPIREAT`    | Sets the expiration of hash fields (absolute Unix timestamp, seconds)      | [ ]  |
| `HEXPIRETIME`  | Returns the expiration of hash fields as a Unix timestamp (seconds)        | [ ]  |
| `HPEXPIRE`     | Sets the expiration of hash fields (relative, milliseconds)                | [ ]  |
| `HPEXPIREAT`   | Sets the expiration of hash fields (absolute Unix timestamp, milliseconds) | [ ]  |
| `HPEXPIRETIME` | Returns the expiration of hash fields as a Unix timestamp (milliseconds)   | [ ]  |
| `HPERSIST`     | Removes the expiration from hash fields                                    | [ ]  |

---

## Set

| Command       | Description                                                         | Done |
|---------------|---------------------------------------------------------------------|------|
| `SADD`        | Adds one or more members to a set                                   | [ ]  |
| `SREM`        | Removes one or more members from a set                              | [ ]  |
| `SMEMBERS`    | Returns all members of a set                                        | [ ]  |
| `SISMEMBER`   | Determines whether a member belongs to a set                        | [ ]  |
| `SMISMEMBER`  | Determines whether multiple members belong to a set                 | [ ]  |
| `SCARD`       | Returns the number of members in a set                              | [ ]  |
| `SPOP`        | Removes and returns one or more random members from a set           | [ ]  |
| `SRANDMEMBER` | Returns one or more random members from a set without removing them | [ ]  |
| `SMOVE`       | Moves a member from one set to another                              | [ ]  |
| `SINTER`      | Returns the intersection of multiple sets                           | [ ]  |
| `SINTERSTORE` | Stores the intersection of multiple sets in a key                   | [ ]  |
| `SINTERCARD`  | Returns the cardinality of the intersection of multiple sets        | [ ]  |
| `SUNION`      | Returns the union of multiple sets                                  | [ ]  |
| `SUNIONSTORE` | Stores the union of multiple sets in a key                          | [ ]  |
| `SDIFF`       | Returns the difference between multiple sets                        | [ ]  |
| `SDIFFSTORE`  | Stores the difference of multiple sets in a key                     | [ ]  |
| `SSCAN`       | Iterates over members of a set                                      | [ ]  |

---

## Sorted Set

| Command            | Description                                                              | Done |
|--------------------|--------------------------------------------------------------------------|------|
| `ZADD`             | Adds one or more members with scores to a sorted set                     | [ ]  |
| `ZCARD`            | Returns the number of members in a sorted set                            | [ ]  |
| `ZCOUNT`           | Returns the count of members with scores within a range                  | [ ]  |
| `ZINCRBY`          | Increments the score of a member in a sorted set                         | [ ]  |
| `ZSCORE`           | Returns the score of a member in a sorted set                            | [ ]  |
| `ZMSCORE`          | Returns the scores of multiple members in a sorted set                   | [ ]  |
| `ZRANK`            | Returns the rank of a member ordered by ascending score                  | [ ]  |
| `ZREVRANK`         | Returns the rank of a member ordered by descending score                 | [ ]  |
| `ZRANGE`           | Returns members within a range of indexes                                | [ ]  |
| `ZRANGESTORE`      | Stores a range of members into a key                                     | [ ]  |
| `ZREM`             | Removes one or more members from a sorted set                            | [ ]  |
| `ZREMRANGEBYRANK`  | Removes members within a range of indexes                                | [ ]  |
| `ZREMRANGEBYSCORE` | Removes members within a range of scores                                 | [ ]  |
| `ZREMRANGEBYLEX`   | Removes members within a lexicographical range                           | [ ]  |
| `ZPOPMIN`          | Removes and returns the lowest-scoring members                           | [ ]  |
| `ZPOPMAX`          | Removes and returns the highest-scoring members                          | [ ]  |
| `BZPOPMIN`         | Blocking version of ZPOPMIN                                              | [ ]  |
| `BZPOPMAX`         | Blocking version of ZPOPMAX                                              | [ ]  |
| `ZMPOP`            | Pops the highest- or lowest-scoring members from one or more sorted sets | [ ]  |
| `BZMPOP`           | Blocking version of ZMPOP                                                | [ ]  |
| `ZLEXCOUNT`        | Returns the count of members within a lexicographical range              | [ ]  |
| `ZSCAN`            | Iterates over members and scores of a sorted set                         | [ ]  |
| `ZRANDMEMBER`      | Returns one or more random members from a sorted set                     | [ ]  |
| `ZINTER`           | Returns the intersection of multiple sorted sets                         | [ ]  |
| `ZINTERSTORE`      | Stores the intersection of multiple sorted sets in a key                 | [ ]  |
| `ZINTERCARD`       | Returns the cardinality of the intersection of multiple sorted sets      | [ ]  |
| `ZUNION`           | Returns the union of multiple sorted sets                                | [ ]  |
| `ZUNIONSTORE`      | Stores the union of multiple sorted sets in a key                        | [ ]  |
| `ZDIFF`            | Returns the difference between multiple sorted sets                      | [ ]  |
| `ZDIFFSTORE`       | Stores the difference of multiple sorted sets in a key                   | [ ]  |
