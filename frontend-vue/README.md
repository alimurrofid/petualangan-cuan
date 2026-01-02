# Petualangan Cuan - Frontend (Vue 3)

Frontend application for Petualangan Cuan, built with Vue 3, TypeScript, and Vite.

## Tech Stack

- **Framework:** Vue 3
- **Build Tool:** Vite
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **State Management:** Pinia
- **Icons:** Lucide Vue
- **Package Manager:** pnpm

## Prerequisites

- [Node.js](https://nodejs.org/) (LTS recommended)
- [pnpm](https://pnpm.io/) (Package manager)

## Installation

1. Navigate to the frontend directory:
   ```bash
   cd frontend-vue
   ```

2. Install dependencies:
   ```bash
   pnpm install
   ```

## Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Configure environment variables (e.g., API Base URL).

## Development

Start the development server:

```bash
pnpm dev
```

The application will be available at `http://localhost:5173` (by default).

## Production Build

Build the application for production:

```bash
pnpm build
```

Preview the production build locally:

```bash
pnpm preview
```

## Project Structure

```bash
frontend-vue/
├── src/
│   ├── assets/          # Static assets (images, fonts)
│   ├── components/      # Reusable UI components
│   ├── composables/     # Vue composables
│   ├── layouts/         # Layout components
│   ├── lib/             # Utilities and configurations
│   ├── router/          # Vue Router configuration
│   ├── stores/          # Pinia state stores
│   ├── views/           # Page components
│   │   ├── auth/        # Authentication pages
│   │   ├── calendar/    # Calendar view
│   │   ├── category/    # Category management
│   │   ├── transaction/ # Transaction management
│   │   ├── wallet/      # Wallet management
│   │   └── ...
│   ├── App.vue          # Root component
│   └── main.ts          # Application entry point
├── public/              # Public static files
├── index.html           # HTML entry point
├── package.json         # Dependencies and scripts
├── tsconfig.json        # TypeScript configuration
└── vite.config.ts       # Vite configuration
```

