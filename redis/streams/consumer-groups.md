## Consumer Groups

### The problem with slow consumers
The stream data structure allows order, storage, and processing of messages.
Messages are created and stored in it by producers, and consumers can read the messages
by iterating ranges, or one by one as they become available. With streams, producers write to the front of data structures,
and consumers read messages from the oldest to the newest. Multiple consumers can consume the same stream
in order to parallelize message processing. In this video, we'll review the cases in which parallel message
processing does not provide a good solution for scaling. Although the stream decouples the operation of consumers
and producers, the former depends on the latter, but not visa versa. Once producers start growing the stream,
it is possible that consumers will begin lagging behind. This is what is called the slow consumer.
But for all intents and purposes, it can be thought also as the fast producer.
It's all relative. Anyway, the reason for the lag is that the rate at which messages arrive
may be greater than the rate at which they can be processed. Either there are too many messages
coming in, or that the time that it takes to process a message is too long, or both. One way to deal with this as 
we've seen in the divides by two and three example, is to employ multiple concurrent consumers that execute sub-tasks to complete a greater shared task.
However, not all tasks can be easily refactored to concurrent independent sub-tasks. And even when it is possible, 
some of the sub-tasks may be harder to perform, processing wise, than others. Let's assume for a second that I possess an extremely fast natural numbers
producer, one that can produce 100,000 numbers every second. Let's also assume that an instance of the running sum consumer is only capable of 
ingesting half that number. Adding another consumer to assist with the running sum will not work, because both consumer instances will receive the same messages. 
And by processing the same numbers twice, they'll effectively double the running sum.

One way to work around this is to partition the stream. Partitioning is basically the use of multiple keys, each storing a subset of a logical stream. 
In our example, we could have one partition for even numbers and another for the odd numbers. We then assign one consumer per partition.
The main problem with this approach, however, is that we'd be creating a tight dependency between the number of stream partitions and the number of consumers. 
If, for any reason, we require more consumers to do our summing, or perhaps less of them, we'd need to repartition the stream accordingly.
We'd like to avoid that, if possible. Partitioning is a powerful pattern for scaling the throughput of the stream's base setup,
and we'll discuss it later in more depth. But in this case, it presents a less than desirable solution, as it isn't dynamic enough. 
The need for dynamically dividing messages between consumers becomes even more pronounced in use cases, where the consumers performance is variable and non-deterministic. 
Performance, which translates to execution time, may be affected by any number of factors. A slow network would increase the latency between the system's components.
Or an I/O congestion would slow down the consumer, or perhaps a busy processing core, or the availability
of some remote service. And sometimes it is the data itself that affects the execution time. The fast producer, slow consumer problem is indeed challenging. 
The processing of tasks that cannot be broken down and parallellized, and ones that execute in non-deterministic time, requires a pool of consumers,
each ingesting a subset of the stream at its own pace. Such challenges are often managed with a queue but Redis's streams provide built in tooling just for that.

### Creating Consumer Groups
A stream's consumer may lag when the rate of messages exceeds its capacity to process them. And some processing needs are either
hard or impossible to parallelize. A logical consumer, that is a consumer that is made up of a group of instances where each processes
a different message from the stream, can be employed to distribute the workload according to the availability of its instances.
Such a consumer is logical in the sense that it does not really exist as an independent entity,
but is rather made up of the operation of its constituent instances. The logical consumer's input is the stream itself. Each message in the stream can be
thought of as being added to a queue dedicated to that logical consumer's instances. Instances of the logical consumer can then dequeue messages for processing.
Each consumer instance is therefore served with an exclusive subset of the stream's contents, allowing for a distribution of the workload among instances.
Redis Streams supports this pattern natively with a feature called consumer groups. A consumer group is an implementation

