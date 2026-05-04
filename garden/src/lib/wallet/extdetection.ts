/**
 * Extension detection — wallet detection for Ethereum, Solana, and Bitcoin.
 * Ported from storefront; detects MetaMask, Phantom, OneKey, OKX, Bitget, TokenPocket, Coinbase, Rainbow, Binance, Unisat.
 */

export type ChainType = "ethereum" | "solana" | "bitcoin";

export type WalletName =
  | "OneKey"
  | "OKX"
  | "Bitget"
  | "TokenPocket"
  | "Coinbase Wallet"
  | "Rainbow"
  | "Binance"
  | "MetaMask"
  | "Phantom"
  | "Bitcoin Wallet"
  | "Unisat"
  | null;

function isOneKeyAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { $onekey?: { ethereum?: unknown }; ethereum?: { isOneKey?: boolean } };
  return !!(w.$onekey?.ethereum || w.ethereum?.isOneKey || w.$onekey);
}

function isPhantomAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as {
    phantom?: { ethereum?: unknown; solana?: unknown; bitcoin?: unknown };
    ethereum?: { isPhantom?: boolean; isOneKey?: boolean };
    $onekey?: unknown;
  };
  const hasPhantom = !!(w.phantom?.ethereum || w.phantom?.solana || w.phantom?.bitcoin || w.ethereum?.isPhantom);
  return hasPhantom && !w.$onekey && !w.ethereum?.isOneKey;
}

function isOKXEthereumAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { okxwallet?: { ethereum?: unknown }; ethereum?: { isOKX?: boolean } };
  return !!(w.okxwallet?.ethereum || w.ethereum?.isOKX);
}

function isOKXSolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  return !!(window as unknown as { okxwallet?: { solana?: unknown } }).okxwallet?.solana;
}

function isOKXBitcoinAvailable(): boolean {
  if (typeof window === "undefined") return false;
  return !!(window as unknown as { okxwallet?: { bitcoin?: unknown } }).okxwallet?.bitcoin;
}

function isBitgetProviderAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const bitkeep = (window as unknown as { bitkeep?: { isBitKeep?: boolean; ethereum?: { isBitEthereum?: boolean } } }).bitkeep;
  if (!bitkeep) return false;
  return bitkeep.isBitKeep === true || bitkeep.ethereum?.isBitEthereum === true;
}

function isBitgetSolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const bitkeep = (window as unknown as { bitkeep?: { solana?: unknown; isBitKeep?: boolean } }).bitkeep;
  return !!(bitkeep?.solana && (bitkeep.isBitKeep === true || (bitkeep.solana as { isBitKeep?: boolean })?.isBitKeep === true));
}

function isTokenPocketProviderAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { tokenpocket?: unknown; ethereum?: { isTokenPocket?: boolean }; tp?: unknown };
  return !!(
    typeof w.tokenpocket !== "undefined" ||
    w.ethereum?.isTokenPocket === true ||
    (w.tp && typeof w.tp === "object")
  );
}

function isTokenPocketSolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const solana = (window as unknown as { solana?: { name?: string; isPhantom?: boolean } }).solana;
  if (!solana) return false;
  if (solana.name === "TokenPocket") return true;
  if (solana.isPhantom === true) return false;
  return !!(window as unknown as { tokenpocket?: unknown }).tokenpocket;
}

function isRainbowAvailable(): boolean {
  if (typeof window === "undefined") return false;
  return (window as unknown as { ethereum?: { isRainbow?: boolean } }).ethereum?.isRainbow === true;
}

function isBinanceProviderAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { binancew3w?: { isExtension?: boolean }; BinanceChain?: unknown; ethereum?: { isBinance?: boolean } };
  return !!(w.binancew3w?.isExtension === true || w.BinanceChain || w.ethereum?.isBinance);
}

function isBinanceSolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { binancew3w?: { isExtension?: boolean; solana?: unknown }; solana?: { isBinance?: boolean } };
  if (w.binancew3w?.isExtension === true && w.binancew3w?.solana) return true;
  return w.solana?.isBinance === true;
}

function isMetaMaskProviderAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const ethereum = (window as unknown as { ethereum?: { isMetaMask?: boolean; _metamask?: unknown; providerInfo?: { rdns?: string }; isOneKey?: boolean; isOKX?: boolean } }).ethereum;
  if (!ethereum) return false;
  if (ethereum.isMetaMask === true) {
    if (ethereum._metamask && typeof ethereum._metamask === "object") return true;
    if (ethereum.providerInfo?.rdns === "io.metamask") return true;
    const hasOKX = !!(window as unknown as { okxwallet?: unknown }).okxwallet;
    const hasBinance = isBinanceProviderAvailable();
    const hasOneKey = !!(window as unknown as { $onekey?: unknown }).$onekey;
    const hasPhantom = !!(window as unknown as { phantom?: { ethereum?: unknown } }).phantom?.ethereum;
    if (!hasOKX && !hasBinance && !hasOneKey && !hasPhantom) return true;
  }
  return false;
}

