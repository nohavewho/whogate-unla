import { Button, Input, Card, CardBody, CardHeader } from "@heroui/react";
import { Icon } from '@iconify/react';
import axios from 'axios';
import React, { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import api from '../../services/api';
import { toast } from '../../utils/toast';

export function LoginPage() {
  const { t } = useTranslation();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [isInitialized, setIsInitialized] = useState<boolean | null>(null);
  const navigate = useNavigate();

  const checkInitialization = useCallback(async () => {
    try {
      const response = await api.get('/auth/init/status');
      setIsInitialized(response.data.initialized);
    } catch {
      toast.error(t('errors.check_system_status'));
    }
  }, [t]);

  useEffect(() => {
    // Check if already logged in
    const token = window.localStorage.getItem('token');
    if (token) {
      navigate('/');
      return;
    }

    // Check if system is initialized
    checkInitialization();
  }, [navigate, checkInitialization]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await api.post('/auth/login', { username, password });
      window.localStorage.setItem('token', response.data.token);
      toast.success(t('auth.login_success'));
      navigate('/');
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.error) {
        toast.error(error.response.data.error);
      } else {
        toast.error(t('auth.login_failed'));
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
      toast.success(t('auth.init_success'));
      setIsInitialized(true);
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.error) {
        toast.error(error.response.data.error);
      } else {
        toast.error(t('auth.init_failed'));
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
            {isInitialized ? t('auth.login_to_continue') : t('auth.set_admin_credentials')}
          </p>
        </CardHeader>
        <CardBody className="p-6">
          <form onSubmit={isInitialized ? handleLogin : handleInitialize} className="flex flex-col gap-4">
            <Input
              label={t('auth.username')}
              placeholder={t('auth.username_placeholder')}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              startContent={<Icon icon="lucide:user" className="text-default-400" />}
              required
            />
            <Input
              label={t('auth.password')}
              type="password"
              placeholder={t('auth.password_placeholder')}
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
              {isInitialized ? t('auth.login') : t('auth.initialize_system')}
            </Button>
          </form>
        </CardBody>
      </Card>
    </div>
  );
} 