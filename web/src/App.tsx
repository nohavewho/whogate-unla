import {
  Navbar,
  NavbarBrand,
  NavbarContent,
  NavbarItem,
  NavbarMenuToggle,
  NavbarMenu,
  NavbarMenuItem,
  Link as HeroLink
} from "@heroui/react";
import { Icon } from '@iconify/react';
import React from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation, Link } from 'react-router-dom';

import { ChatInterface } from "./pages/chat/chat-interface";
import { GatewayManager } from "./pages/gateway/gateway-manager";

function Navigation() {
  const location = useLocation();
  const [isMenuOpen, setIsMenuOpen] = React.useState(false);

  const menuItems = [
    { to: "/", icon: "lucide:server", label: "Gateway Manager" },
    { to: "/chat", icon: "lucide:message-square", label: "MCP Chat" }
  ];

  return (
    <Navbar
      isBordered
      isMenuOpen={isMenuOpen}
      onMenuOpenChange={setIsMenuOpen}
      className="bg-background"
      maxWidth="full"
    >
      <NavbarContent justify="start" className="gap-2">
        <NavbarMenuToggle
          aria-label={isMenuOpen ? "Close menu" : "Open menu"}
          className="sm:hidden"
        />
        <NavbarBrand className="gap-2">
          <Icon icon="lucide:server" className="text-primary text-xl" />
          <p className="font-bold text-inherit">MCP Admin</p>
        </NavbarBrand>
      </NavbarContent>

      <NavbarContent className="hidden sm:flex gap-4" justify="center">
        {menuItems.map((item) => (
          <NavbarItem key={item.to} isActive={
            item.to === "/"
              ? location.pathname === "/"
              : location.pathname.startsWith(item.to)
          }>
            <Link
              to={item.to}
              className={`flex items-center gap-2 px-3 py-1.5 rounded-lg transition-colors ${
                (item.to === "/" 
                  ? location.pathname === "/"
                  : location.pathname.startsWith(item.to))
                  ? 'bg-primary/10 text-primary' 
                  : 'hover:bg-default-100'
              }`}
            >
              <Icon icon={item.icon} className="text-xl" />
              {item.label}
            </Link>
          </NavbarItem>
        ))}
      </NavbarContent>

      <NavbarContent justify="end" className="gap-2">
        <NavbarItem>
          <HeroLink
            href="https://github.com/mcp-ecosystem/mcp-gateway"
            target="_blank"
            className="relative tap-highlight-transparent outline-none data-[focus-visible=true]:z-10 data-[focus-visible=true]:outline-2 data-[focus-visible=true]:outline-focus data-[focus-visible=true]:outline-offset-2 text-medium no-underline hover:opacity-90 active:opacity-disabled transition-opacity"
          >
            <Icon icon="ri:github-fill" className="text-4xl text-black" />
          </HeroLink>
        </NavbarItem>
      </NavbarContent>

      <NavbarMenu>
        {menuItems.map((item) => (
          <NavbarMenuItem key={item.to}>
            <Link
              to={item.to}
              className={`flex items-center gap-2 w-full ${
                (item.to === "/" 
                  ? location.pathname === "/"
                  : location.pathname.startsWith(item.to)) 
                  ? 'text-primary' : ''
              }`}
              onClick={() => setIsMenuOpen(false)}
            >
              <Icon icon={item.icon} className="text-xl" />
              {item.label}
            </Link>
          </NavbarMenuItem>
        ))}
        <NavbarMenuItem>
          <HeroLink
            href="https://github.com/mcp-ecosystem/mcp-gateway"
            target="_blank"
            className="flex items-center gap-2 w-full"
          >
            <Icon icon="ri:github-fill" className="text-xl" />
            GitHub
          </HeroLink>
        </NavbarMenuItem>
      </NavbarMenu>
    </Navbar>
  );
}

function AppContent() {
  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main className="container mx-auto px-4 sm:px-6 py-8 h-[calc(100vh-4rem)] overflow-auto scrollbar-hide">
        <Routes>
          <Route path="/" element={<GatewayManager />} />
          <Route path="/chat" element={<ChatInterface />} />
          <Route path="/chat/:sessionId" element={<ChatInterface />} />
        </Routes>
      </main>
    </div>
  );
}

export default function App() {
  return (
    <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <AppContent />
    </Router>
  );
}
