import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from "react";

type ToastFn = (message: string) => void;

const ToastContext = createContext<ToastFn | null>(null);

export function AgentFloorToastProvider({ children }: { children: ReactNode }) {
  const [message, setMessage] = useState("");
  const [visible, setVisible] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const toast = useCallback((msg: string) => {
    setMessage(msg);
    setVisible(true);
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      setVisible(false);
    }, 3000);
  }, []);

  const value = useMemo(() => toast, [toast]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div
        className="toast"
        style={{ display: visible ? "block" : "none" }}
      >
        {message}
      </div>
    </ToastContext.Provider>
  );
}

/** Toast API for AgentFloor HTML screens (event delegation). */
// eslint-disable-next-line react-refresh/only-export-components -- hook paired with Provider in this module
export function useAgentFloorToast(): ToastFn {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error("useAgentFloorToast must be used under AgentFloorToastProvider");
  }
  return ctx;
}
