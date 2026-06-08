import { hydrateRoot } from "react-dom/client";
import { StrictMode, startTransition } from "react";
import { StartClient } from "@tanstack/react-start/client";

startTransition(() => {
    hydrateRoot(
        document,
        <StrictMode>
            <StartClient />
        </StrictMode>,
    );
});
