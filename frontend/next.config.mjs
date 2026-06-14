/** @type {import('next').NextConfig} */
const nextConfig = {
    output: 'export',
    // 根路径访问，不带 /web 前缀
    basePath: '',
    assetPrefix: '',
    typescript: {
        ignoreBuildErrors: true,
    },
    images: {
        unoptimized: true,
    },
}

export default nextConfig
