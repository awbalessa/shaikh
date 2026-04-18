import type { FC, PropsWithChildren } from "react";
import "./globals.css";
import "streamdown/styles.css";

const RootLayout: FC<PropsWithChildren> = ({ children }) => <>{children}</>;

export default RootLayout;
