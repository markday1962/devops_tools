# Final Exam
## Part 1
* 1 What benefits can Redis streams bring to a distributed application architecture?
```
A standard approach for managing communications between heterogeneous system components
The ability to provide buffers against surges in system activity
The ability to easily perform time-based range queries
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
* 5 Which of the following commands will create the stream "numbers" if it does not already exist?
```
XADD numbers * hello world
XGROUP CREATE numbers primes $ MKSTREAM
```

## Part 2
