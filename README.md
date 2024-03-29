## GT.M/YottaDB Replication Filter

A GT.M/YottaDB change event capture (CDC) mechansm by using a replication filter that convert M instructions into Kafka messages for the purpose of propagating change event to external systems.

The code here is just a skeleton. A lot more business logic needs to be implemented before it can be useful.

The scripts in [scripts/original_from_ydb_doc](scripts/original_from_ydb_doc) directory are copied from [YottaDB documentation](https://gitlab.com/YottaDB/DB/YDBDoc/tree/master/AdminOpsGuide/repl_procedures). Shout out to the great work by [YottaDB team](https://yottadb.com/)

### System Requirements

Tested with:

1. [Ubuntu 18.04](http://releases.ubuntu.com/18.04/)
2. [Go 1.13](https://golang.org/dl/)
3. [Confluent Platform 5.3](https://www.confluent.io/download/)
4. [MongoDB server](https://www.mongodb.com/download-center/community)
5. [MongoDB Kafka connector](https://www.confluent.io/hub/mongodb/kafka-connect-mongodb)
6. [YottaDB r1.28](https://yottadb.com/product/get-started/)
7. [FIS GT.M](https://en.wikipedia.org/wiki/GT.M)

### Build

```
go build ./cmd/cdcfilter

```

### Setup the environment

1. Install Confluent Platform along with confluent cli. Set environment variable CONFLUENT_HOME to the installation directory.
2. Install MongoDB kafka sink compoent from ConfluentHub.
3. Install MongoDB server
4. Create a user dev with password dev, by running the script below using mongo shell

### Test cdcfilter using Journal file without YottaDB setup

You should probably start here.

Install and set the prerequisites.

1. MongoDB 4.2 Community Edition
2. Confluent Platform or Community edition
3. Setup a user dev with password dev who owns the dev database, using something like below

```
use dev
db.createUser(
  {
    user: "dev",
    pwd:  "dev",
    roles: [ { role: "readWrite", db: "dev" } ]
  }
)

```

The follow the steps to setup KSQL streams and Sink in Kafka.

```

# install confluent kafka and confluent CLI, set the $CONFLUENT_HOME environment variable.
# if you install using the RPM or DEB file, the kafka utilities are in your PATH, so the $CONFLUENT_HOME/bin
# below can be omitted

# startup confluent platform
$CONFLUENT_HOME/bin/confluent local start

cd scripts/kafka

# create the topics and streams
$CONFLUENT_HOME/bin/ksql

ksql> RUN SCRIPT setup.ksql

# should print bunch of stuff but no error messages.
# ctrl-D to exit ksql prompt
# at this point 2 KSQL streams should have been setup and along with their topics

# setup the MongoDB kakfa sink
./create_sink

# at this point all things Kafka should be runing. If you have full Confluent Platform, you can check things out using the Web console.

# now run the cdcfilter with journal file
cd ../..
./cdcfilter -i testdata/t.txt > x.out

# data should be in MongoDB now
mongo -u dev -p dev dev
>db.accounts.find()

# you should see 2 documents here.
# if you don't see it, check your kafka log and cdcfilter.log to see if there're any error messages

```

### Test with YottaDB or GT.M replication setup

Install GT.M by running the ```apt install fis-gtm``` or [YottaDB](https://yottadb.com/product/get-started/).
Then follow the steps below to create source database A and target database B, start the replicating processes.

```
cd scripts/ydb 
# or 
cd scripts/gtm

# setup 2 databases, A and B
./dbinit A
./dbinit B

# start replication processes
./repl_start A B

# check replication filter log
# the log should show the filter started with no errors
tail -f cdcfilter.log



```

Write test data.
Run the acc.m script to write data to YottaDB site A, the replication filter should receive the writes and create events in Kafka, which will be processed by MongoDB kafka connector sink component, data will be written to MongoDB database dev and collection accounts.

```
# copy test data
cp acc.m A/.
$gtm_dist/mumps -r acc

# open mongo shell and check data db.accounts.find()

# edit the test data
vi A/acc.m
# save data, and run the updated program
$gtm_dist/mumps -r acc


```

To tear down the setup and cleanup everything

```
./repl_stop

# run this only if you want to delete both A and B databases
rm -rf A B cdcfilter.log

```
