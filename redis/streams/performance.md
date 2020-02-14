## Performance and Memory Management

### Performance Considerations

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
### Stream Memory Management

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

### Stream Capping Strategies

https://redis.io/commands/expire

https://redis.io/commands/xlen

There are three strategies for capping a stream's length.
* 1 using a producer can manage the stream's length using `XADD`.
* 2 using `XTRIM` to cap the stream.
* 3 using a technique for capping a stream based upon its age.

### XADD
The first approach is to explicitly cap the stream's length when producing to the stream.
When adding messages to the stream with `XADD`, you can specify the `MAXLEN` subcommand.
This command tells Redis to add new messages and also remove the oldest existing messages as needed,
keeping the stream at the requested length.
An advantage of this strategy is that the stream's growth is tightly controlled.
It results in frequent checking of the stream's length against the desired length, as these
are compared on every insert into the stream. This approach can also be used without adding
additional components of the system, as the capping is performed by existing producers.
The strategy does, however, incur the performance hit associated with checking the stream's length
and potentially trimming it on every call to `XADD`. To mitigate this, you can provide the tilde option
with `MAXLEN`, which will cause the stream's length to be trimmed to approximately the number of specified
entries. When the tilde modifier is used, Redis will only trim messages when it can remove a whole node
from the underlying radix tree. This is more efficient than partial removal of the nodes entries, which is what's
required when an exact MAX length is requested. We recommend that you use the tilde approximation approach,
in general. This uses fewer server resources, freeing Redis to service other requests faster, 
while still ensuring that the length of your stream remains approximately within the bounds that you have chosen for it.
```
XLEN numbers
XADD numbers MAXLEN 1000 * n 144450
XLEN numbers
XADD numbers MAXLEN ~ 800 * n 144451
```

### XTRIM
The second stream management strategy to consider is similar to the first in that it involves periodically
trimming the stream's length. The difference is that it can be performed independently of a stream producer or a consumer,
for example, by a management component
or by an administrator. The `XTRIM` command is used to cap the stream's length and works in the same way as `ADD` and `MAXLEN` command.
Like `MAXLEN, XTRIM` can be used to cap the stream to an exact length or to an approximate length.
You should use the strategy if you wish to cap the stream periodically, rather than as a function of the operation of a producer.
This strategy can be implemented manually as an administrator or by using some other component that monitors the stream, 
capping only when its length exceeds your chosen threshold.

```
XLEN numbers
XTRIM numbers MAXLEN 800
XTRIM numbers MAXLEN ~ 500
XLEN numbers
```

### Time based
But sometimes you want to manage the stream's contents by time, rather than by overall length.
There's currently no built-in method to trim a stream by time range-- for example, to remove entries more than a week old.
If your application requires this, you should consider partitioning the stream by date/time
and expiring partitions periodically using Redis `EXPIRE` command.
Let's take a look at how that might work. One way to achieve time-based expiry is by partitioning a stream by date/time.
This way, we can either delete old partitions manually. Or we can auto expire them using the EXPIRE command.
For example, we might partition by day using a year, month, day pattern to name each partition.
The producer would then write to the current partition for the day, and consumers would read from it.
Implementing this strategy requires both producers and consumers to implement the application specific logic
for the naming of stream partitions and to know when to begin writing to and reading from them.
Some consumer read activities may need to be performed over multiple partitions
of the stream. Redis streams provides some assistance with implementing such a strategy.
The `XADD` command will create a stream if it doesn't already exist, saving
the need for an explicit creation step for each new partition. And because Redis streams are implemented as a regular data
type that exists in the main key space, the `EXPIRE` command can be used to remove stream partitions
after a given time period is passed.

The stream management strategies outlined rely on controlling the stream's growth by managing the number of messages in it.
In common with other Redis data structures, a stream's overall bite size and memory cannot be directly limited.
If you wish to size your stream such that it takes up an approximate amount of memory, then you
will need to understand the average size of your message payloads and trim by length accordingly.
If you can estimate the number of messages per day, then you can use some of the strategies described 
in the previous unit to estimate the size of your stream. When using streams, it's important to understand
how much memory you might need for your specific data set and the options available for you to manage that.

Question: Which two approaches can be used with both the XADD and XTRIM commands to control a stream’s growth?

Answer: Both `XADD` and `XTRIM` provide the ability to trim a stream’s length to either an exact or an approximate number of entries. 
Capping to an approximate number using `MAXLEN` and the tilde modifier is the most efficient and recommended way to do this.
Trimming a stream to a specific byte size or time period are not currently supported by Redis Streams.

