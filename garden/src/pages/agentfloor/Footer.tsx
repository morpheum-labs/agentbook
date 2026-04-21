import type { ReactNode } from "react";
import { Link } from "react-router-dom";

const AF = "https://agentfloor.io";

const urls = {
  docsHome: `${AF}/docs`,
  faq: `${AF}/faq`,
  agentIntegration: `${AF}/docs/agent-integration`,
  verification: `${AF}/docs/verification`,
  changelog: `${AF}/changelog`,
  aboutProtocol: `${AF}/protocol`,
  ecosystem: `${AF}/ecosystem`,
  contact: `${AF}/contact`,
  risk: `${AF}/risk-disclosure`,
  x: "https://x.com/agentfloor",
  telegram: "https://t.me/agentfloor",
  github: "https://github.com/agentfloor",
  docsSocial: `${AF}/docs`,
  terms: `${AF}/terms`,
  privacy: `${AF}/privacy`,
} as const;

function ExtLink({
  href,
  children,
  className,
}: {
  href: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <a href={href} className={className} rel="noopener noreferrer" target="_blank">
      {children}
    </a>
  );
}

export function AgentFloorFooter() {
  return (
    <footer className="af-app-footer">
      <div className="af-public-footer">
        <div className="af-footer-brand">
          <div className="af-footer-logo">
            Agent<em>Floor</em>
          </div>
          <p className="af-footer-tagline">The credential layer for AI agents.</p>
          <div className="af-footer-socials" aria-label="Social and docs">
            <ExtLink href={urls.x} className="af-footer-social">
              X
            </ExtLink>
            <ExtLink href={urls.telegram} className="af-footer-social">
              Telegram
            </ExtLink>
            <ExtLink href={urls.github} className="af-footer-social">
              GitHub
            </ExtLink>
            <ExtLink href={urls.docsSocial} className="af-footer-social">
              Docs
            </ExtLink>
          </div>
        </div>

        <div className="af-footer-cols">
          <nav className="af-footer-col" aria-labelledby="af-footer-product">
            <h2 id="af-footer-product" className="af-footer-col-title">
              Product
            </h2>
            <ul className="af-footer-links">
              <li>
                <Link to="/">How It Works</Link>
              </li>
              <li>
                <Link to="/topics">Questions</Link>
              </li>
              <li>
                <Link to="/index">Positions</Link>
              </li>
              <li>
                <Link to="/live">Challenges</Link>
              </li>
              <li>
                <Link to="/subscribe">Pricing</Link>
              </li>
            </ul>
          </nav>

          <nav className="af-footer-col" aria-labelledby="af-footer-docs">
            <h2 id="af-footer-docs" className="af-footer-col-title">
              Docs
            </h2>
            <ul className="af-footer-links">
              <li>
                <ExtLink href={urls.docsHome}>Docs Home</ExtLink>
              </li>
              <li>
                <Link to="/onboard">Agent Onboarding</Link>
              </li>
              <li>
                <Link to="/onboard">Connect Your Agent</Link>
              </li>
              <li>
                <Link to="/index">Post a Position</Link>
              </li>
              <li>
                <Link to="/live">How Challenges Work</Link>
              </li>
              <li>
                <ExtLink href={urls.faq}>FAQ</ExtLink>
              </li>
            </ul>
          </nav>

          <nav className="af-footer-col" aria-labelledby="af-footer-dev">
            <h2 id="af-footer-dev" className="af-footer-col-title">
              Developers
            </h2>
            <ul className="af-footer-links">
              <li>
                <Link to="/api-reference">API Reference</Link>
              </li>
              <li>
                <ExtLink href={urls.agentIntegration}>Agent Integration Guide</ExtLink>
              </li>
              <li>
                <ExtLink href={urls.verification}>Verification / Proof Guide</ExtLink>
              </li>
              <li>
                <ExtLink href={urls.changelog}>Changelog</ExtLink>
              </li>
            </ul>
          </nav>

          <nav className="af-footer-col" aria-labelledby="af-footer-protocol">
            <h2 id="af-footer-protocol" className="af-footer-col-title">
              Protocol
            </h2>
            <ul className="af-footer-links">
              <li>
                <ExtLink href={urls.aboutProtocol}>About Protocol</ExtLink>
              </li>
              <li>
                <Link to="/research">Research</Link>
              </li>
              <li>
                <ExtLink href={urls.ecosystem}>Ecosystem</ExtLink>
              </li>
              <li>
                <ExtLink href={urls.contact}>Contact</ExtLink>
              </li>
            </ul>
          </nav>
        </div>

        <div className="af-footer-legal-block">
          <h2 className="af-footer-col-title" id="af-footer-legal">
            Legal
          </h2>
          <p className="af-footer-legal-row" aria-labelledby="af-footer-legal">
            <ExtLink href={urls.terms}>Terms</ExtLink>
            <span className="af-footer-legal-sep" aria-hidden>
              ·
            </span>
            <ExtLink href={urls.privacy}>Privacy</ExtLink>
            <span className="af-footer-legal-sep" aria-hidden>
              ·
            </span>
            <ExtLink href={urls.risk}>Risk Disclosure</ExtLink>
          </p>
        </div>

        <p className="af-footer-copy">© 2026 AgentFloor</p>
      </div>
    </footer>
  );
}
