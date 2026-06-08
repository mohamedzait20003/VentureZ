import "@/index.css";
import appCss from '../index.css?url'

import { Suspense } from "react";
import type { ReactNode } from "react";
import { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { HeadContent, Scripts, Outlet, createRootRouteWithContext } from "@tanstack/react-router";

import Navbar from "@/components/common/Navbar";
import Footer from "@/components/common/Footer";
import { Skeleton } from "@/components/ui/skeleton";

const RootDocument = ({ children }: Readonly<{ children: ReactNode }>) => {
    return (
        <html>
            <head>
                <HeadContent />
            </head>
            <body>
                {children}
                <Scripts />
            </body>
        </html>
    );
};

const RootComponent = () => {
    return (
        <RootDocument>
            <div className="flex flex-col min-h-screen">
                <Navbar />
                <main className="flex-1 flex flex-col">
                    <Suspense fallback={<Skeleton />}>
                        <Outlet />
                    </Suspense>
                </main>
                <Footer />
            </div>
            <ReactQueryDevtools initialIsOpen={false} />
            <TanStackRouterDevtools />
        </RootDocument>
    );
};

export const rootRoute = createRootRouteWithContext<{queryClient: QueryClient}>()({
    head: () => ({
        meta: [
            {
                charSet: "utf-8",
            },
            {
                name: "viewport",
                content: "width=device-width, initial-scale=1",
            },
            {
                title: "Safarni",
            },
        ],
        links: [
            {
                rel: "stylesheet",
                href: appCss,
            },
            {
                rel: "icon",
                type: "image/svg+xml",
                href: "/src/assets/WebLogo.svg",
            },
        ],
    }),
    component: RootComponent,
});
