
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
* support incremental test

## Install

1. git clone and build aergo in debug mode (release is ok, but you cannot debug contracts), unix & mac: `make debug`, windows + mingw64: `mingw32-make debug`
2. create a directory, where you want
3. copy binary `brick` to the directory, or add aergo's `bin` folder to system PATH
4. move to the directory
5. run `brick`
6. (optional) copy log config file `cmd/brick/arglog.toml` to the directory, and adjust levels to get for more detailed info

## Usage

### inject

creates an account and deposite aergo. `inject inject <account_name> <amount>`

``` lua
0> inject tester 100
  INF inject an account successfully cmd=inject module=brick
```

### getstate

get current balance of an account. `getstate getstate <account_name> [expected_balance]` This will returns in a form of `<internal_address>=<remaining_balance>`

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
2> deploy tester 0 helloContract ./example/hello.lua
  INF deploy a smart contract successfully cmd=deploy module=brick
```

deploy a raw text file on http also available.

``` lua
2> deploy tester 0 helloContract https://raw.githubusercontent.com/aergoio/aergo-contract-ex/master/contracts/helloworld/test/test-helloworld.brick
  INF deploy a smart contract successfully cmd=deploy module=brick
```

### call

call to execute a smart contract. `call <sender_name> <amount> <contract_name> <func_name> <call_json_str> [expected_error]`

``` lua
3> call tester 0 helloContract set_name `["aergo"]`
  INF call a smart contract successfully cmd=call module=brick
```

### query

query to a smart contract. `query <contract_name> <func_name> <query_json_str> [expected_query_result] [expected_error]`

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

keeps commands in a text file (local or http) and use at later. `batch <batch_file_path>`

``` lua
4> batch ./example/hello.brick
Batch is successfully finished
8> 
```

### undo

cancels the last tx (inject, send, deploy, call). `undo`

``` lua
8> undo
  INF Undo, Succesfully cmd=undo module=brick
```

### forward

skip blocks. `forward [height_to_skip]`

``` lua
7> forward 100
  INF fast forward blocks successfully cmd=forward module=brick
```

### reset

clear all txs and reset the chain. `reset`

``` lua
107> reset
  INF reset a dummy chain successfully cmd=reset module=brick
0>
```

Number before cursor is a block height. Each block contains one tx. So after reset, number becames 0

### batch in command line

In command line, users can run a brick batch file. A running result contains line numbers and original texts for debugging purpose.

``` bash
$ ./brick ./example/hello.brick
Batch is successfully finished
  INF batch exec is finished cmd=batch module=brick
```
User can check the detail results by setting `-v` options.

``` bash
$ ./brick ./example/hello.brick -v
1 # create an account and deposit coin
2 inject bj 100
  INF inject an account successfully cmd=inject module=brick
3
4 # check balance
5 getstate bj
  INF AmgiDbXB3x5Zv5cBBcmyzSJiqMbLXG1yZN2xE4gUuXMBaXPfGoun = 100 cmd=getstate module=brick
6
7 # delpoy helloworld smart contract
8 deploy bj 0 helloctr `./example/hello.lua`
  INF deploy a smart contract successfully cmd=deploy module=brick
9
10 # query to contract, this will print "hello world"
11 query helloctr hello `[]` `"hello world"`
  INF query compare successfully cmd=query module=brick
12
13 # call and execute contract
14 call bj 0 helloctr set_name `["aergo"]`
  INF call a smart contract successfully cmd=call module=brick
15
16 # query again, this now will print "hello aergo"
17 query helloctr hello `[]` `"hello aergo"`
  INF query compare successfully cmd=query module=brick
Batch is successfully finished
  INF batch exec is finished cmd=batch module=brick
```
Or user can set the option `-w` to display the batch execution results continuously according to the file changes. This is an useful feature for the development phase.

## Debugging

If you build in debug mode (`make debug`), you can use `os, io, debug` modules which is not allowed in release mode. There is no limit to which debugger to use, but brick provides built-in debugger using customized [clidebugger](https://github.com/ToddWegner/clidebugger). For debugging purpose, brick has extended commands.

### setb (brick / debugmode)

Set a breakpoint to the contract. When vm reach the line of breakpoint during a call or query of a contract, it enters debugmode only. contract_name is optional in debugmode. `setb <line> [contract_name]`

### delb (brick / debugmode) 

Delete an existing breakpoint. contract_name is optional in debugmode. `delb <line> [contract_name]`

### listb (brick / debugmode)

Prints all breakpoints. `listb`

### resetb (brick / debugmode)

Clear all breakpoints. `resetb`

### setw (brick / debugmode)

Set an watchpoint expression. If one of watchpoint expressions is satisfied without regard to which contract is being executed, debug mode is activated. `setw <watch_expression>`

### delw (brick / debugmode)

Delete an existing watchpoint. `delw <watch_index>`

### listw (brick / debugmode)

Prints all watchpoints. `listw`

### resetw (brick / debugmode)

Clear all watchpoints. `resetw`

### in debugmode

When vm enters debugmode, prompt changes to `[DEBUG]>`. In debugmode, command set is changed for debugging purpose, like `run`, `exit`, `show`, `vars`. For more detail, type `help`.

## Debug using Zerobrane Studio

Here we describe GUI based debugging using the zerobrane studio.

1. download [zerobrane studio](https://studio.zerobrane.com/support) (lua ide)
2. set envoronment var `LUA_PATH, LUACPATH` following an [instruction](https://studio.zerobrane.com/doc-remote-debugging)
3. Paste `require('mobdebug').start()` at the beginning of a smart contract, you want to investigate
4. Open the contract file you want to investigate in the editor
5. In the editor, run debug server `project -> start debugger server`
6. Using brick, deploy and run  the contract
7. When you get to that line, it will automatically be switched to the editor

 (CAUTION!) After testing, When distributing to an actual blockchain, you must remove the code used for debugging. It will cause error.