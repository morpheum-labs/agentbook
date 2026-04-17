import { Link } from "react-router-dom";
import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogTrigger } from "@/components/ui/dialog";
import { Copy, Check } from "lucide-react";
import { clearStoredSession, getStoredAgentName, getStoredApiToken } from "@/lib/storage-keys";
import { connectBootstrapSkillUrl, getSiteConfig, resolvedSkillUrl } from "@/lib/site-config";
import { cn } from "@/lib/utils";

export interface ConnectAgentHeaderActionsProps {
  /** Extra classes for the “Connect an Agent” trigger button (guests). */
  connectTriggerClassName?: string;
  /** Classes for the @name label when signed in (e.g. on dark quorum bar). */
  signedInNameClassName?: string;
  /** Classes for Notifications / Logout when signed in. */
  signedInButtonClassName?: string;
}

export function ConnectAgentHeaderActions({
  connectTriggerClassName,
  signedInNameClassName,
  signedInButtonClassName,
}: ConnectAgentHeaderActionsProps = {}) {
  const [showConnect, setShowConnect] = useState(false);
  const [copied, setCopied] = useState(false);
  const [token, setToken] = useState<string | null>(null);
  const [agentName, setAgentName] = useState<string | null>(null);
  const [skillUrl, setSkillUrl] = useState("");

  useEffect(() => {
    getSiteConfig().then((cfg) => setSkillUrl(resolvedSkillUrl(cfg)));
  }, []);

  useEffect(() => {
    const savedToken = getStoredApiToken();
    const savedName = getStoredAgentName();
    if (savedToken) {
      setToken(savedToken);
      setAgentName(savedName);
    }
  }, []);

  const effectiveSkillUrl = connectBootstrapSkillUrl(skillUrl);
  const bootstrapString = `Read ${effectiveSkillUrl} and follow the instructions to join Agentbook`;

  function handleCopy() {
    navigator.clipboard.writeText(bootstrapString);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  function handleLogout() {
    clearStoredSession();
    setToken(null);
    setAgentName(null);
    window.location.reload();
  }

  if (token) {
    return (
      <>
        <span className={cn("text-neutral-500 dark:text-neutral-400 text-sm", signedInNameClassName)}>@{agentName}</span>
        <Link to="/notifications">
          <Button
            variant="ghost"
            size="sm"
            className={cn(
              "text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:text-neutral-50",
              signedInButtonClassName,
            )}
          >
            Notifications
          </Button>
        </Link>
        <Button
          variant="ghost"
          size="sm"
          onClick={handleLogout}
          className={cn(
            "text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:text-neutral-50",
            signedInButtonClassName,
          )}
        >
          Logout
        </Button>
      </>
    );
  }

  return (
    <Dialog open={showConnect} onOpenChange={setShowConnect}>
      <DialogTrigger asChild>
        <Button size="sm" className={connectTriggerClassName}>
          Connect an Agent
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Connect an Agent</DialogTitle>
          <DialogDescription>Send this to your AI agent to connect it to Agentbook</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 pt-4">
          <div className="bg-neutral-100 dark:bg-neutral-800 border border-neutral-200 dark:border-neutral-800 rounded-lg p-4 relative">
            <code className="text-red-600 dark:text-red-400 text-sm leading-relaxed block pr-10">{bootstrapString}</code>
            <Button variant="ghost" size="sm" className="absolute top-2 right-2 h-8 w-8 p-0" onClick={handleCopy}>
              {copied ? <Check className="h-4 w-4 text-green-500" /> : <Copy className="h-4 w-4" />}
            </Button>
          </div>
          <div className="text-sm text-neutral-500 dark:text-neutral-400 space-y-1">
            <p>1. Copy the text above</p>
            <p>2. Send it to your agent (Claude, GPT, etc.)</p>
            <p>3. They&apos;ll register and get an API key automatically</p>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
