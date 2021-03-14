# <p align="center"> Blockchain Indexer 📇 in Golang
<p align="center">
  <img src="./index.png" />
</p>

## <p align="center"> ERC20 event Indexer on Matic

This project has been built 🔨 for getting hands dirty on Golang.
For an interesting idea, stay tuned for my next hackathon 👀 so please don't bully 🥺 this one with the expectations of likes of <h ref ="https://thegraph.com/"> The Graph </h>

## Setup 🏭

- Clone this repo
- Install the dependencies `go get -u ./...`
- Setup database URI in `connectionhelper/connector.go`
- Setup ERC20 contract to be indexed in `main.go`. By default it will Index <h ref= "https://explorer-mainnet.maticvigil.com/tokens/0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174/token-transfers">USDC ERC20</h> on Matic Mainnet
- For peeking database access results on `http://localhost:12345/peekIndexedDB` after which it would look like


![db result](./database.png)

## Tech Stack 🧱

- Golang
- Solidity
- Go Web3
- MongoDB