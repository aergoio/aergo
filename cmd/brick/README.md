
# Brick - Toy for Contract Developers

This is an interactive shell program to communicate with an aergo vm for testing.
This also provides a batch function to test and help to develop smart contrat.

![brick_ex_gif](./brick_ex.gif)

## Features

### Aergo VM for Testing

* provides an way to run smart contracts in same environment with aergo
* supports state db, sql
* alias for account and contract address

### Interactive Shell

powered by [go-prompt](https://github.com/c-bata/go-prompt)

* auto complete
* keywords suggestion
* history

### Batch for Integration Tests

* load and execute bulk of commands from file
* contract integration test

## Install

1. git clone and build aergo, unix & mac: `cmake . && make`, windows + mingw64: `cmake -G "Unix Makefiles" -D CMAKE_MAKE_PROGRAM=mingw32-make.exe . && mingw32-make`
2. create a directory, where you want
3. copy binary `brick` and log config file `cmd/brick/arglog.toml` to the directory
4. move to the directory
5. run `brick`

## Usage

### inject

creates an account and deposite aergo. `inject inject <account_name> <amount>`

``` lua
0> inject tester 100
  INF inject an account successfully cmd=inject module=brick
```

### getstate

get current balance of an account. `getstate getstate <account_name>` This will returns in a form of `<internal_address>=<remaining_balance>`

``` lua
1> getstate tester
  INF Amh8FdZavYGu9ABhZQF6Y7LnSMi2a714bGvpm6mQuMnRpmoYn11C=100 cmd=getstate module=brick
```

### send

transfer aergo between two addresses. `send <sender_name> <receiver_name> <amount>`

``` lua
1> send tester receiver 10
  INF send aergo successfully cmd=send module=brick
```

### deploy

deploy a smart contract. `deploy <sender_name> <fee_amount> <contract_name> <definition_file_path>`

``` lua
2> deploy tester 1 helloContract `../example/hello.lua`
  INF deploy a smart contract successfully cmd=deploy module=brick
```

### call

call to execute a smart contract. `call <sender_name> <amount> <contract_name> <func_name> <call_json_str>`

``` lua
3> call tester 1 helloContract set_name `["aergo"]`
  INF call a smart contract successfully cmd=call module=brick
```

### query

query to a smart contract. `query <contract_name> <func_name> <query_json_str> [expected_query_result]`

``` lua
4> query helloContract hello `[]`
  INF "hello aergo" cmd=query module=brick
  ```

`expected_query_result`, surrounded by [], is optional. But you can test whether the results are correct using this.

  ``` lua
4> query helloContract hello `[]` `"hello aergo"`
  INF query compare successfully cmd=query module=brick
4> query helloContract hello `[]` `"hello incorrect example"`
  ERR execution fail error="expected: \"hello incorrect example\", but got: \"hello aergo\"" cmd=query module=brick
```

### batch

keeps commands in a text file and use at later. `batch <batch_file_path>`

``` lua
5> batch `../cmd/brick/example/hello.brick`
  INF inject an account successfully cmd=inject module=brick
  INF 100 cmd=getstate module=brick
  INF deploy a smart contract successfully cmd=deploy module=brick
  INF query compare successfully cmd=query module=brick
  INF call a smart contract successfully cmd=call module=brick
  INF query compare successfully cmd=query module=brick
  INF batch exec is finished cmd=batch module=brick
```

### undo

cancels the last tx (inject, send, deploy, call). `undo`

``` lua
8> undo
  INF Undo, Succesfully cmd=undo module=brick
```

### reset

clear all txs and reset the chain. `reset`

``` lua
7> reset
  INF reset a dummy chain successfully cmd=reset module=brick
0>
```

Number before cursor is a block height. Each block contains one tx. So after reset, number becames 0

### batch in command line

In command line, users can run a brick batch file.

``` bash
$ ./brick ./example/hello.brick
  INF inject an account successfully cmd=inject module=brick
  INF 100 cmd=getstate module=brick
  INF deploy a smart contract successfully cmd=deploy module=brick
  INF query compare successfully cmd=query module=brick
  INF call a smart contract successfully cmd=call module=brick
  INF query compare successfully cmd=query module=brick
  INF batch exec is finished cmd=batch module=brick
```