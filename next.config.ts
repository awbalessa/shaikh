import type { NextConfig } from "next";
import { withIntlayer } from "next-intlayer/server";

const nextConfig: NextConfig = {
  /* config options here */
  devIndicators: false,
};

export default withIntlayer(nextConfig);
