/**
 * Solana — connect via standard browser wallet provider.
 */

export interface SolanaProvider {
  isConnected: boolean;
  publicKey: { toString(): string } | null;
  connect?(opts?: { onlyIfTrusted?: boolean }): Promise<{ publicKey: { toString(): string } }>;
  disconnect?(): Promise<void>;
  signMessage(message: Uint8Array | string, display?: "utf8" | "hex"): Promise<{
    signature: Uint8Array | number[];
    publicKey?: { toString(): string };
  }>;
}

function getSolanaProvider(): SolanaProvider | null {
  if (typeof window === "undefined") return null;
  const w = window as unknown as {
    $onekey?: { solana?: SolanaProvider };
    okxwallet?: { solana?: SolanaProvider };
    bitkeep?: { solana?: SolanaProvider };
    solana?: SolanaProvider;
    coinbaseSolana?: SolanaProvider;
    binancew3w?: { solana?: SolanaProvider };
    phantom?: { solana?: SolanaProvider };
  };
  if (w.$onekey?.solana) return w.$onekey.solana;
  if (w.okxwallet?.solana) return w.okxwallet.solana;
  if (w.bitkeep?.solana) return w.bitkeep.solana;
  if (w.coinbaseSolana) return w.coinbaseSolana;
  if (w.binancew3w?.solana) return w.binancew3w.solana;
  if (w.solana) return w.solana;
  if (w.phantom?.solana) return w.phantom.solana;
  return null;
}

export async function getSolanaAddress(): Promise<string> {
  const provider = getSolanaProvider();
  if (!provider) throw new Error("No Solana wallet detected");
  if (!provider.isConnected && provider.connect) {
    const res = await provider.connect();
    return res.publicKey.toString();
  }
  if (provider.publicKey) return provider.publicKey.toString();
  throw new Error("Solana wallet not connected");
}

export async function disconnectSolana(): Promise<void> {
  const provider = getSolanaProvider();
  if (!provider?.disconnect) return;
  try {
    await provider.disconnect();
  } catch {
    /* ignore */
  }
}
