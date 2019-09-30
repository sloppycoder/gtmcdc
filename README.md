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
go build -o filter ./cmd

```

### Test with replication setup

First, [install YottaDB](https://yottadb.com/product/get-started/), then follow the steps below to create source database A and target database B, start the replicating processes.

```
cd repl_procedures

# setup A site as source
source ./ydbenv A r128
./db_create
./repl_setup

# start replication process at source site
./originating_start A B 4001

# set up B site as receiver site
source ./ydbenv B r128
./db_create
./repl_setup

# start replication receiver
# in this default setup the Kafka publishing 
# and Prometheus listener are disabled
export GTMCDC_DEVMODE=1
./replicating_start_with_filter B 4001
tail -f B/filter.log

# the log should show the filter started with no errors 


```

Then write some data in A, and check them in B. Watch the replication filter log too.

```
# open another shell prompt
source ./ydbenv A r128
/usr/local/lib/yottadb/r128/ydb

# then enter some M statement that change data, i.e.
set ^TEST("123")=456.00

# at this time the filter log should show the the transaction
# then check if the data is replicated to B
source ./ydbenv B r128
/usr/local/lib/yottadb/r128/ydb

# then enter some M statement that change data, i.e.
set X=^TEST("123")
write X

# should print "456.00"


```

To tear down the setup and cleanup everything
```
source ./ydbenv A r128
./originating_stop

source ./ydbenv B r128
./replicating_stop

# run this only if you want to delete both A and B databases
rm -rf A B

```
