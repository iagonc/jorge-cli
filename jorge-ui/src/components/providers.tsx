"use client";

import { SWRConfig } from "swr";
import { ReactNode } from "react";

interface ProvidersProps {
  children: ReactNode;
}

export function Providers({ children }: ProvidersProps) {
  return (
    <SWRConfig
      value={{
        revalidateOnFocus: true,
        revalidateOnReconnect: true,
        dedupingInterval: 2000,
        errorRetryCount: 3,
        errorRetryInterval: 5000,
        onError: (error) => {
          console.error("SWR Error:", error);
        },
      }}
    >
      {children}
    </SWRConfig>
  );
}
