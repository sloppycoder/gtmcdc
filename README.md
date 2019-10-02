## GT.M/YottaDB Replication Filter 

A GT.M/YottaDB change event capture (CDC) mechansm by using a replication filter that convert M instructions into Kafka messages for the purpose of propagating change event to external systems.

The code here is just a skeleton. A lot more business logic needs to be implemented before it can be useful.

The scripts in [repl_procedures](repl_procedures) directory are copied from [YottaDB documentation](https://gitlab.com/YottaDB/DB/YDBDoc/tree/master/AdminOpsGuide/repl_procedures). Shout out to the great work by [YottaDB team](https://yottadb.com/)

### System Requirements
Tested with:
1. [Ubuntu 18.04](http://releases.ubuntu.com/18.04/)
2. [YottaDB r1.28](https://yottadb.com/product/get-started/)
3. [Go 1.13](https://golang.org/dl/)

### Build
```
go build ./cmd/cdcfilter

```

### Test with replication setup

First, [install YottaDB](https://yottadb.com/product/get-started/), then follow the steps below to create source database A and target database B, start the replicating processes.

```
cd repl_procedures

# setup 2 databases, A and B 
./dbinit A
./dbinit B

# start replication processes 
export GTMCDC_KAFKA_BROKERS=off
./repl_start A B

# check replication filter log
# the log should show the filter started with no errors 
tail -f filter.log



```

Then write some data in A, and check them in B. Watch the replication filter log too.

```
# open another shell prompt
./db A

# then enter some M statement that change data, i.e.
YDB> set ^TEST("123")=456.00

# at this time the filter log should show the the transaction
# then check if the data is replicated to B
./db B

# then enter some M statement that change data, i.e.
YDB> write ^TEST("123")

# should print 456.00

```

To tear down the setup and cleanup everything
```
./repl_stop

# run this only if you want to delete both A and B databases
rm -rf A B filter.log

```