function isCoinbaseWalletProviderAvailableSync(): boolean {
  if (typeof window === "undefined") return false;
  const ethereum = (window as unknown as { ethereum?: { isCoinbaseWallet?: boolean; host?: string; jsonRpcUrl?: string; providers?: unknown[] } }).ethereum;
  if (!ethereum) return false;
  let provider = ethereum;
  if (Array.isArray(ethereum.providers)) {
    provider = ethereum.providers.find((p: { isCoinbaseWallet?: boolean }) => p.isCoinbaseWallet) ?? ethereum;
  }
  if (!provider.isCoinbaseWallet) return false;
  if (typeof provider.host !== "string" || !provider.host.includes("coinbase.com")) return false;
  return true;
}

function isCoinbaseWalletSolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const coinbaseSolana = (window as unknown as { coinbaseSolana?: { _storage?: { scope?: string }; connect?: unknown; signMessage?: unknown } }).coinbaseSolana;
  if (!coinbaseSolana) return false;
  if (coinbaseSolana._storage?.scope === "coinbaseSolana") return true;
  return !!(typeof coinbaseSolana.connect === "function" && typeof coinbaseSolana.signMessage === "function");
}

function isOneKeySolanaAvailable(): boolean {
  if (typeof window === "undefined") return false;
  return !!(window as unknown as { $onekey?: { solana?: unknown } }).$onekey?.solana;
}

function isOneKeyBitcoinAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { $onekey?: { btc?: unknown }; unisat?: unknown };
  return !!(w.$onekey?.btc || (w.$onekey && w.unisat));
}

function isBitgetBitcoinAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { bitkeep?: { isBitKeep?: boolean; ethereum?: { isBitEthereum?: boolean }; unisat?: unknown }; unisat?: unknown; $onekey?: unknown };
  const hasBitget = !!(w.bitkeep?.isBitKeep || w.bitkeep?.ethereum?.isBitEthereum);
  return hasBitget && !!(w.bitkeep?.unisat || (w.unisat && !w.$onekey));
}

function isUnisatBitcoinAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const w = window as unknown as { unisat?: { requestAccounts?: unknown; signMessage?: unknown }; $onekey?: unknown; bitkeep?: { isBitKeep?: boolean } };
  if (!w.unisat) return false;
  if (w.$onekey) return false;
  if (w.bitkeep?.isBitKeep) return false;
  return !!(typeof w.unisat.requestAccounts === "function" && typeof w.unisat.signMessage === "function");
}

function isPhantomBitcoinAvailable(): boolean {
  if (typeof window === "undefined") return false;
  const phantom = (window as unknown as { phantom?: { bitcoin?: { requestAccounts?: unknown; signMessage?: unknown } } }).phantom;
  return !!(phantom?.bitcoin && typeof phantom.bitcoin.requestAccounts === "function" && typeof phantom.bitcoin.signMessage === "function");
}

function isBitcoinWalletAvailable(): boolean {
  return isOneKeyBitcoinAvailable() || isOKXBitcoinAvailable() || isBitgetBitcoinAvailable() || isUnisatBitcoinAvailable() || isPhantomBitcoinAvailable();
}

/** Best available wallet name for the chain (check order matches storefront). */
export function getBestWallet(chainType: ChainType): WalletName {
  if (chainType === "ethereum") {
    if (isOneKeyAvailable()) return "OneKey";
    if (isOKXSolanaAvailable() && isOKXBitcoinAvailable()) return "OKX";
    if (isOKXEthereumAvailable()) return "OKX";
    if (isBitgetProviderAvailable()) return "Bitget";
    if (isTokenPocketProviderAvailable()) return "TokenPocket";
    if (isCoinbaseWalletProviderAvailableSync()) return "Coinbase Wallet";
    if (isRainbowAvailable()) return "Rainbow";
    if (isBinanceProviderAvailable()) return "Binance";
    if (isMetaMaskProviderAvailable()) return "MetaMask";
    if (isPhantomAvailable()) return "Phantom";
  } else if (chainType === "solana") {
    if (isPhantomAvailable()) return "Phantom";
    if (isOneKeySolanaAvailable()) return "OneKey";
    if (isOKXSolanaAvailable()) return "OKX";
    if (isBitgetSolanaAvailable()) return "Bitget";
    if (isTokenPocketSolanaAvailable()) return "TokenPocket";
    if (isCoinbaseWalletSolanaAvailable()) return "Coinbase Wallet";
    if (isBinanceSolanaAvailable()) return "Binance";
  } else if (chainType === "bitcoin") {
    if (isOneKeyBitcoinAvailable()) return "OneKey";
    if (isOKXBitcoinAvailable()) return "OKX";
    if (isBitgetBitcoinAvailable()) return "Bitget";
    if (isUnisatBitcoinAvailable()) return "Unisat";
    if (isPhantomBitcoinAvailable()) return "Phantom";
    if (isBitcoinWalletAvailable()) return "Bitcoin Wallet";
  }
  return null;
}
