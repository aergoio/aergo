
# Brick - Toy for Contract Developers

This is an interactive shell program to communicate with an aergo vm for testing.
This also provides a batch function to test and help to develop smart contrat.

![brick_ex_gif](./brick_ex.gif)

## Features

### Aergo VM for Testing

* provides an way to run smart contracts in same environment with aergo
* supports state db, sql
* alias for account and contract address
* able to debug smart contracts

### Interactive Shell

powered by [go-prompt](https://github.com/c-bata/go-prompt)

* auto complete
* keywords suggestion
* history

### Batch for Integration Tests

* load and execute bulk of commands from file
* contract integration test

## Install

1. git clone and build aergo in debug mode (release is ok, but you cannot debug contracts), unix & mac: `make debug`, windows + mingw64: `mingw32-make debug`
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

deploy a smart contract. `deploy <sender_name> <fee_amount> <contract_name> <definition_file_path> [contructor_json_arg]`

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


## Debugging

If you build in debug mode, you can use `os, io, coroutine` modules which is not allowed in release mode. There is no limit to which debugger to use, but here we describe the zerobrane studio, which provides ui and is easy to install.

1. download [zerobrane studio](https://studio.zerobrane.com/support) (lua ide)
2. set envoronment var `LUA_PATH, LUACPATH` following an [instruction](https://studio.zerobrane.com/doc-remote-debugging)
3. Paste `require('mobdebug').start()` at the beginning of a smart contract, you want to investigate
4. Open the contract file you want to investigate in the editor
5. In the editor, run debug server `project -> start debugger server`
6. Using brick, deploy and run  the contract
7. When you get to that line, it will automatically be switched to the editor

 (CAUTION!) After testing, When distributing to an actual blockchain, you must remove the code used for debugging. It will cause error.