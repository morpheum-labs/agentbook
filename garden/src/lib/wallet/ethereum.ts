/**
 * Ethereum — eth_requestAccounts / disconnect via wallet_revokePermissions when supported.
 */

export interface EthereumProvider {
  request(args: { method: "eth_requestAccounts" }): Promise<string[]>;
  request(args: { method: "eth_sign"; params: [string, string] }): Promise<string>;
  request(args: { method: "eth_call"; params: [{ to: string; data: string }, string] }): Promise<string>;
  request(args: { method: "eth_chainId" }): Promise<string>;
}

function getEthereumProvider(): EthereumProvider | null {
  if (typeof window === "undefined") return null;
  const w = window as unknown as {
    $onekey?: { ethereum?: EthereumProvider };
    okxwallet?: { ethereum?: EthereumProvider };
    bitkeep?: { ethereum?: EthereumProvider };
    ethereum?: EthereumProvider;
    phantom?: { ethereum?: EthereumProvider };
  };
  if (w.$onekey?.ethereum) return w.$onekey.ethereum;
  if (w.okxwallet?.ethereum) return w.okxwallet.ethereum;
  if (w.bitkeep?.ethereum) return w.bitkeep.ethereum;
  if (w.ethereum) return w.ethereum;
  if (w.phantom?.ethereum) return w.phantom.ethereum;
  return null;
}

export async function getEthereumAddress(): Promise<string> {
  const provider = getEthereumProvider();
  if (!provider) throw new Error("No Ethereum wallet detected");
  const accounts = await provider.request({ method: "eth_requestAccounts" });
  if (!accounts?.length) throw new Error("No accounts found. Please connect your wallet.");
  return accounts[0];
}

export async function disconnectEthereum(): Promise<void> {
  const provider = getEthereumProvider();
  if (!provider) return;
  try {
    await (provider as { request: (args: unknown) => Promise<unknown> }).request({
      method: "wallet_revokePermissions",
      params: [{ eth_accounts: {} }],
    });
  } catch {
    /* not all wallets support wallet_revokePermissions */
  }
}
