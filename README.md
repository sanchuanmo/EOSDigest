# EOSDigest

## 项目背景
本项目是EOS区块链的区块交易的hash实现，作为实现EOS区块链默克尔证明的前置实现。
由于go中采用的sha256算法难以复现EOS区块链源码中的sha256算法，且go版实现的openssl库与c++原版库sha256的hash结果不一致，故选取openssl库（c++）作为sha256算法的实现底层接口。

## 环境依赖

* [eosio_2.1.0-1](https://github.com/eosio/eos/releases/download/v2.1.0/eosio_2.1.0-1-ubuntu-18.04_amd64.deb)
* [eosio.cdt v1.7.x](https://github.com/EOSIO/eosio.cdt/releases/tag/v1.7.0)
* [openssl](https://github.com/openssl/openssl/releases/tag/OpenSSL_1_1_1t)


## 编译步骤

### 前置条件
确保安装了适当版本的`eosio`区块链。如需安装`eosio`区块链，请按照官网教程进行安装 [详细说明步骤](https://developers.eos.io/welcome/latest/getting-started-guide/local-development-environment/installing-eosio-binaries)进行。

确保安装了适当版本的`openssl`。如需安装`openssl`，请按照官网教程进行安装 [详细说明步骤](https://github.com/openssl/openssl#build-and-install)进行。要验证是否安装适当版本，请运行一下命令:
```bash
openssl version -a
```

### 测试函数

```sh
go test -run TestMerkleTree

res: 3af423e81b49e33686b446d8e8d46e9a46f455de2d164483a8ff900558f59304   //块内交易1
res: d48880c9c13c9ff17daa5434b98513ce1997e6fe9dad0b391486cb97fb52d422   //块内交易2
enc.Sum(nil): e14869d99fdcf8944ac67d0f1f30fe9fb8ccaa609bb58caf8621a536fe4865ca  //计算的默克尔根
```

### PS

你已成功验证该工具类的可行性。
