## Redis Streams Overview
Redis Streams is a sophisticated feature. But you can make sense of it by understanding
the following main ideas--
* First, Redis Streams is essentially a new Redis data structure. 
This data structure is known as a stream.
* Second, the stream behaves like an append-only list or log.
* Third, each entry in a stream is structured as a set of key value pairs.  So you can 
think of a stream entry as something akin to a Redis hash.
* Fourth, every stream entry has a unique ID. And by default, these IDs are time-prefixed.
* Fifth, streams support ID-based range queries.
* Finally, streams can be consumed and processed by multiple distinct sets of consumers. 
These are known as consumer groups.

## The Redis Stream Producer
The producer API allows a producer to append an arbitrary message to a Redis stream.
The Redis stream producer interface is made up of a single command, XADD,
that the producer calls with the explicit name of the stream. That is, the key's name, 
and the message's payload or data.
Let's look at an example to understand how a producer operates.
To make the producer extremely simple to reason about, our stream's data will be the 
sequence of natural numbers as defined by ISO 80000-2. These are all the numbers beginning 
with 0 that correspond to the non-negative integers 0, 1, 2, and so forth. 
To start the stream, the producer will add a message with a single field. Let's call it "n", 
that contains the first numberin the sequence, that is 0.
```
XADD numbers * n 0
```
We've called XADD with several arguments,
* 1 numbers, is the name of the key at which the stream is stored. The name of the stream's key is the contract
between your stream's producers and consumers. And it is implicitly created if it doesn't exist,
like all other Redis data structures. Note that `XADD` is a single key command, so a message is always
added to a single stream.
* 2 The second argument to `XADD` is the new message's ID. In most cases, we want Redis to 
manage the generation of message IDs. In fact, one of the compelling reasons to use a stream 
is for its inherent ordering of messages, which is achieved by having Redis manage it.
By setting the ID argument to an asterisk, as in our example, we let Redis generate the ID for the new message.
* 3 The remainder of the arguments are the actual data. They are provided as pairs of field names 
and their values, exactly like how `HSET` is used with Redis' hashes. 
And just like Redis' hashes, every value and field name in the message is a Redis string.
A Redis string can store up to half a gigabyte of binary-safe data if needed, although messages usually
tend to be much shorter.

## Advanced Consumer Management
It's always possible for a consumer to receive a messageand then go offline before the 
message can be acknowledged.
When that happens, there will be at least one message in the consumer's Pending Entries List.
If that's the case, we need a way of getting that stranded message to a consumer that can 
effectively process and acknowledge it.

### Manage Pending Messages
XINFO and XPENDING allows for the examination of pending messages on a stream.
XCLAIM allows for the reassignment of messages

https://redis.io/commands/xinfo
https://redis.io/commands/xpending
https://redis.io/commands/xclaim

Creating a group called evens that starts at the beginning of the stream.
```
XGROUP CREATE numbers evens 0
```
Let's next create two consumers, A and B, each consuming a single message.
```
XREADGROUP evens A COUNT 1 STREAMS numbers
XREADGROUP evens A COUNT 1 STREAMS numbers
```
For consumer B, we'll acknowledge the message that it just consumed. So now what's the state 
of the consumer group?
```
XACK numbers evens 1556645412546-0
```
So now what's the state of the consumer group? Running `XINFO GROUPS`, you can see here
that we have two consumers. We also have one pending message.That should be no surprise. 
We can run `XINFO CONSUMERS` to find out where that pending message lives.
```
XINFO GROUPS
XINFO CONSUMERS
```
### XPENDING
And you can see here that it must belong to consumer A, as consumer B has no pending messages.
Now, suppose that consumer A has gone offline. In this case, we need a way to reassign 
its pending messages. There are two commands that can help with this, `XPENDING` and `XCLAIM`.
Let's start with `XPENDING`. In its most basic form, `XPENDING` takes two arguments,
the name of the stream and the name of a consumer group.
```
XPENDING [stream-name] [group-name]
```
Let's run XPENDING on our numbers stream and evens group.
```
XPENDING numbers evens
```
First, we get the total number of pending messages. Next, we can see the range of message IDs
in the Pending Entries List. We also get to see how many pending entries per consumer exist.

