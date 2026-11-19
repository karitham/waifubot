# Waifu Bot Frontend

Web interface for viewing character collections and wishlists.

## Development

```bash
# From project root, enter nix shell
nix develop

# Or if using direnv
direnv allow

# Start dev server
cd frontend
npm run dev
```

The frontend will be available at `http://localhost:5173`.

## Build

```bash
cd frontend
npm install
npm run build
```

This produces a static site in the `dist/` directory.

## Dependency notes

The frontend uses Solid 2. Two packages are pinned deliberately:

- `@kobalte/core` is on the Solid 2 alpha; its peer range still names an RC, so
  `frontend/.npmrc` enables `legacy-peer-deps` to keep installs working.
- `@tanstack/virtual-core` stays at 3.13.23, the version the previous adapter
  used. `frontend/src/utils/windowVirtualizer.ts` is the Solid 2 bridge; newer
  core versions change the mounted-window size.

## API Configuration

The frontend fetches data from the waifubot API. The API URL is set at build time:

```bash
# Default: https://waifuapi.karitham.dev
VITE_API_URL=https://your-api.example.com npm run build
```

## Deployment

Deploy the `dist/` directory to any static hosting:

- [Cloudflare Pages](https://pages.cloudflare.com/)
- [Vercel](https://vercel.com)
- [Netlify](https://www.netlify.com)
- Nginx, Apache, etc.

Example for Cloudflare Pages:

1. Connect your GitHub repository
2. Build command: `npm run build`
3. Output directory: `dist`
