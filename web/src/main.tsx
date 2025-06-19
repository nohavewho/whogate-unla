import {HeroUIProvider, ToastProvider} from "@heroui/react";
import React from "react";
import ReactDOM from "react-dom/client";

import App from "./App.tsx";
import './i18n';

import "./index.css";

// Initialize theme immediately before React renders
const savedTheme = window.localStorage.getItem('theme');
if (savedTheme === 'dark') {
  document.documentElement.classList.add('dark');
} else if (savedTheme === 'light') {
  document.documentElement.classList.remove('dark');
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <HeroUIProvider>
      <ToastProvider placement="bottom-right" />
      <main className="text-foreground bg-background h-screen overflow-hidden">
        <App />
      </main>
    </HeroUIProvider>
  </React.StrictMode>,
);
