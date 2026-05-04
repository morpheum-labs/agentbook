export type { ChainType, WalletName } from "./extdetection";
export { getBestWallet } from "./extdetection";
export { disconnectBitcoin } from "./bitcoin";
export { disconnectEthereum } from "./ethereum";
export { disconnectSolana } from "./solana";
export {
  WALLET_CHAIN_OPTIONS,
  connectWalletForChain,
  type WalletConnectedSession,
} from "./connect";