of the logical consumer for a single stream. It can be made up of any number of physical consumers.
Consumers in a group may join or leave as needed, which is a good fit for the needs of the dynamic setups we've mentioned. The Redis server keeps track of which
consumers belong to which group and the messages each has processed. We'll begin our group orientation by looking at one of Redis `XGROUP` command
forms of the invocation. To create a consumer group on an existing stream, one only needs to invoke the `XGROUP` command with the `CREATE` sub-command, 
the name of the stream's key, and the initial message ID that the group's members should start processing from. In this example, the first attempt to create "group0"
fails because the group's stream does not exist yet.

```
XGROUP CREATE numbers 0 0
```

By adding the optional MKSTREAM sub-command let Redis know that it is OK to create the stream even though it has no messages in it yet.
In case, the stream already exists, Redis just ignores that optional directive.

```
XGROUP CREATE numbers group0 0 MKSTREAM
XEXISTS
XLEN numbers
```

It is worth noting that consumer groups, unlike perhaps any other Redis construct, must be explicitly created with the `XGROUP CREATE` invocation.
`group0` was created using the partial message ID 0, and just like with `XREAD` that ID is interpreted as 0-0.
This means that the group is poised at the stream's very beginning before any messages in it. So the consumers belonging to it will start with the first message.
Let's add that first message to our natural numbers stream.

```
XADD numbers * n 0

```
`XGROUP CREATE` can also be called with the special `$` message ID, signifying that the group should begin consumption at the 
next new message in the stream exactly like how `XREAD` operates.
```
XGROUP CREATE numbers group1 $
XADD numbers * n 1
XADD numbers * n 2
```
As shown in the terminal, the newly created consumer group `group1` will start its processing at the stream's second message
being the natural number 1 that was added after its creation. 

We've created two consumer groups -- `group0` and `group1` -- on the numbers stream. `group0` was created before the streams even
existed and had triggered its creation with the `MKSTREAM` command. The group was set to begin at the partial message ID 0 or 
the stream's very first message. After adding that first message to the stream, we created `group1` with the special dollar message ID.
This had positioned that group at the tip of the stream after any messages already in it and just before the next message.

There is another useful Redis command called `XINFO` that, when used with the group's sub-command, provides exactly that information and then some.
In the example's output, you can see both groups 0 and 1 as well as their last-delivered message IDs. The `last-delivered-id` of the group
is an implementation of the same approach that we used by storing the ID in a Redis hash with our single running sum consumer example.

```
XINFO GROUPS numbers
```

A Redis stream consumer group is explicitly created and initialized with an initial message ID. Each group consumes the stream independently from the other
with the pace of consumption being dictated by the group's members. In the next chapter, we'll see how consumers join the groups for performing the 
equivalent of an `XREAD` operation.

### Adding Consumers to a Group
A consumer group is made up of multiple consumers, each of which processes a portion of the stream. Each group is identified by name and is associated
with a single stream and the last message ID that was delivered to the group's members. After a consumer group is created, members can join the group to 
access the stream's contents. Joining a group is a decision made by the consumer instance. To join and start processing messages, all that the consumer needs to know
are the names of the stream and the group that it wants to belong to. This means there is no need to tell Redis beforehand about the comings and goings of consumers.
Groups are open to any consumer that wishes to join them. Besides the knowledge of the key and the groups' names, in order to become a member, the instance
must provide its own name. A consumer's name uniquely identifies that instance among all other consumer instances in the group. Redis will use that name to track
the consumer's progress in the group, and we will discuss that shortly. The consumer's name is only meaningful in the context of a given stream and a consumer group.

As a side note refresher, Redis' `CLIENT SETNAME` can be used to assign a name for the connection. This can be quite useful when you're trying to track
and debug complex behaviors in general, so it is considered good practice to use it. Because consumer group members require a name, you should consider giving them the 
same name as the client. At a later time, if and when needed, the output from `CLIENT LIST` will be infinitely more helpful.

```
CLIENT SETNAME redis-c01
CLIENT LIST
```

That's almost all there is to it. We've seen how a single instance consumer can use `XREAD` to process a stream. And there's a similar command called
`XREADGROUP` to do the same, only as part of a group. There's also another special message ID that you need to know about, but this is the last one.

```
XREADGROUP GROUP group0 redis-c01 COUNT 1 BLOCK 1000 STREAMS numbers >
```

