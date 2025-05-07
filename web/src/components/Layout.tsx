import {
  Navbar,
  NavbarContent,
  NavbarItem,
  Button,
  Link as HeroLink,
  Avatar,
  Tooltip,
  Dropdown,
  DropdownTrigger,
  DropdownMenu,
  DropdownItem
} from "@heroui/react";
import { Icon } from '@iconify/react';
import React, { useEffect, useState } from 'react';
import { useLocation, Link, useNavigate } from 'react-router-dom';

import { getCurrentUser } from '../api/auth';

import { ChangePasswordDialog } from './ChangePasswordDialog';
import { WechatQRCode } from './WechatQRCode';

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const [isCollapsed, setIsCollapsed] = React.useState(false);
  const [isDark, setIsDark] = React.useState(() => {
    const savedTheme = window.localStorage.getItem('theme');
    return savedTheme === 'dark';
  });
  const [isChangePasswordOpen, setIsChangePasswordOpen] = React.useState(false);
  const [isWechatQRCodeOpen, setIsWechatQRCodeOpen] = React.useState(false);
  const [userInfo, setUserInfo] = useState<{ username: string; role: string } | null>(null);

  // Initialize theme on mount
  React.useEffect(() => {
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [isDark]);

  useEffect(() => {
    const fetchUserInfo = async () => {
      try {
        const response = await getCurrentUser();
        console.log('User info response:', response);
        setUserInfo(response.data);
      } catch (error) {
        console.error('Failed to fetch user info:', error);
      }
    };

    fetchUserInfo();
  }, []);

  const menuItems = [
    {
      key: 'chat',
      label: 'Chat Playground',
      icon: 'lucide:message-square',
      path: '/chat',
    },
    {
      key: 'gateway',
      label: 'Gateway Manager',
      icon: 'lucide:server',
      path: '/gateway',
    },
    ...(userInfo?.role === 'admin' ? [{
      key: 'users',
      label: 'User Management',
      icon: 'lucide:users',
      path: '/users',
    }] : []),
  ];

  const handleLogout = () => {
    window.localStorage.removeItem('token');
    navigate('/login');
  };

  const toggleTheme = () => {
    setIsDark(!isDark);
    document.documentElement.classList.toggle('dark');
    window.localStorage.setItem('theme', !isDark ? 'dark' : 'light');
  };

  return (
    <div className="min-h-screen bg-background text-foreground">
      {/* Top Navigation Bar */}
      <Navbar
        className="bg-card border-b border-border shadow-sm"
        maxWidth="full"
        height="4rem"
      >
        <NavbarContent justify="end" className="gap-4">
          <NavbarItem>
            <Tooltip content="加入微信群">
              <Button
                variant="light"
                isIconOnly
                onPress={() => setIsWechatQRCodeOpen(true)}
              >
                <Icon icon="mdi:wechat" className="text-2xl" />
              </Button>
            </Tooltip>
          </NavbarItem>
          <NavbarItem>
            <Tooltip content="Join Discord">
              <Button
                as={HeroLink}
                href="https://discord.gg/udf69cT9TY"
                target="_blank"
                variant="light"
                isIconOnly
              >
                <Icon icon="ic:baseline-discord" className="text-2xl" />
              </Button>
            </Tooltip>
          </NavbarItem>
          <NavbarItem>
            <Tooltip content="View on GitHub">
              <Button
                as={HeroLink}
                href="https://github.com/mcp-ecosystem/mcp-gateway"
                target="_blank"
                variant="light"
                isIconOnly
              >
                <Icon icon="mdi:github" className="text-2xl" />
              </Button>
            </Tooltip>
          </NavbarItem>
          <NavbarItem>
            <Tooltip content={`Switch to ${isDark ? "light" : "dark"} mode`}>
              <Button
                variant="light"
                isIconOnly
                onPress={toggleTheme}
              >
                <Icon
                  icon={isDark ? "lucide:sun" : "lucide:moon"}
                  className="text-2xl"
                />
              </Button>
            </Tooltip>
          </NavbarItem>
        </NavbarContent>
      </Navbar>

      <div className="flex h-[calc(100vh-4rem)]">
        {/* Sidebar */}
        <div
          className={`h-screen bg-card text-foreground flex flex-col fixed left-0 top-0 z-40 transition-all duration-300 border-r border-border shadow-lg ${
            isCollapsed ? "w-20" : "w-64"
          }`}
        >
          <div className="flex items-center justify-between p-4 border-b border-border h-16">
            {!isCollapsed && <span className="text-xl font-bold">MCP Admin</span>}
            <Button
              isIconOnly
              variant="light"
              onPress={() => setIsCollapsed(!isCollapsed)}
              aria-label="Toggle sidebar"
            >
              <Icon icon={isCollapsed ? "lucide:chevron-right" : "lucide:chevron-left"} />
            </Button>
          </div>

          <nav className="flex-1 overflow-y-auto p-2">
            {menuItems.map((item) => (
              isCollapsed ? (
                <Tooltip
                  key={item.path}
                  content={item.label}
                  placement="right"
                >
                  <Link
                    to={item.path}
                    className={`flex items-center w-full px-4 py-2 rounded-lg mb-1 ${
                      (item.path === "/"
                        ? location.pathname === "/"
                        : location.pathname.startsWith(item.path))
                        ? 'bg-primary/10 text-primary'
                        : 'hover:bg-accent text-foreground'
                    }`}
                  >
                    <Icon icon={item.icon} className="text-xl" />
                  </Link>
                </Tooltip>
              ) : (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`flex items-center w-full px-4 py-2 rounded-lg mb-1 ${
                    (item.path === "/"
                      ? location.pathname === "/"
                      : location.pathname.startsWith(item.path))
                      ? 'bg-primary/10 text-primary'
                      : 'hover:bg-accent text-foreground'
                  }`}
                >
                  <Icon icon={item.icon} className="text-xl" />
                  <span className="ml-2">{item.label}</span>
                </Link>
              )
            ))}
          </nav>

          <div className="border-t border-border p-4">
            <Dropdown placement="top-end">
              <DropdownTrigger>
                <div className="flex items-center gap-3 cursor-pointer hover:bg-accent p-2 rounded-lg transition-colors">
                  <Avatar
                    icon={<Icon icon="lucide:user" className="text-xl" />}
                    name={userInfo?.username || 'User'}
                    size="sm"
                    color="primary"
                  />
                  {!isCollapsed && (
                    <div className="flex-1">
                      <p className="text-sm font-semibold text-foreground">
                        {userInfo?.username || 'Loading...'}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {userInfo?.role === 'admin' ? 'Administrator' : 'Normal User'}
                      </p>
                    </div>
                  )}
                </div>
              </DropdownTrigger>
              <DropdownMenu aria-label="User Actions">
                <DropdownItem
                  key="change-password"
                  startContent={<Icon icon="lucide:key" className="text-xl" />}
                  onPress={() => setIsChangePasswordOpen(true)}
                >
                  修改密码
                </DropdownItem>
                <DropdownItem
                  key="logout"
                  className="text-destructive"
                  color="danger"
                  startContent={<Icon icon="lucide:log-out" className="text-xl" />}
                  onPress={handleLogout}
                >
                  退出登录
                </DropdownItem>
              </DropdownMenu>
            </Dropdown>
          </div>
        </div>

        {/* Main Content */}
        <main className={`flex-1 overflow-auto transition-all duration-300 ${isCollapsed ? "ml-20" : "ml-64"} bg-background text-foreground p-6`}>
          {children}
        </main>
      </div>

      {/* Change Password Dialog */}
      <ChangePasswordDialog
        isOpen={isChangePasswordOpen}
        onOpenChange={() => setIsChangePasswordOpen(false)}
      />

      {/* WeChat QR Code Dialog */}
      <WechatQRCode
        isOpen={isWechatQRCodeOpen}
        onOpenChange={setIsWechatQRCodeOpen}
      />
    </div>
  );
}
