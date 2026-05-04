/**
 * Bitcoin — browser extension wallets (UniSat, Phantom BTC, OKX, OneKey, Bitget).
 */

export interface BitcoinProvider {
  requestAccounts?(): Promise<string[]>;
  getAccounts?(): Promise<string[]>;
  connect?(): Promise<string[]>;
  disconnect?(): Promise<void>;
  signMessage(message: string, type?: "ecdsa" | "bip322-simple"): Promise<string>;
}

function getBitcoinProvider(): BitcoinProvider | null {
  if (typeof window === "undefined") return null;
  const w = window as unknown as {
    $onekey?: { btc?: BitcoinProvider };
    okxwallet?: { bitcoin?: BitcoinProvider };
    bitkeep?: { unisat?: BitcoinProvider; isBitKeep?: boolean };
    unisat?: BitcoinProvider;
    phantom?: { bitcoin?: BitcoinProvider };
  };
  if (w.$onekey?.btc) return w.$onekey.btc;
  if (w.$onekey && w.unisat) return w.unisat;
  if (w.okxwallet?.bitcoin) return w.okxwallet.bitcoin;
  if (w.bitkeep?.unisat) return w.bitkeep.unisat;
  if (w.unisat && !w.$onekey && !w.bitkeep?.isBitKeep) return w.unisat;
  if (w.phantom?.bitcoin) return w.phantom.bitcoin;
  return null;
}

export async function getBitcoinAddress(): Promise<string> {
  const provider = getBitcoinProvider();
  if (!provider) throw new Error("No Bitcoin wallet detected");
  let accounts: string[] = [];
  if (provider.requestAccounts) {
    accounts = await provider.requestAccounts();
  } else if (provider.getAccounts) {
    accounts = await provider.getAccounts();
  } else if (provider.connect) {
    accounts = await provider.connect();
  }
  if (!accounts?.length) throw new Error("No Bitcoin address found. Please connect your wallet.");
  return accounts[0];
}

export async function disconnectBitcoin(): Promise<void> {
  const provider = getBitcoinProvider();
  if (!provider?.disconnect) return;
  try {
    await provider.disconnect();
  } catch {
    /* ignore */
  }
}
