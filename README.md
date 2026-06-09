# Web3 Multi-Chain Query API

多链 Web3 数据查询服务，支持 ETH、BNB、Tron、Solana 四条链，可灵活扩展。

## 功能

- 查询指定地址拥有的代币余额
- 查询本币（ETH/BNB/TRX/SOL）交易记录
- 查询指定代币的交易记录
- 查询地址拥有的 NFT（自动转换 IPFS/Arweave 链接）
- 实时本币 USD 价格 + 24h 涨跌幅（多源聚合 + 缓存）
- 主网 / 测试网双环境支持

## 支持的链

| 链 | 主网 | 测试网 | 数据源 |
|---|---|---|---|
| Ethereum | `eth` | `eth_sepolia` | Moralis |
| BNB Chain | `bnb` | `bnb_testnet` | Moralis |
| Tron | `tron` | `tron_shasta` | TronGrid |
| Solana | `solana` | `solana_devnet` | Helius |

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置

```bash
cp config.example.yaml config.yaml
```

编辑 `config.yaml`，填入各链的 API Key：

- **Moralis**（ETH/BNB）：https://admin.moralis.io
- **TronGrid**（Tron）：https://www.trongrid.io
- **Helius**（Solana）：https://helius.dev

### 3. 启动

```bash
go run ./cmd/server
```

服务默认运行在 `http://localhost:8080`

### 4. Swagger 文档

启动后访问：http://localhost:8080/swagger/index.html

## API 接口

### 代币查询

```
GET /api/v1/chain/{chain}/tokens?address={addr}&cursor={cursor}
```

### 本币交易记录

```
GET /api/v1/chain/{chain}/transactions?address={addr}&cursor={cursor}
```

### 代币交易记录

```
GET /api/v1/chain/{chain}/token/transactions?address={addr}&contract={contract}&cursor={cursor}
```

### NFT 查询

```
GET /api/v1/chain/{chain}/nfts?address={addr}&cursor={cursor}
```

### 价格查询

```
# 单链价格
GET /api/v1/chain/{chain}/price

# 批量查询（不传 chains 则返回全部）
GET /api/v1/prices?chains=eth,bnb,tron,solana
```

## 价格数据源

价格服务采用多源 fallback + 缓存策略：

| 优先级 | 数据源 | 特点 |
|---|---|---|
| 1 | CoinGecko | 聚合多交易所，最接近钱包显示 |
| 2 | Huobi (HTX) | 国内可用 |
| 3 | Binance | 流动性最高 |

- 每 30 秒自动刷新缓存
- 接口直接读缓存返回，零延迟
- 任一数据源失败自动切换下一个

## 项目结构

```
web3-api/
├── cmd/server/main.go              # 入口
├── internal/
│   ├── config/config.go            # 配置加载
│   ├── model/model.go              # 领域模型 + URL 转换工具
│   ├── chain/
│   │   ├── provider.go             # ChainProvider 接口 + 注册表
│   │   ├── evm.go                  # ETH/BNB（Moralis）
│   │   ├── tron.go                 # Tron（TronGrid）
│   │   └── solana.go               # Solana（Helius）
│   ├── handler/
│   │   ├── handler.go              # 路由注册
│   │   ├── token.go                # 代币查询
│   │   ├── transaction.go          # 交易记录
│   │   ├── nft.go                  # NFT 查询
│   │   └── price.go                # 价格查询
│   ├── service/service.go          # 业务逻辑
│   └── price/
│       ├── price.go                # 价格服务 + 缓存
│       ├── binance.go              # Binance 数据源
│       ├── coingecko.go            # CoinGecko 数据源
│       └── huobi.go                # Huobi 数据源
├── docs/                           # Swagger 生成文件
├── config.example.yaml             # 配置模板
├── go.mod
└── go.sum
```

## 扩展新链

1. 实现 `ChainProvider` 接口
2. 在 `provider.go` 注册

```go
chain.RegisterChain("polygon", newEVMProvider("polygon"))
```

## 技术栈

- **HTTP 框架**：[gin](https://github.com/gin-gonic/gin)
- **配置管理**：[viper](https://github.com/spf13/viper)
- **日志**：[zap](https://go.uber.org/zap)
- **API 文档**：[swaggo/swag](https://github.com/swaggo/swag)

## License

MIT
