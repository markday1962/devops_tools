# Final Exam
## Part 1
* 1 What benefits can Redis streams bring to a distributed application architecture?
```
A standard approach for managing communications between heterogeneous system components
The ability to provide buffers against surges in system activity
The ability to easily perform time-based range queries
```
```
Implementing streaming allows you to decouple components of your architecture, and 
provides a buffering system that can protect against surges in activity.
Using Redis Streams gives you the additional benefits of built-in time-based range queries 
for your dataset, and a standard set of APIs and proven client SDKs for many popular languages.
```

* 2 In a stream pipeline architecture, producers append new data to source streams. 
Data passes through a set of intermediate streams and consumers before arriving at one or more sinks.
```
A data warehouse
A notification mailbox
```

* 3 You have a Redis stream where new messages are added faster than a single consumer 
is able to read and process them. What could you change to speed up the overall rate of processing?
```
Introduce a consumer group containing multiple consumers correct
```
```
Consumer groups are used with Redis streams to allow one or more consumer processes to 
collaboratively process the stream. The messages in the stream are divided among each 
consumer in the group. This allows processing in parallel which can speed up the overall 
progress of the consumer group through stream's messages.
```

* 4 The default message ID format for Redis streams consists of a timestamp in 
milliseconds, a "-", and another number, for example:
```
1557357407821-0
```
In this case, the second number is 0. What purpose does this second number serve?
```
It's a sequence number that ensures that more than one message can be added to the stream 
in the same millisecond correct.
```
```
The second number in the default Redis stream message ID is a sequence number. 
This is used to allow multiple messages to be produced in the same millisecond 
while keeping each message's ID unique.
```

* 5 Which of the following commands will create the stream "numbers" if it does not already exist?
```
XADD numbers * hello world
XGROUP CREATE numbers primes $ MKSTREAM
```
```
The XADD command implicitly creates a new stream if one didn't previously exist. 
XGROUP CREATE will also create a stream if the MKSTREAM option is provided.
```

## Part 2
* 1 A stream message's payload consists of a set of field-value pairs, 
similar to a Redis hash. How does the message payload differ from a regular Redis hash?
```
A stream message's payload is organized as a set of field-value pairs which resemble a 
Redis hash.  However, when reading from a stream the entire message payload hash will 
always be returned to the client; there is no equivalent to the HGET or HMGET commands.
```

* 2 Suppose you've created a stream using these commands:
```
XADD numbers 1-0 n 0
XADD numbers 2-0 n 1
```
Right now, `XLEN numbers returns 2`.
Using a single command, you next need to update the stream in a way that will make 
XLEN numbers return 0. Which of the following commands does accomplish this goal?
```
XTRIM numbers MAXLEN 0
XDEL numbers 1-0 2-0
DEL numbers
```
``` 
To remove messages from the stream, the XDEL and XTRIM commands can be used. 
As streams are treated like any other data type in Redis, 
the DEL command can be used to remove the entire stream.
```

* 3 Suppose you've created a stream containing 4 messages using the following commands:
```
XADD numbers 1-0 n 0
XADD numbers 2-0 n 1
XADD numbers 3-0 n 2
XADD numbers 4-0 n 3
```
Which commands can you use to return only messages 2-0 and 3-0, in any order?
```
XRANGE numbers 1-1 + COUNT 2
XREVRANGE numbers 3-0 2-0
```
```
Returns messages 3-0 and 2-0 as the range of IDs supplied to XRANGE and XREVRANGE 
is inclusive, and those are the only two messages in that range.
```

* 4 Which two features distinguish the XREAD command from the XRANGE command?
```
XREAD can read from multiple streams at once, whereas XRANGE is limited to one
XREAD can retrieve new messages by providing the ID of the last message read from the 
stream; XRANGE cannot
```

* 5 Suppose you've created a stream containing 4 messages using the following commands:
```
XADD numbers 1-0 n 0
XADD numbers 2-0 n 1
XADD numbers 3-0 n 2
XADD numbers 4-0 n 3
```
You then run the following command:
```
XREAD BLOCK 5000 STREAMS numbers $
```
And immediately, in a separate window, you run:
```
XADD numbers 5-0 n 4
XADD numbers 6-0 n 5
```
What will the XREAD command return?
```
Only message 5-0 correct
```

## Part 3
* 1 You've created a consumer group called "primes" which reads from the stream "numbers". 
Each consumer in this group must have a unique name. What is the scope of this uniqueness?

* 2 You have a stream named "numbers" containing 4 messages whose 
IDs are 1-0, 2-0, 3-0, and 4-0. 

You then create a consumer group as follows:
```
XGROUP CREATE numbers primes 0
```
Two consumers join the group and read messages:
```
XREADGROUP GROUP primes consumer1 COUNT 2 STREAMS numbers >
XREADGROUP GROUP primes consumer2 COUNT 3 STREAMS numbers >
```
Which messages did consumer2 read?
```
3-0 and 4-0
```
```
Members of a consumer group work together to process the stream. The consumer group's 
members will, as a whole, read all messages. An individual consumer in a group will 
only see a subset of the stream's messages.In this case, consumer1 first reads messages 1-0 and 2-0.
Next consumer2 reads up to three messages that the consumer group as a whole had not yet 
processed. Therefore, it receives messages 3-0 and 4-0.
```

* 3 When consumers read a stream with XREADGROUP, Redis maintains a Pending Entries List. 
What is the purpose of this list?
```
To track messages which have been delivered to a consumer but not yet acknowledged by it
```
```
The Redis server uses a Pending Entries List (PEL) to store the IDs of those messages 
read by consumers with XREADGROUP, but not yet acknowledged with XACK. 
This workflow allows messages to be tracked and potentially reassigned in the event that 
a consumer crashes while processing them.

Using the NOACK subcommand with XREADGROUP will prevent messages from being added to the PEL. 
In this case, all messages read by consumers will be considered automatically acknowledged 
regardless of whether the consumer successfully processes them.
```

* 4 Suppose you run the XPENDING command and see the following output:
```
XPENDING numbers primes
1) (integer) 4
2) "1-0"
3) "4-0"
4) 1) 1) "consumer1"
      2) "2"
   2) 1) "consumer2"
      2) "2"
```
You notice that consumer2 has 2 pending messages. 
Which command will tell you the message IDs of those pending messages without 
also including the message payload?