As we saw earlier, consumer A has one pending entry. But what does that entry look like?
To get that information, we need to run `XPENDING` with a few more arguments. The next variant 
of the `XPENDING` command takes a range of message IDs and a count. As with `XRANGE`, 
here we can use the special minus and plus operators to iterate the entire range of pending
IDs. When we do that, we can learn a bit more about the pending message ID currently 
assigned to consumer A. 
```
XPENDING [stream-name] [group-name] [start] [end] [count]
```
```
XPENDING numbers evens - + 1
```
Here we see the message ID, the name of the consumer it's assigned to, 
the number of milliseconds elapsed since it was delivered to the consumer, and a delivery 
counter, which, as you might expect, records the number of times the message has been delivered.
So what does XPENDING do for us? Effectively, it gives us more information about the currently 
pending entries for our consumer group. Specifically, it gives us the IDs of the pending messages,
the number of times each message has been read by the consumer,and the elapsed time since 
the message was last delivered.
### XCLAIM command
So suppose we've decided to reassign the pending message from consumer A to consumer B. 
We'll need to use the `XCLAIM` command. In its most basic usage, the `XCLAIM` command 
takes four arguments, the `stream name`, the `consumer group name`, the `name of the consumer`
claiming the message, and a `min idle time`. This gives you an automatic way of 
preventing `XCLAIM` from claiming a message that hasn't had a chance to be processed yet.
```
XCLAIM [stream-name] [group-name] [consumer] [min-dile-time]
```
So to claim A's pending message for B, we run the command like so. When successful, `XCLAIM` 
returns the message that was claimed.
```
XCLAIM numbers evens B 1000 [message-id]
```
if you want to reset the delivery counter, use `XCLAIM's` `RETRYCOUNT` option.
This allows you to set the counter to zero, for instance, on an `XCLAIM`.
So to summarize, the `XPENDING` and `XCLAIM` commands, combined with `XINFO`, 
giveing you the ability to detect and then reassign messages from one consumer's 
`Pending Entries List` PEL to another's.

## Consumer Recovery and Poison-Pill Messages
When exactly do you need to recover a Redis Streams consumer? And what exactly does 
recovery mean? Basically, if you want at-least-once delivery semantics, and you lose a 
consumer from a group, then you need to ensure that the messages from that consumer's 
`PEL` or pending entries list are reassigned.After all, we can't guarantee that a consumer
has processed a message until it has acknowledged the message. Redis does not automatically 
recover unacknowledged messages, that's the job of the Redis user.

### Continuously monitoring pending messages.
One This means frequent calls to the XPENDING command. Two, deciding upon a heuristic for when
a message should be assigned to a different consumer. Some combination of message delivery count,
time since delivery, and how long a consumer is idle will figure into this calculation.
Three, deciding how to reassign messages. Are they reassigned to other consumers randomly
or in a round-robin fashion or to the consumer with the lowest idle time?

### Question
Question: Using the XPENDING command, we can see the value of the time-since-delivery 
field for each message in the Pending Entries List. What happens to the time-since-delivery 
value when a consumer reads the message using XREADGROUP?
Answer: The value is set to 0

Question: Which type of process can XCLAIM a pending message and reassign it to another consumer in the group?
Answer: Any process with a connection to the Redis server correct

### Poison Pill Messages
A poison-pill is a message that's effectively unprocessable, preventing consumers
from ever acknowledging it. In the worst cases, such messages cause a consumer process to die.
But in other cases, they may trigger bugs that cause a consumer's load to spike,
thus rendering the consumer unavailable. You can imagine that if you had an automated recovery
strategy as I just described, you'd also want to think about the possibility of a poison-pill message.

After all, such a message can wreak havoc on a system as it's assigned from one consumer to the next.
So it may be important to detect the case where a message is continuously being reassigned.
In other words, you may want to keep track of how many times the same message has been `XCLAIMED`.
This may indicate the presence of the dreaded poison-pill.

### Question
Question: The criteria that can be  to determine when to claim a message for another consumer?
Answer: Redis exposes consumer idle time, time since message delivery, and the message delivery count. 
These are all reasonable metrics to use to decide when to reassign a message. 




