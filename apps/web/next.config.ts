import type { NextConfig } from "next"

const nextConfig: NextConfig = {
    reactStrictMode: true,
    allowedDevOrigins: ["redbird-trusting-macaque.ngrok-free.app"],
    async rewrites() {
        return [
            {
                source: "/api/:path*",
                destination: "http://127.0.0.1:8000/api/:path*",
            },
        ];
    },
};

export default nextConfig;
