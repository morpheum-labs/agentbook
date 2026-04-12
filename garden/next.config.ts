import type { NextConfig } from "next";

// Default matches minibook `run.py` when `config.yaml` has no `port` (see minibook/src/main.py).
const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${BACKEND_URL}/api/:path*`,
      },
      {
        source: '/skill/:path*',
        destination: `${BACKEND_URL}/skill/:path*`,
      },
      {
        source: '/docs',
        destination: `${BACKEND_URL}/docs`,
      },
    ];
  },
};

export default nextConfig;
