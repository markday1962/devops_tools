## Performance Considerations

The only way to add messages to a Redis stream is by using the XADD command.
As we have seen XADD places new messages into the stream by appending them.
No other insertion strategies are available. So we can think of a stream as an append only log.
XADD has a time complexity of O of one, or constant time. This means that we can expect XADD's performance
to remain the same regardless of whether we are adding the first, hundredth, or 10 millionth message to a stream.
```
XADD mystream * hello world
XADD mystream * hello there
XADD mystream * hello everybody
```
Producers need to be able to add new messages into the stream at rates not affected by the overall length of the stream.
When reading data from the stream it's also desirable to be able to fetch one or more messages from it with predictable performance,
ensuring consistent consumer experience regardless of the streams length or the position of the messages in it.
When consuming from the end of the stream using XREAD or XREADGROUP it should come as no surprise that their performance is constant.
However, the fact that their performance remains constant when reading from arbitrary positions in the middle of the stream, is perhaps more noteworthy.

This holds not only for XREAD and XREADGROUP, but also for the XRANGE and XDEL commands. Each of these commands is able to access
any message in a stream with constant time performance because of the way that Redis streams are implemented.
```
XREAD COUNT 3 STREAMS mystream 0
XRANGE mystream <message-id> <message-id>
XDEL mystream <message-id>
```
One thing to keep in mind however, even though locating a key is an O of 1 operation,
scanning from that key onwards is still O of N on the number of messages returned.
For example, running an XRANGE command with a counter argument of a million
is still O of a million. So in general be careful with the count argument when running
XRANGE, XREAD, and XREADGROUP. It's hard to make a rule of thumb here, but if you need more than 1,000 messages returned,
it's probably best to break this into multiple calls.
```
XRANGE n - + COUNT 1000000
```
## Stream Memory Management

https://redis.io/commands/memory-usage

https://redis.io/commands/xtrim

For many use cases with large data sets, streams can require significantly less space
than other Redis data structures. As we saw in the previous module,
streams are implemented using a Radix tree data structure, an extremely efficient way
of storing key value data. The Redis stream's implementation
contains a further optimization to reduce memory usage, field name compression.
When multiple consecutive messages have the same field names and are appended to the stream,
Redis will not replicate the storage of the field names. Instead, each subsequent message having the same set
of field names is marked using a flag rather than repeating the same field names over and over.
This means that you can save memory by having the same structure for a series of messagesin a stream while retaining the flexibility to change
that structure over time.

A stream processing approach is commonly chosen when working with a data set that has the potential
to grow indefinitely. For example, a stream of temperature readings from a fleet of IoT sensors has no logical end to it.
Redis does not, of course, have the ability to retain an infinitely growing data set.
A Redis server will reject write operations when it runs out of memory. This means that it is important to manage the size of a stream
so that it doesn't grow on forever. Remember that consumers read from the stream,
but do not cause messages to be deleted from it once consumed. You need to manage the stream's length separately.
When using streams, it's important to understand how much memory you might need for your specific data set 
and the options available to you to manage that. In this unit, we saw how streams can be more memory efficient than other data structures.

```
MEMORY USAGE
```

Question: Suppose you're trying to decide whether to use a Redis stream, sorted set, or list. 
Memory efficiency is the most important criterion. How do you decide which data structure to use?

Answer: It's always best to generate a sample data set and compare the memory usage using the built-in MEMORY USAGE command.


Question: If a stream’s growth is allowed to continue unchecked, what will eventually happen once the Redis Server runs out of memory?

Answer: Redis does not automatically truncate or otherwise manage the length of streams or the resources required to maintain them. 
A stream’s length has to be managed separately from its production and consumption. In a low memory situation, the Redis Server will 
reject further write operations but will continue to serve reads.

Question: Which implementation details account for the memory efficiency of Redis streams?

Answer: The use of a Radix Tree provides a space-optimized data structure for the stream. Redis Streams keys are particularly suited to 
such a structure as they are relatively small and have repeating elements in them due to the use of millisecond timestamps in the keys.
Additionally, the storage overhead associated with message payload metadata is reduced by flagging a series of messages 
all with same field names, avoiding redundant storage of the field name strings
