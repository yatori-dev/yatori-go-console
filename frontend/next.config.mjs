/** @type {import('next').NextConfig} */
const nextConfig = {
    output: 'export',
    basePath: '/web',
    assetPrefix: '/web',
    typescript: {
        ignoreBuildErrors: true,
    },
    images: {
        unoptimized: true,
    },
}

export default nextConfig
