import {
	createContext,
	createSignal,
	onMount,
	useContext,
	type ParentProps,
} from "solid-js";
import { apiBaseUrl, clearToken, getToken, setToken } from "../api/auth";

interface AuthContextValue {
	token: () => string | null;
	isAuthenticated: () => boolean;
	login: () => void;
	logout: () => void;
}

const AuthContext = createContext<AuthContextValue>();

export function AuthProvider(props: ParentProps) {
	const [token, setTokenSignal] = createSignal<string | null>(getToken());

	// On mount, capture a token from the URL fragment if present (post-OAuth
	// callback), store it, then strip the fragment from the address bar so the
	// token doesn't linger in browser history.
	onMount(() => {
		const hash = window.location.hash;
		if (hash.startsWith("#token=")) {
			const t = hash.slice("#token=".length);
			if (t) {
				setToken(t);
				setTokenSignal(t);
			}
			window.history.replaceState(
				null,
				"",
				window.location.pathname + window.location.search,
			);
		}
	});

	const login = () => {
		// Browser navigates to the backend's login route, which 302s to Discord,
		// which 302s back to /api/v1/auth/callback, which 302s to the frontend
		// with #token=... in the URL fragment. The mount hook above captures it.
		//
		// We pass our own origin as ?origin=... because top-level navigations
		// don't reliably send the Origin header, and Referrer-Policy often
		// strips the Referer on cross-origin requests. The backend validates
		// this against ALLOWED_FRONTEND_ORIGINS.
		const origin = encodeURIComponent(window.location.origin);
		window.location.href = `${apiBaseUrl()}/api/v1/auth/login?origin=${origin}`;
	};

	const logout = () => {
		clearToken();
		setTokenSignal(null);
		// Hard reload so any auth-gated resources re-fetch in the unauthenticated
		// state. (Cheap; user expects a visible state change on logout.)
		window.location.reload();
	};

	return (
		<AuthContext.Provider
			value={{
				token,
				isAuthenticated: () => token() !== null,
				login,
				logout,
			}}
		>
			{props.children}
		</AuthContext.Provider>
	);
}

export function useAuth(): AuthContextValue {
	const ctx = useContext(AuthContext);
	if (!ctx) {
		throw new Error("useAuth must be used within AuthProvider");
	}
	return ctx;
}
