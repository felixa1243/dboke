# Dboke Web (Frontend)

This is the frontend component for Dboke, built with **Next.js 14+** (App Router) and **Tailwind CSS v4**.

## Architecture & Styling

- **`app/`**: Next.js App Router directories (`page.tsx`, `layout.tsx`).
- **`app/globals.css`**: Tailwind v4 configuration using the `@theme` directive, defining custom animations (`animate-blob`) and design tokens.
- **Design System**: Employs a premium "Glassmorphism" aesthetic with smooth CSS transitions, dark mode palettes (`slate-950`), and dynamic gradient backgrounds.

## Tailwind CSS v4 Notice

This project utilizes the next-generation **Tailwind CSS v4**. 
There is **no** `tailwind.config.ts` or `tailwind.config.js` file in this directory. 

All utility classes, theme variables (like `--color-primary`), and keyframe animations are configured natively in CSS inside `app/globals.css` using standard CSS nesting and the new `@theme` directive.

## Development

To run the Next.js development server:

```bash
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.
