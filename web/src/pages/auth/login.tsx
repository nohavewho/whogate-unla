import { Button, Input, Card, CardBody, CardHeader } from "@heroui/react";
import { Icon } from '@iconify/react';
import axios from 'axios';
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

import api from '../../services/api';
import { toast } from '../../utils/toast';

export function LoginPage() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Check if already logged in
    const token = window.localStorage.getItem('token');
    if (token) {
      navigate('/');
      return;
    }

    // Check if system is initialized
    checkInitialization();
  }, [navigate]);

  const checkInitialization = async () => {
    try {
      const response = await api.get('/auth/init/status');
      setIsInitialized(response.data.initialized);
    } catch {
      toast.error('检查系统状态失败');
    }
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await api.post('/auth/login', { username, password });
      window.localStorage.setItem('token', response.data.token);
      toast.success('登录成功');
      navigate('/');
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.error) {
        toast.error(error.response.data.error);
      } else {
        toast.error('登录失败，请重试');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleInitialize = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await api.post('/auth/init', { username, password });
      toast.success('系统初始化成功，请登录');
      setIsInitialized(true);
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.error) {
        toast.error(error.response.data.error);
      } else {
        toast.error('初始化失败，请重试');
      }
    } finally {
      setLoading(false);
    }
  };

  if (isInitialized === null) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Card className="w-full max-w-md">
          <CardBody className="p-6">
            <div className="flex items-center justify-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
          </CardBody>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="flex flex-col gap-1.5 p-6">
          <div className="flex items-center gap-2">
            <Icon icon="lucide:server" className="text-primary text-2xl" />
            <h1 className="text-2xl font-bold">MCP Admin</h1>
          </div>
          <p className="text-default-500">
            {isInitialized ? '请登录以继续' : '请设置管理员账号密码'}
          </p>
        </CardHeader>
        <CardBody className="p-6">
          <form onSubmit={isInitialized ? handleLogin : handleInitialize} className="flex flex-col gap-4">
            <Input
              label="用户名"
              placeholder="请输入用户名"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              startContent={<Icon icon="lucide:user" className="text-default-400" />}
              required
            />
            <Input
              label="密码"
              type="password"
              placeholder="请输入密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              startContent={<Icon icon="lucide:lock" className="text-default-400" />}
              required
            />
            <Button
              type="submit"
              color="primary"
              isLoading={loading}
              className="w-full"
            >
              {isInitialized ? '登录' : '初始化系统'}
            </Button>
          </form>
        </CardBody>
      </Card>
    </div>
  );
} 