Question: True or False? The XREAD, XRANGE and XREVRANGE commands allow consumers to trim a stream while reading from it.

Answer: XREAD, XRANGE and XREVRANGE do not provide options to trim the stream. You need to manage the stream’s length 
by using XADD with the MAXLEN sub-command or with XTRIM.

## Redis Streams Usage Patterns
Review a few important Redis Streams usage patterns. First, we'll talk about how to deal with large payloads.
Next, we'll consider the trade-offs between having one large stream and many smaller streams.
And finally, we'll talk about when to use a single consumer versus consumer groups.
There are a couple of techniques you can use to handle streams whose messages are large.
Large is, of course, relative, but keep in mind that no single stream can use more memory
than is available on any single Redis server. For example, suppose I have a stream whose messages are 50 kilobytes in size.
Once the stream is a million messages in length, that's 50 gigabytes.
And that might be larger than any of our Redis servers. So what are the solutions to this?

* First there's stream trimming, which was discussed earlier.

* Second, you can consider storing the large payload outside of the stream itself
and referencing it from the stream. If the payload needs to be hot in memory
but doesn't always have to be accessed within the stream, then store it in a secondary Redis data
structure, such as a hash. If you're running a clustered Redis deployment,
these payloads will naturally distribute across the cluster. But for especially large payloads
that don't need to be in memory, consider storing these on disk or in an external large object store.

Let's now talk about the single stream versus multiple streams question.
When using Redis streams, we often need to decide between using one large stream or multiple 
smaller streams to represent a domain. The answer to this dilemma often depends
on the stream's access patterns, or how you're going to access the stream.

Let's take the example of user notifications. Typically, a user has a mailbox of notifications,
and we need to be able to view all of the latest notifications for a single user.
For this case, one stream per user notification mailbox makes the most sense.
We can then view the latest notifications for that user with a call to `XREVRANGE`.

An example of a single global stream might be an API access log. Suppose we have tens of thousands
of users all hitting our public API. Each API access is an event in the stream. 
If our goal is to be able to analyze the overall usage patterns, then a single stream for all API access events
makes the most sense. Another point to consider here-- with many smaller streams, these can
be partitioned across nodes in a Redis cluster. With one large stream, you'll need to be more careful to cap the stream so
that it doesn't take up all the memory on a single server. This is true even if you're running a clustered Redis
deployment.

* When to use a single consumer versus consumer groups. To review, the single consumer involves
using `XREAD` to process the stream in order. While doing so, you need to keep track of your last processed
ID, possibly storing that ID in Redis. 

The consumer group pattern involves a few different commands, and we described this in detail
last week. Let's start with consumer groups. There are two scenarios where consumer groups make the most
sense.

* The first is when you need out-of-order processing.
For example, suppose each event in a stream represents a photo that requires processing.
Suppose further that we have a set of microservices that can process 10 photos at a time.
With `XREADGROUP`, we can send 10 messages at a time to the microservice, and then `XACK` each message as soon
as the photo it points to has completed processing. So we don't have to do this out-of-order accounting
on our own.

* The second case for consumer groups is when processing each message requires significant CPU.
Suppose we're doing text classification. We need to fan this out to multiple processes/threads/CPUs
to keep up with the growth of the stream. To solve this, we create a consumer group,
with one consumer per thread per process, across, perhaps, a cluster of machines.
Now, if you don't need out-of-order processing, or if your processing isn't too computationally expensive
for a single consumer, then it's best to opt for `XREAD` with a single consumer.

And here's a really important point to remember. Even if you need stream segmentation,
you can still accomplish this using a single consumer with `XREAD`.
To do this, each consumer simply needs to keep track of its own offset in the stream.

Question: What's the limit in size for a single Redis stream data structure?

Answer: A single stream can be as large as, but no larger than, the Redis instance where it resides. 
Thus, the largest Redis stream on a 50GB Redis instance will be 50GB, assuming no other data is stored on the instance.
A stream must be partitioned to work across a Redis cluster. A single stream will not automatically be partitioned by Redis across the cluster.

Question: What are the reasons for using consumer groups over a single consumer when processing a stream?

Answer: Choose consumer groups when you need to acknowledge messages out-of-order and/or when you require multiple consumers 
/ threads / CPUs in order to keep up. You can segment a stream by consumer type using both single consumers and consumer groups. 
In the single consumer case, each consumer must keep track of its own last-processed offset.Also, you can write the results of a processed 
stream to a new stream with both single consumers and consumer groups. Submit Some problems have options 
such as save, reset, hints, or show answer. These options follow the Submit button.



