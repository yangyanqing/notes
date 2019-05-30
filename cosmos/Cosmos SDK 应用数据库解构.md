# Cosmos SDK 应用数据库解构



## 1. 前言

[Cosmos](https://cosmos.network/) 是一个独立并行区块链的分散网络，每个区块链都由BFT共识算法（如Tendermint共识）提供支持。[Cosmos SDK](https://cosmos.network/sdk) 是一个通用框架，内置了 Tendermint 共识引擎和一些基础设施（如存储、账户、治理等）。基于 Cosmos SDK 可以快速构建自己的 PBFT 共识区块链。

ToDo

## 2. 环境准备

### 2.1 准备 Golang 环境

需要保证 Go 的版本在 1.12 以上，下载 [Go 1.12.3+](https://golang.org/dl)

Golang 安装文档：

1. [https://golang.org/doc/install](https://golang.org/doc/install)
2. [https://github.com/golang/go/wiki/Ubuntu](https://github.com/golang/go/wiki/Ubuntu)

此外，你需要指定运行 Go 所需的 `$GOPATH`、`$GOBIN` 和 `$PATH` 变量, 例如:

```bash
mkdir -p $HOME/go/bin
echo 'export GOPATH=$HOME/go' >> ~/.bash_profile
echo 'export GOBIN=$GOPATH/bin' >> ~/.bash_profile
echo 'export PATH=$PATH:$GOBIN' >> ~/.bash_profile
source ~/.bash_profile
```

### 2.2 编译&安装 nameservice

为简化模型，使用 cosmos-sdk 教程 nameservice 数据库作为分析样本。按照如下命令

```bash
cd ${GOPATH}/src/github.com/cosmos
git clone https://github.com/cosmos/sdk-application-tutorial.git
cd sdk-application-tutorial
git checkout 50c050185505c2ad98c0c1a1fed92674256f7641
make install
```

### 2.3 初始化数据

1. 运行如下脚本初始化链以及测试账户。

    ```bash
    #!/bin/bash
    
    CHAIN_ID=namechain
    PASSWORD=11111111
    
    rm -rf ~/.nsd/ ~/.nscli/
    
    nscli config chain-id ${CHAIN_ID}
    nscli config output json
    nscli config indent true
    nscli config trust-node true
    
    nsd init --chain-id ${CHAIN_ID}
    echo "${PASSWORD}" | nscli keys add jack
    echo "${PASSWORD}" | nscli keys add alice
    
    nsd add-genesis-account $(nscli keys show jack -a)  1000nametoken,1000jackcoin
    nsd add-genesis-account $(nscli keys show alice -a) 1000nametoken,1000alicecoin
    ```

    

2. 运行 `nsd start`命令， 生成两个块后停止，此时在 `~/.nsd/data/` 目录下会生成多个数据库目录，其中 `application.db` 为本文将要分析的目标数据库

    ```bash
    ➜  data $ pwd
    /Users/yangyanqing/.nsd/data
    ➜  data $ ls -1
    application.db
    blockstore.db
    cs.wal
    evidence.db
    priv_validator_state.json
    state.db
    tx_index.db
    ➜  data $
    ```

    

3. 亦可直接下载[压缩包](./attachment/nsdb.zip)后解压，结果同2。

## 3. 数据库分析

