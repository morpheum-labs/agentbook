/** Session keys shared by Multi-agent chat and runtime pairing flow. */
export const MULTI_CHAT_SESSION = {
  GATEWAY_BASE: "agentswarm:multi_chat:gateway_ws_base",
  GATEWAY_TOKEN: "agentswarm:multi_chat:gateway_ws_token",
  GATEWAY_SOURCE: "agentswarm:multi_chat:gateway_source",
  GATEWAY_INSTANCE: "agentswarm:multi_chat:gateway_instance_name",
  FRESH_EACH: "agentswarm:multi_chat:fresh_each_send",
} as const;

/** After pairing, open Multi-agent chat with runtime mode + stored bearer token. */
export function persistMultiChatGatewayRuntime(opts: {
  instanceName: string;
  gatewayPairingToken: string;
}): void {
  try {
    sessionStorage.setItem(MULTI_CHAT_SESSION.GATEWAY_SOURCE, "runtime");
    sessionStorage.setItem(MULTI_CHAT_SESSION.GATEWAY_INSTANCE, opts.instanceName);
    sessionStorage.setItem(MULTI_CHAT_SESSION.GATEWAY_TOKEN, opts.gatewayPairingToken);
    sessionStorage.setItem(MULTI_CHAT_SESSION.GATEWAY_BASE, "");
  } catch {
    // ignore
  }
}
