interface RGB {
	r: number;
	g: number;
	b: number;
}

interface Threshold {
	val: number;
	color: RGB;
}

const thresholdColors: Threshold[] = [
	{ val: 1, color: { r: 150, g: 150, b: 150 } },
	{ val: 100, color: { r: 46, g: 204, b: 113 } },
	{ val: 1000, color: { r: 52, g: 152, b: 219 } },
	{ val: 5000, color: { r: 241, g: 196, b: 15 } },
	{ val: 15000, color: { r: 230, g: 126, b: 34 } },
];

function rgbToHex(c: RGB): string {
	const r = Math.min(255, Math.round(c.r));
	const g = Math.min(255, Math.round(c.g));
	const b = Math.min(255, Math.round(c.b));
	return `#${r.toString(16).padStart(2, "0")}${g.toString(16).padStart(2, "0")}${b.toString(16).padStart(2, "0")}`;
}

function rgbLerp(c1: RGB, c2: RGB, t: number): RGB {
	return {
		r: c1.r + (c2.r - c1.r) * t,
		g: c1.g + (c2.g - c1.g) * t,
		b: c1.b + (c2.b - c1.b) * t,
	};
}

export function getRarityHex(favorites: number): string {
	if (!Number.isFinite(favorites) || favorites <= 0) {
		return rgbToHex(thresholdColors[0].color);
	}

	let i: number;
	for (i = thresholdColors.length - 1; i > 0; i--) {
		if (favorites >= thresholdColors[i].val) {
			break;
		}
	}

	if (i >= thresholdColors.length - 1) {
		return rgbToHex(thresholdColors[thresholdColors.length - 1].color);
	}

	const logVal = Math.log(favorites);
	const logLower = Math.log(thresholdColors[i].val);
	const logUpper = Math.log(thresholdColors[i + 1].val);

	if (logUpper === logLower) {
		return rgbToHex(thresholdColors[i].color);
	}

	let t = (logVal - logLower) / (logUpper - logLower);
	t = Math.max(0, Math.min(1, t));

	const interpolated = rgbLerp(thresholdColors[i].color, thresholdColors[i + 1].color, t);
	return rgbToHex(interpolated);
}

export function formatFavorites(count: number): string {
	if (!Number.isFinite(count) || count < 0) {
		return "0";
	}

	if (count >= 1_000_000) {
		return `${(count / 1_000_000).toFixed(1).replace(/\.0$/, "")}m`;
	}
	if (count >= 1_000) {
		return `${(count / 1_000).toFixed(1).replace(/\.0$/, "")}k`;
	}
	return count.toString();
}