Let's turn to the CLI to check it out. That's somewhat of a long command, but most of its pieces we are already familiar with. `XREADGROUP` shares with 
`XREAD` the same `COUNT`, `BLOCK`, and `STREAMS` sub-commands and argument that these expect.

The first difference between the two commands, beside their names, however, is that `XREADGROUP` requires the `GROUP` sub-command followed by the consumer group's
name, `group0` in the example, and the consumer's name, `consumerA`. By issuing this `XREADGROUP` request, `consumerA` has officially joined the group.

The second difference to notice is, as promised, the use of a special message ID -- the greater-than sign (">"). The ">" is used exclusively by `XREADGROUP`.
This means the message ID that is greater than any of the previously delivered IDs to the group's member. Put differently, the special ">" means,
"Give me the next undelivered message", where "me" is a named consumer within the group. As expected, the command returns immediately with a reply that consists of the stream's name
and the first message in it. Calling `XREADGROUP` with the `>` special message ID is what we'd usually want to do. 

After all, a single consumer's purpose is usually to process new, yet to be delivered messages from the stream. And that's exactly what the command does.
But messages also need to be acknowledged by their consumers. A message that has been delivered to one of the group's members is said to be "pending" in the sense that it
was delivered, but its processing is yet to be acknowledged by the consumer. Redis keeps track of which messages have been delivered to which consumer group members.
For that purpose, it uses an internal data structure called the `Pending Entries List`, or `PEL` for short. We'll discuss later how a message's state can be changed from "pending" to something else,
or, put differently, how it can be removed from the `PEL`. Under normal conditions, the removal from the `PEL` is done once the consumer acknowledges
successfully processing it. The Pending Entries List is what makes a consumer group tick. In it, as its name suggests, are the IDs of messages in the stream that are
pending for each consumer, along with additional metadata.

This core structure is what gives the group its power, allowing Redis to keep tabs on the current state. `XREADGROUP` can also handle partial or full message IDs.
And just like with `XREAD`, these are interpretedas the non-exclusive lower bound, or the message preceding the next one.Unlike `XREAD` however, when used with the message ID,
`XREADGROUP` will only return messages that were already delivered to the named consumer, effectively masking the activity of other consumers from itand providing it with its 
own unique point of view. This masking, of course, is possible because of the `PEL`.
Here's how that looks like in practice. When called by `consumerA` from `group0` again, this time with the partial message ID 0, we get the same first message delivered again.
The consumer in a group can only view that part of the stream that was already delivered to it and is yet to be processed.

```
XREADGROUP GROUP group0 consumerA COUNT 1 BLOCK 1000 STREAMS numbers 0
```

The reply is effectively the entire `PEL` for "consumerA". From the point of view of another consumer group member,`consumerB`, the stream currently appears to be empty.
That's the expected behavior. "consumerB" had yet to ask for any new messages to be delivered to it, so its `PEL` is empty. 

```
XREADGROUP GROUP group0 consumerB STREAMS numbers 0
```

Of course, "consumerB" can `XREADGROUP` the next message with the ">" special message ID and receive the stream's second message, which contains the number 1. 

```
XREADGROUP GROUP group0 consumerB COUNT 1 BLOCK 1000 STREAMS numbers >
```

It is now possible to re-run `XREADGROUP` for both consumers to obtain each one's unique perspective. We saw how `XINFO GROUPS` provides an overview of the stream's groups.

```
XINFO GROUPS numbers
```

Let's call it again. This time, "group0"'s consumers value shows 2. That value is the count of consumers that belong to the group and accounts for "consumerA" and "consumerB".
The group's pending value is 3. That value is the total count of pending messages for all consumers, or the sum of the length of their PELs.
Also note how the last delivered ID is now set to that of "consumerB"'s message, allowing Redis to serve the next group request from there.

```
XINFO GROUPS numbers
```

Another way to inspect the state of the group's consumers is to call `XINFO` with the `CONSUMERS` sub-command. This form of XINFO accepts the stream's key and group names
as input, and in return, provides a paired member breakdown. 

