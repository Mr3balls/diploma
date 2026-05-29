import { RouterProvider } from "react-router-dom";
import { Toaster } from "sonner";
import { AuthProvider } from "@/app/providers/auth-provider";
import { LangProvider } from "@/app/providers/lang-provider";
import { QueryProvider } from "@/app/providers/query-provider";
import { router } from "@/app/router";

export function App() {
  return (
    <QueryProvider>
      <LangProvider>
        <AuthProvider>
          <RouterProvider router={router} />
          <Toaster richColors position="top-right" />
        </AuthProvider>
      </LangProvider>
    </QueryProvider>
  );
}