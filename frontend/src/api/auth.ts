const TOKEN_KEY = "auth_token";

export function getToken(): string | null {
	return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null): void {
	if (token) localStorage.setItem(TOKEN_KEY, token);
	else localStorage.removeItem(TOKEN_KEY);
}

export function clearToken(): void {
	localStorage.removeItem(TOKEN_KEY);
}

export function apiBaseUrl(): string {
	return import.meta.env.VITE_API_URL || "https://waifuapi.karitham.dev";
}

// Inject the bearer token into every oazapfts request. Module-load side effect
// so callers don't need to remember.
import { defaults } from "./generated";
const baseFetch = defaults.fetch ?? fetch;
defaults.fetch = async (input, init) => {
	const token = getToken();
	if (token) {
		init = {
			...init,
			headers: {
				...((init?.headers as Record<string, string>) || {}),
				Authorization: `Bearer ${token}`,
			},
		};
	}
	return baseFetch(input as RequestInfo, init);
};
