ripple-mngr
===========

A ripple key generator & transaction watcher.

Compilation
-----------

Make sure `golang` and `git` are installed:

```shell
$ sudo apt-get install golang-go git
```

Then, install `ripple-mngr` wherever you want on the filesystem:

```shell
$ export GOPATH=${HOME}/go
$ git clone https://gitlab.mkz.me/mycroft/ripple-mngr
$ cd ripple-mngr
$ go get -d -v
$ go build
$ ./ripple-mngr -h
Usage of ./ripple-mngr:
  -config string
        Configuration file (default "config.ini")
  -debug
        Debug mode
  -file string
        File for export
  -init
        DB Init
  -refresh
        Refresh data from database
  -status
        Show key statuses
  -watch
        Search for transactions for existing addresses
```

Configuration
-------------

`ripple-mngr` is using a `config.ini` INI file for configuration:

```ini
[general]
debug = false
file = ./private-keys
# Main: https://s1.ripple.com:51234
api_url = https://s.altnet.rippletest.net:51234

[keys]
# Number of keys in pool
num = 10

[db]
# Set true to disable DB
# disabled = true

# DB Configuration
host = 172.17.0.2
name = mysql
user = root
pass = pass
```

Note that there is config.ini.sample ready for use in the repository.

## Create tables (initialization)

```shell
$ ./ripple-mngr -debug -init
2018/04/06 07:51:52 Key pool size is 3.
2018/04/06 07:51:52 Using API url: https://s.altnet.rippletest.net:51234
2018/04/06 07:51:52 Using file: ./private-keys
2018/04/06 07:51:52 Tables created.
```

## Generate new keys, if needed

```shell
$ ./ripple-mngr -debug 
2018/04/06 07:51:54 Key pool size is 3.
2018/04/06 07:51:54 Using API url: https://s.altnet.rippletest.net:51234
2018/04/06 07:51:54 Using file: ./private-keys
2018/04/06 07:51:54 Required to create 3 new keys.
2018/04/06 07:51:54 Seed: shYTLrgshHtpa8SaimVKYGHREsBPH
2018/04/06 07:51:54 Pub: rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6
2018/04/06 07:51:54 DB Store: rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6
2018/04/06 07:51:54 DB Store: returned id:1
2018/04/06 07:51:54 Seed: sn3ok2ZfP9ZanvRhcZU6jHyzEgwYJ
2018/04/06 07:51:54 Pub: ryz4JEJqEBQryMHw4qVavEMvYbYqiDt9o
2018/04/06 07:51:54 DB Store: ryz4JEJqEBQryMHw4qVavEMvYbYqiDt9o
2018/04/06 07:51:54 DB Store: returned id:2
2018/04/06 07:51:54 Seed: ssemKFArwu7KFesKLFWkEPPuSybx5
2018/04/06 07:51:54 Pub: rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1
2018/04/06 07:51:54 DB Store: rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1
2018/04/06 07:51:54 DB Store: returned id:3
$
```

It will store in database:

```sql
MariaDB [mysql]> select id, pub, tx_metadata, tx_value, received, used, completed from xrpkeys;
+----+------------------------------------+-------------+----------+----------+------+-----------+
| id | pub                                | tx_metadata | tx_value | received | used | completed |
+----+------------------------------------+-------------+----------+----------+------+-----------+
|  1 | rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6 | NULL        | 0        | 0        |    0 |         0 |
|  2 | ryz4JEJqEBQryMHw4qVavEMvYbYqiDt9o  | NULL        | 0        | 0        |    0 |         0 |
|  3 | rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1 | NULL        | 0        | 0        |    0 |         0 |
+----+------------------------------------+-------------+----------+----------+------+-----------+
3 rows in set (0.00 sec)

$
```

## Show key statuses

```shell
$ ./ripple-mngr -status
2018/04/06 07:56:03 id:1 rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6 used:false waited:0 received:0 
2018/04/06 07:56:03 id:2 rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1 used:false waited:0 received:0 
2018/04/06 07:56:03 id:3 rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1 used:false waited:0 received:0 
$
```

## Watch for transactions

