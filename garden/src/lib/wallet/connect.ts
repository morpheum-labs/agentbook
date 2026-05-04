import type { ChainType, WalletName } from "./extdetection";
import { getBestWallet } from "./extdetection";
import { getBitcoinAddress } from "./bitcoin";
import { getEthereumAddress } from "./ethereum";
import { getSolanaAddress } from "./solana";

export type WalletConnectedSession = {
  chain: ChainType;
  address: string;
  walletName: WalletName;
};

/** Prompt the user to connect the best detected wallet for the given chain. */
export async function connectWalletForChain(chain: ChainType): Promise<WalletConnectedSession> {
  const walletName = getBestWallet(chain);
  if (!walletName) {
    throw new Error("No wallet detected for this chain. Install a supported browser wallet.");
  }
  let address: string;
  if (chain === "ethereum") {
    address = await getEthereumAddress();
  } else if (chain === "solana") {
    address = await getSolanaAddress();
  } else {
    address = await getBitcoinAddress();
  }
  return { chain, address, walletName };
}

export const WALLET_CHAIN_OPTIONS: { chain: ChainType; label: string; icon: string }[] = [
  { chain: "ethereum", label: "Ethereum", icon: "⟠" },
  { chain: "solana", label: "Solana", icon: "◎" },
  { chain: "bitcoin", label: "Bitcoin", icon: "₿" },
];
