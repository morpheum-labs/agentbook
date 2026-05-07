import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";
import {
  applyTerminalFxDocumentClass,
  cycleTerminalFxMode,
  getStoredTerminalFxMode,
  setStoredTerminalFxMode,
  type TerminalFxMode,
} from "@/lib/terminal-fx-storage";

function useHtmlHasClass(className: string): boolean {
  const [on, setOn] = useState(
    () => typeof document !== "undefined" && document.documentElement.classList.contains(className)
  );

  useEffect(() => {
    const el = document.documentElement;
    const sync = () => setOn(el.classList.contains(className));
    sync();
    const obs = new MutationObserver(sync);
    obs.observe(el, { attributes: true, attributeFilter: ["class"] });
    return () => obs.disconnect();
  }, [className]);

  return on;
}

function TerminalFxSnow({ enabled }: { enabled: boolean }) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = ref.current;
    if (!host || !enabled) {
      if (host) host.innerHTML = "";
      return;
    }

    const snowflakeCount = 90;
    const snowflakeSize = 6;
    const snowflakeDuration = 15;

    for (let i = 0; i < snowflakeCount; i++) {
      const snowflake = document.createElement("div");
      snowflake.classList.add("terminal-fx-snowflake");

      const size = Math.random() * snowflakeSize + 0.25;
      snowflake.style.width = `${size}px`;
      snowflake.style.height = `${size}px`;
      snowflake.style.left = `${Math.random() * 100}%`;
      snowflake.style.top = `${Math.random() * -20}vh`;
      snowflake.style.opacity = `${Math.random() * 0.6 + 0.1}`;

      let anim =
        `terminal-snowflake-fall ${snowflakeDuration + Math.random() * 10}s linear infinite`;
      if (Math.random() > 0.3) {
        const horizontalAnimation =
          Math.random() > 0.5 ? "terminal-snowflake-fall-horizontal-1" : "terminal-snowflake-fall-horizontal-2";
        anim = `${horizontalAnimation} ${snowflakeDuration + Math.random() * 10}s linear infinite`;
      }
      snowflake.style.animation = anim;
      snowflake.style.animationDelay = `-${(snowflakeDuration + Math.random() * 50) / 4}s`;

      host.appendChild(snowflake);
    }

    return () => {
      host.innerHTML = "";
    };
  }, [enabled]);

  return (
    <div
      ref={ref}
      className="pointer-events-none absolute inset-0 z-[1] overflow-hidden"
      aria-hidden
    />
  );
}

/** Full-viewport CRT / scanlines when terminal FX are active (same stacking as original app-wide layers). */
function TerminalFxGlobalCrt({ mode, isDark }: { mode: TerminalFxMode; isDark: boolean }) {
  const active = isDark && mode !== "off";
  if (!active) return null;
  return <div className="terminal-fx-crt" aria-hidden />;
}

/** Rain / snow / noise / grunge scoped to a `.terminal-fx-hero` ancestor — does not alter the main page background. */
export function TerminalFxHeroLayers({ mode, isDark }: { mode: TerminalFxMode; isDark: boolean }) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const active = isDark && mode !== "off";
  const winter = mode === "winter";
  const showRain = active && !winter;

  useEffect(() => {
    const v = videoRef.current;
    if (!v) return;
    if (!showRain) {
      v.pause();
      return;
    }
    v.playbackRate = 0.8;
    v.play().catch(() => {});
  }, [showRain]);

  if (!active) return null;

  return (
    <>
      <video
        ref={videoRef}
        className="terminal-fx-rain"
        src="/terminal-effects/videos/rain-bg.mp4"
        poster="/terminal-effects/images/rain-bg-static.png"
        preload="metadata"
        muted
        playsInline
        loop
        aria-hidden
      />
      <TerminalFxSnow enabled={winter} />
      <div className="terminal-fx-noise" aria-hidden />
      <div className="terminal-fx-grunge" aria-hidden />
    </>
  );
}

type TerminalFxContextValue = {
  mode: TerminalFxMode;
  setMode: (m: TerminalFxMode) => void;
  cycleMode: () => void;
};

const TerminalFxContext = createContext<TerminalFxContextValue | null>(null);

export function TerminalFxProvider({ children }: { children: ReactNode }) {
  const [mode, setModeState] = useState<TerminalFxMode>(() => getStoredTerminalFxMode());
  const isDark = useHtmlHasClass("dark");

  const setMode = useCallback((m: TerminalFxMode) => {
    setStoredTerminalFxMode(m);
    setModeState(m);
  }, []);

  const cycleMode = useCallback(() => {
    setModeState((prev) => {
      const next = cycleTerminalFxMode(prev);
      setStoredTerminalFxMode(next);
      return next;
    });
  }, []);

  useEffect(() => {
    applyTerminalFxDocumentClass(mode, isDark);
  }, [mode, isDark]);

  const value = useMemo(
    () => ({
      mode,
      setMode,
      cycleMode,
    }),
    [mode, setMode, cycleMode]
  );

  return (
    <TerminalFxContext.Provider value={value}>
      <TerminalFxGlobalCrt mode={mode} isDark={isDark} />
      {children}
    </TerminalFxContext.Provider>
  );
}

/** Use inside a container with class `terminal-fx-hero` (see CronJobsHero / RuntimeInstancesHero). */
export function TerminalFxHeroDecor() {
  const { mode } = useTerminalFx();
  const isDark = useHtmlHasClass("dark");
  return <TerminalFxHeroLayers mode={mode} isDark={isDark} />;
}

export function useTerminalFx(): TerminalFxContextValue {
  const ctx = useContext(TerminalFxContext);
  if (!ctx) throw new Error("useTerminalFx must be used within TerminalFxProvider");
  return ctx;
}
