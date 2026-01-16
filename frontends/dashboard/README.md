# IoT Platform Dashboard

Dashboard frontend for the IoT Platform, built with React 19, TypeScript, and shadcn/ui (Base UI).

## Tech Stack

- **React 19** - UI library
- **TypeScript 5.9** - Type safety
- **Vite 7** - Build tool
- **Tailwind CSS 4** - Styling
- **shadcn/ui (Base UI)** - Component library
- **Tabler Icons** - Icon set
- **React Router 7** - Routing

## Getting Started

### Prerequisites

- Node.js >= 20.19 or >= 22.12
- npm or pnpm

### Installation

```bash
npm install
```

### Development

```bash
npm run dev
```

### Build

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

### Lint

```bash
npm run lint
```

## Project Structure

```
src/
├── assets/                      # Static assets (images, fonts, etc.)
│
├── components/
│   ├── ui/                      # shadcn/ui primitives (do not edit manually)
│   └── shared/                  # Shared business components
│
├── features/                    # Feature modules (business logic)
│   └── [feature]/
│       ├── components/          # Feature-specific components
│       ├── pages/               # Feature pages
│       ├── hooks/               # Feature-specific hooks
│       ├── types/               # Feature-specific types
│       └── index.ts             # Public exports
│
├── layouts/                     # Layout components
│   └── [layout-name]/
│       ├── index.tsx            # Layout export
│       └── ...                  # Layout-specific components
│
├── hooks/                       # Global hooks
├── lib/                         # Utilities
├── types/                       # Global types
│
├── App.tsx                      # App entry point
├── main.tsx                     # React bootstrap
└── index.css                    # Global styles
```

## Architecture Principles

### Feature-Based Architecture

Each feature is self-contained with its own components, pages, hooks, and types. Features expose only what's necessary through their `index.ts`.

```typescript
// features/dashboard/index.ts
export { DashboardPage } from "./pages/dashboard-page"
```

### Layouts vs Features

- **Layouts** (`layouts/`): Infrastructure, page structure, navigation
- **Features** (`features/`): Business logic, domain-specific code

### Component Organization

- `components/ui/`: shadcn/ui primitives - generated, do not edit
- `components/shared/`: Reusable business components across features
- `features/[name]/components/`: Feature-specific components

## Adding shadcn/ui Components

```bash
npx shadcn@latest add [component-name]
```

Components are configured to use Base UI style. See `components.json` for configuration.

## License

Private