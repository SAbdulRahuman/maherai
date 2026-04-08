import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  basePath: "/ui",
  assetPrefix: "/ui",
  trailingSlash: true,
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
