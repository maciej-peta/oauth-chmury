import fs from 'fs';
import path from 'path';
import type { NextConfig } from 'next';

// Helper function to read a secret file safely
function readSecret(filePath?: string): string | undefined {
  try {
    return filePath ? fs.readFileSync(filePath, 'utf8').trim() : undefined;
  } catch {
    return undefined;
  }
}

const nextConfig: NextConfig = {
  reactStrictMode: true,
  env: {
    NEXTAUTH_SECRET: process.env.NEXTAUTH_SECRET || readSecret(process.env.NEXTAUTH_SECRET_FILE),
    AUTH0_CLIENT_ID: process.env.AUTH0_CLIENT_ID,
    AUTH0_CLIENT_SECRET: process.env.AUTH0_CLIENT_SECRET || readSecret(process.env.AUTH0_CLIENT_SECRET_FILE),
    AUTH0_DOMAIN: process.env.AUTH0_DOMAIN,
  },
  webpack(config) {
    config.resolve.alias = {
      ...(config.resolve.alias || {}),
      '@': path.resolve(__dirname, 'src'),
    };
    return config;
  },
};

export default nextConfig;
