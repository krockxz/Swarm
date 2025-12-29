import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Allow connections to backend API
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:8080/api/:path*",
      },
      {
        source: "/ws",
        destination: "http://localhost:8080/ws",
      },
    ];
  },
};

export default nextConfig;
