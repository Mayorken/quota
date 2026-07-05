import { useEffect, useRef, useState } from "react";
import { api } from "../api";

// Minimal typing for the Google Identity Services global.
declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (opts: {
            client_id: string;
            callback: (resp: { credential: string }) => void;
          }) => void;
          renderButton: (el: HTMLElement, opts: Record<string, unknown>) => void;
        };
      };
    };
  }
}

interface AuthConfig {
  google_client_id: string;
  google_sign_in: boolean;
}

/**
 * Renders the official "Sign in with Google" button. It self-hides unless the
 * backend reports a configured client ID, so email/password stays the default
 * when Google isn't set up.
 */
export default function GoogleButton({
  onCredential,
}: {
  onCredential: (credential: string) => void;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [clientId, setClientId] = useState<string>("");
  const [ready, setReady] = useState(false);
  const cbRef = useRef(onCredential);
  cbRef.current = onCredential;

  // Discover whether Google sign-in is enabled.
  useEffect(() => {
    api
      .get<AuthConfig>("/auth/config")
      .then((cfg) => {
        if (cfg.google_sign_in) setClientId(cfg.google_client_id);
      })
      .catch(() => {});
  }, []);

  // Load the Google Identity Services script on demand — only once we know
  // Google sign-in is actually enabled — so it never touches the network in
  // the default email/password flow.
  useEffect(() => {
    if (!clientId) return;
    if (window.google?.accounts?.id) {
      setReady(true);
      return;
    }
    const SRC = "https://accounts.google.com/gsi/client";
    let script = document.querySelector<HTMLScriptElement>(`script[src="${SRC}"]`);
    if (!script) {
      script = document.createElement("script");
      script.src = SRC;
      script.async = true;
      script.defer = true;
      document.head.appendChild(script);
    }
    const id = setInterval(() => {
      if (window.google?.accounts?.id) {
        setReady(true);
        clearInterval(id);
      }
    }, 150);
    return () => clearInterval(id);
  }, [clientId]);

  // Initialize and render the button once both the client ID and script exist.
  useEffect(() => {
    if (!ready || !clientId || !ref.current) return;
    window.google!.accounts.id.initialize({
      client_id: clientId,
      callback: (resp) => cbRef.current(resp.credential),
    });
    ref.current.innerHTML = "";
    window.google!.accounts.id.renderButton(ref.current, {
      theme: "outline",
      size: "large",
      width: 316,
      text: "continue_with",
      shape: "rectangular",
      logo_alignment: "center",
    });
  }, [ready, clientId]);

  if (!clientId) return null;

  return (
    <>
      <div className="or-divider">
        <span>or</span>
      </div>
      <div className="google-btn" ref={ref} />
    </>
  );
}
