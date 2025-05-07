import { Button, Input, Modal, ModalContent, ModalHeader, ModalBody, ModalFooter } from "@heroui/react";
import { Icon } from '@iconify/react';
import axios from 'axios';
import React, { useState } from 'react';
import toast from 'react-hot-toast';

import api from '../services/api';

interface ChangePasswordDialogProps {
  isOpen: boolean;
  onOpenChange: () => void;
}

export function ChangePasswordDialog({ isOpen, onOpenChange }: ChangePasswordDialogProps) {
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    if (newPassword !== confirmPassword) {
      toast.error('两次输入的新密码不一致');
      return;
    }

    setLoading(true);
    try {
      await api.post('/auth/change-password', {
        oldPassword,
        newPassword,
      });
      toast.success('密码修改成功');
      onOpenChange();
      // Clear form
      setOldPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (error) {
      if (axios.isAxiosError(error) && error.response?.data?.error) {
        toast.error(error.response.data.error);
      } else {
        toast.error('修改密码失败，请重试');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onOpenChange={onOpenChange}>
      <ModalContent>
        <ModalHeader>修改密码</ModalHeader>
        <ModalBody>
          <div className="flex flex-col gap-4">
            <Input
              label="当前密码"
              type="password"
              placeholder="请输入当前密码"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              startContent={<Icon icon="lucide:lock" className="text-default-400" />}
            />
            <Input
              label="新密码"
              type="password"
              placeholder="请输入新密码"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              startContent={<Icon icon="lucide:lock" className="text-default-400" />}
            />
            <Input
              label="确认新密码"
              type="password"
              placeholder="请再次输入新密码"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              startContent={<Icon icon="lucide:lock" className="text-default-400" />}
            />
          </div>
        </ModalBody>
        <ModalFooter>
          <Button color="danger" variant="light" onPress={onOpenChange}>
            取消
          </Button>
          <Button color="primary" onPress={handleSubmit} isLoading={loading}>
            确认修改
          </Button>
        </ModalFooter>
      </ModalContent>
    </Modal>
  );
} 