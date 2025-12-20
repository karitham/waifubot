export const formatDate = (date: string) =>
	new Date(date).toLocaleDateString(navigator.language);

export const mapCharType = (type: string) =>
	type === "OLD" ? "unknown" : type?.toLowerCase();

export const getSearchParams = (sp: Record<string, unknown>) => {
	const params = new URLSearchParams();
	Object.entries(sp).forEach(([k, v]) => {
		if (v !== undefined) params.set(k, String(v));
	});
	return params.toString();
};