```
XINFO GROUPS numbers group0
```

For each consumer in the group, it returns the PEL's length and idle time, which is the count in milliseconds
since that consumer had last read a new message. Members of groups have unique names that identify them. They read undelivered messages from the stream
with the `XREADGROUP` command and can only access messages that were delivered to them and not yet processed.

Redis maintains an internal list of pending entries, acronymed as PEL, for each consumer in the group to track its progress. The `XINFO` command allows us to inspect the stream's group
and consumers with the `GROUPS` and `CONSUMERS` sub-commands, respectively.

### Processing Messages
The consumer group is a Redis-operated mechanism for distributing messages between members. Consumers can join a group simply
by asking for new messages and then providing their name. Once a consumer in a group reads messages, these are added to its Pending Entries List.
It is up to the consumer now to actually process them. Every consumer is different. But ultimately, the processing of the given message can end in either success or failure.
Like in real life, dealing with success is much easier than coping with failure. So let's begin with that.

To deal with success, we first need to define it. Here's the thing. For a consumer instance to succeed, it needs to read the message, perhaps do some meaningful work
with the data in it, and then get to the point where it is ready to process the next message. Put slightly differently, in order to succeed,
the consumer needs only not to fail. The actual work performed by the consumer and any results obtained from it, while important in the grander scheme
of things, are outside of the scope. Getting the message and surviving the processing is the only measure of success in the context of stream consumer groups.

Consumer groups provide at-least-once message delivery semantics by default. We've seen how the consumer instance can read its PEL, and that ability is what allows it to recover.
But when the consumer does not fail, that is, when it succeeds processing a message, it needs to let the server know about that. This act, which is called acknowledgment,
is just another task among other housekeeping tasks that the recipient does when it finishes its work. Recall that consumerA from "group0" requested a single message, the first message
in the natural numbers stream.

```
XREADGROUP GROUP group0 consumerA STREAMS numbers 0
```

From the group's perspective, this message is currently pending processing. To signal that processing was completed successfullyfor that message, we acknowledge it with the `XACK` command.
`XACK` is a straightforward command, that has only one form of invocation. It accepts the name of the stream's key, the name of the consumer group, and one or more messages IDs that are acknowledged.
The reply returned by `XACK` is the number of messages that were actually acknowledged that is removed from the group's PEL. The first message ID was in "consumerA"'s `PEL`.

```
XACK numbers group0 <message-id>
```
It is interesting to note that `XACK` does not require the consumers names, only that of the group.
While in most cases, it will be the consumer itself doing the `XACK'ing` as part of its successful housekeeping, that is not mandatory. This liberty makes it possible for other processes
to acknowledge messages, which is definitely useful even if only for administrative purposes. Once we've acknowledged the message, we can verify that "consumerA"'s `PEL` is now empty.

```
XREADGROUP GROUP group0 consumerA STREAMS numbers 0
```

At this point, since it has no pending messages to process, "consumerA" can issue a call to `XREADGROUP` and use the greater than sign to get new messages.

```
XREADGROUP GROUP group0 consumerA COUNT 1 block 1000 STREAMS numbers
```

To recap this pattern, a consumer group member requests new messages but can fail at any time before acknowledging them.
To avoid the loss of pending messages, the consumer first reads its history of pending messages before requesting new ones.
This provides us with at-least-once message delivery semantics.


## Basic Consumer Group Managerment
But how do you manage consumer groups in the longer term? Let's start to answer that question by looking at some important consumer group administration
tasks.

* First, we'll learn how to change a consumer group's position in the stream.
* Second, we'll learn how to delete a consumer group. 
* Third, we'll learn how to remove an individual consumer from a consumer group, and some of the important considerationssurrounding this.

To illustrate these operations, let's go back to our natural numbers stream and create a few groups. First, we'll create a group called
"primes" that starts at the end of the stream.

```
XGROUP CREATE numbers primes $
```

Next, we'll create a group called "sums" that starts at ID 0, or the beginning of the stream.