By default, `ripple-mngr` will not watch for transactions if transaction is:
* not used (used = false in database);
* started_ts < NOW() - 24h (it will start to watch only on addresses that were asked to be looked the last 24h, no more, to avoid to query too long the API.);
* completed (completed = true in database);
* received >= tx_value (won't watch any more if balance is bigger than waited value).

Therefore, the `upstream app` must inform `ripple-mngr` to watch for transaction using the following query, updating `used`, `tx_value` and `started_ts` fields:

```sql
MariaDB [mysql]> update xrpkeys set used = true, tx_value = 10000000000, started_ts = NOW() where pub = 'rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1';
Query OK, 1 row affected (0.01 sec)
Rows matched: 1  Changed: 1  Warnings: 0
```

Status will then show:

```shell
$ ./ripple-mngr -status
2018/04/06 07:59:05 id:1 rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6 used:false waited:0 received:0 
2018/04/06 07:59:05 id:2 rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1 used:true waited:10000000000 received:0 started_ts:'2018-04-06 05:58:46'
2018/04/06 07:59:05 id:3 rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1 used:false waited:0 received:0 
$
```

Note that once used, the address is no longer in the pool and calling `ripple-mngr` without any flag will create a new address:

```shell
$ ./ripple-mngr -debug
2018/04/06 07:59:30 Key pool size is 3.
2018/04/06 07:59:30 Using API url: https://s.altnet.rippletest.net:51234
2018/04/06 07:59:30 Using file: ./private-keys
2018/04/06 07:59:30 Required to create 1 new keys.
2018/04/06 07:59:30 Seed: ssPpc7L8RePg3cWQTPnWea4xRZTMj
2018/04/06 07:59:30 Pub: rEfR6Yw8axEE9Qg9ReQCkYx7Hovj4BBP2K
2018/04/06 07:59:30 DB Store: rEfR6Yw8axEE9Qg9ReQCkYx7Hovj4BBP2K
2018/04/06 07:59:30 DB Store: returned id:4
$
```

The watcher will, for each addresses that must be watched:
* Query the API to check for `balance`;
* If `balance` differs that the one in database, it will store the new `balance`;
* It will query the API to retrieve all transactions associated to this address, and store them in database. Once in database, those transactions are no longer used by `ripple-mngr`.

To run the watcher, just use the `-watch` flag:

```shell
$ ./ripple-mngr -debug -watch
2018/04/06 08:00:01 Key pool size is 3.
2018/04/06 08:00:01 Using API url: https://s.altnet.rippletest.net:51234
2018/04/06 08:00:01 Using file: ./private-keys
2018/04/06 08:00:01 Watch()
2018/04/06 08:00:01 Looking for key rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1
2018/04/06 08:00:01 Current: pub:rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1 received:0
2018/04/06 08:00:01 Storing new received value (10000000000) in database.
2018/04/06 08:00:01 UpdateValue pub:rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1 value:10000000000 completed without error
2018/04/06 08:00:01 Query for TX for rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1
2018/04/06 08:00:01 DB StoreTXs: 1 records
2018/04/06 08:00:01 DB StoreTXs: Storing tx hash:DA6F578C70B52ADB606B7EA4592D86C0B06106EA472A679E123BAC32E09F4599
2018/04/06 08:00:01 DB StoreTXs: returned id:1
2018/04/06 08:00:01 DB StoreTXs Done.
2018/04/06 08:00:01 Watch() done successfully!
$
```

Status will then show:

```shell
$ ./ripple-mngr -status
2018/04/06 08:00:49 id:1 rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6 used:false waited:0 received:0 
2018/04/06 08:00:49 id:2 rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1 used:true waited:10000000000 received:10000000000 started_ts:'2018-04-06 05:58:46'
2018/04/06 08:00:49 id:3 rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1 used:false waited:0 received:0 
2018/04/06 08:00:49 id:4 rEfR6Yw8axEE9Qg9ReQCkYx7Hovj4BBP2K used:false waited:0 received:0 
$
```

And in DB, transactions will be stored (the database schema for this table can be found at the end of this page):

```mysql
MariaDB [mysql]> select hash, from_addr, amount from xrptx where to_addr = 'rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1';
+------------------------------------------------------------------+------------------------------------+-------------+
| hash                                                             | from_addr                          | amount      |
+------------------------------------------------------------------+------------------------------------+-------------+
| DA6F578C70B52ADB606B7EA4592D86C0B06106EA472A679E123BAC32E09F4599 | rPT1Sjq2YGrBMTttX4GZHjKu9dyfzbpAYe | 10000000000 |
+------------------------------------------------------------------+------------------------------------+-------------+
1 row in set (0.01 sec)
```

When balance is reached, the address won't be watched anymore:

```shell
$ ./ripple-mngr -debug -watch
2018/04/06 08:03:37 Key pool size is 3.
2018/04/06 08:03:37 Using API url: https://s.altnet.rippletest.net:51234
2018/04/06 08:03:37 Using file: ./private-keys
2018/04/06 08:03:37 Watch()
2018/04/06 08:03:37 No record to look after.
2018/04/06 08:03:37 Watch() done successfully!
$
```

At this moment, the `upstream app` can mark the address as completed:

```mysql
MariaDB [mysql]> update xrpkeys set completed = true where pub = 'rMHBFRZZKTHs2zj5Rz1ot5kLz3i3tT7iy1';
Query OK, 1 row affected (0.01 sec)
Rows matched: 1  Changed: 1  Warnings: 0
```

The `completed` addresses won't be reported in status anymore:

```shell
$ ./eth-generator -debug -status
2018/04/06 08:04:13 id:1 rByPDe61QhiN3WYuXjdrm9DPAG6VEsyGp6 used:false waited:0 received:0 
2018/04/06 08:04:13 id:3 rMwfkCRi4QL6U6JZWYSgLJ27nmAgywH9s1 used:false waited:0 received:0 
2018/04/06 08:04:13 id:4 rEfR6Yw8axEE9Qg9ReQCkYx7Hovj4BBP2K used:false waited:0 received:0 
$
```
