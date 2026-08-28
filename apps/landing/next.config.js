/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: {
    unoptimized: true,
  },
  // Ensure the base path works if this is deployed to a subpath like github.io/dboke
  // If deployed to a custom domain, this isn't needed. Assuming github.io/repo for now.
  basePath: '/dboke',
};

module.exports = nextConfig;