```
XGROUP CREATE numbers sums 0
```

Finally, let's create another group called "averages" that also starts at the beginning of the stream.

```
XGROUP CREATE numbers sums 0
```

We'll also create a few consumers for the averages group. 

```
XREADGROUP GROUP averages A COUNT 2 STREAMS numbers >
XREADGROUP GROUP averages B COUNT 1 STREAMS numbers >
XACK numbers averages <message-id>
XREADGROUP GROUP averages C COUNT 1 STREAMS numbers >
XACK numbers averages <message-id>
```

Consumer "A" has consumed two messages without acknowledging them. Consumers "B" and "C" have both consumed a single message,
and have acknowledged that message. Now, we need to talk about changing a consumer group's position.

The primes consumer group is currently at the end of the stream, while the sums and averages consumer
groups are at the beginning. We can confirm this by running a couple of `XINFO` commands.

```
XINFO GROUPS numbers
```

When we run XINFO GROUPS, we can see the last delivered ID for each consumer group. For primes, it's a very specific ID.
In fact, when we run `XINFO STREAM numbers`, we see that the last entry ID is the same as the primes consumer group's last delivered ID.

```
XINFO STREAM numbers
```

So the upshot is that consumer groups always have a specific position in the stream that's dictated by their last entry ID.
When consuming a stream, the consumers in a group will receive messages whose IDs start after the last entry ID.
So if we want to change a consumer's position in the stream, we need to alter this last entry ID. For that, we can use `XGROUP SETID` sub-command.

For example, suppose we eventually decide that we want the primes consumer group to start reading from the beginning of the stream
instead of from the end. We simply run this command. 

```
XGROUP SETID numbers primes 0
```

If we want to position the group at an arbitrary ID, we can do that too.

```
XGROUP SETID numbers primes <message-id>
```

Or we can place the consumption back at the end of the stream using the special dollar sign ID.

```
XGROUP SETID numbers primes $
```

The use cases here should be obvious. Either we want to replay a stream from 
* the beginning,
* start consumption from somewhere in the middle -- perhaps at a specific timestamp.
* start consuming the stream from this moment onward, only processing new messages from here on out.

Now, let's learn how to delete a consumer group. It's first important to remember that Redis does not clean up
unused consumer groups, so we need a way to clean them up on our own.

Maybe we're done with a particular consumer group, or perhaps we accidentally created it.
Either way, the `XGROUP DESTROY` command will permanently remove the specified consumer group and any associated consumers, so you
need to use this command with some care.

Here's how to use it. Suppose we decide we're done with the sums consumer group. We simply run `XGROUP DESTROY`, 
specifying the stream name and the group name.

```
XGROUP DESTORY [stream-name] [group-name]
```
```
XGROUP DESTROY numbers sums
```

Now, in addition to removing entire consumer groups, we can also remove individual consumers from a group.
There are a number of reasons why we might want to do this. One is that the system running the consumer no longer exists.
Whatever the reason, consumers themselves are easy to delete. We simply use the `XGROUP DELCONSUMER` command,
passing in the name of the stream, the name of the group, and the name of the consumer.
For example, here's how to delete consumer "C" from the averages consumer group.

```
XGROUP DELCONSUMER [stream-name] [group-name] [consumer-name]
```

```
XGROUP DELCONSUMER numbers averages C
```

This command returns the number of pending entries owned by the deleted consumer.Consumer "C" had zero pending entries.
That's straightforward, and it's perfectly safe to do when the consumer's pending entries list is empty,
as in this case. But to delve a bit deeper here, let's take a look at the remaining consumers for the averages
group. 

```
XINFO CONSUMERS numbers averages
```

You'll notice that consumer A has two pending entries. As pending entries, these entries have not been acknowledged by the consumer.

If we delete this consumer, we won't know if they've been processed correctly. This is a potentially common scenario if processes 
have not been processed correctly then you'll need to assign them elsewhere before deleting the consumer.
How do you do that? The short answer is that you use a combination of the `XCLAIM`and `XPENDING` commands, 


