// This file is disabled because Next.js static export does not support dynamic [id] routes
// without pre-rendering all possible IDs. We have migrated this logic to /databases/page.tsx
// using URL query parameters (?id=) instead of path parameters.

export function generateStaticParams() {
  return [{ id: 'dummy' }];
}

export default function DeprecatedDynamicRoute() {
  return null;
}
