import {
  Table,
  TableHeader,
  TableColumn,
  TableBody,
  TableRow,
  TableCell,
  Button,
  Modal,
  ModalContent,
  ModalHeader,
  ModalBody,
  ModalFooter,
  useDisclosure,
  Input,
  Select,
  SelectItem,
  Switch,
} from '@heroui/react';
import { Icon } from '@iconify/react';
import  { useEffect, useState } from 'react';

import api from '../../services/api';
import { toast } from '../../utils/toast';

interface User {
  id: number;
  username: string;
  role: 'admin' | 'normal';
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

interface CreateUserForm {
  username: string;
  password: string;
  role: 'admin' | 'normal';
}

interface UpdateUserForm {
  username: string;
  password?: string;
  role?: 'admin' | 'normal';
  isActive?: boolean;
}

export function UserManagement() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [createForm, setCreateForm] = useState<CreateUserForm>({
    username: '',
    password: '',
    role: 'normal',
  });
  const [updateForm, setUpdateForm] = useState<UpdateUserForm>({
    username: '',
  });

  const {
    isOpen: isCreateOpen,
    onOpen: onCreateOpen,
    onClose: onCreateClose,
  } = useDisclosure();
  const {
    isOpen: isUpdateOpen,
    onOpen: onUpdateOpen,
    onClose: onUpdateClose,
  } = useDisclosure();

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await api.get('/auth/users');
      setUsers(response.data);
    } catch {
      toast.error('Failed to fetch user list');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = async () => {
    try {
      await api.post('/auth/users', createForm);
      toast.success('User created successfully');
      onCreateClose();
      fetchUsers();
      setCreateForm({ username: '', password: '', role: 'normal' });
    } catch {
      toast.error('Failed to create user');
    }
  };

  const handleUpdate = async () => {
    if (!selectedUser) return;
    try {
      await api.put('/auth/users', updateForm);
      toast.success('User updated successfully');
      onUpdateClose();
      fetchUsers();
      setSelectedUser(null);
      setUpdateForm({ username: '' });
    } catch {
      toast.error('Failed to update user');
    }
  };

  const handleDelete = async (username: string) => {
    if (!window.confirm('Are you sure you want to delete this user?')) return;
    try {
      await api.delete(`/auth/users/${username}`);
      toast.success('User deleted successfully');
      fetchUsers();
    } catch {
      toast.error('Failed to delete user');
    }
  };

  const handleEdit = (user: User) => {
    setSelectedUser(user);
    setUpdateForm({
      username: user.username,
      role: user.role,
      isActive: user.isActive,
    });
    onUpdateOpen();
  };

  const handleToggleStatus = async (user: User) => {
    try {
      await api.put('/auth/users', {
        username: user.username,
        isActive: !user.isActive,
      });
      toast.success(`User ${user.isActive ? 'disabled' : 'enabled'} successfully`);
      fetchUsers();
    } catch {
      toast.error(`Failed to ${user.isActive ? 'disable' : 'enable'} user`);
    }
  };

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">User Management</h1>
        <Button
          color="primary"
          startContent={<Icon icon="lucide:plus" />}
          onPress={onCreateOpen}
        >
          Create User
        </Button>
      </div>

      <Table aria-label="User List">
        <TableHeader>
          <TableColumn>Username</TableColumn>
          <TableColumn>Role</TableColumn>
          <TableColumn>Status</TableColumn>
          <TableColumn>Created At</TableColumn>
          <TableColumn>Actions</TableColumn>
        </TableHeader>
        <TableBody
          loadingContent={<div>Loading...</div>}
          loadingState={loading ? 'loading' : 'idle'}
        >
          {users.map((user) => (
            <TableRow key={user.id}>
              <TableCell>{user.username}</TableCell>
              <TableCell>{user.role === 'admin' ? 'Administrator' : 'Normal User'}</TableCell>
              <TableCell>
                <span
                  className={`px-2 py-1 rounded-full text-xs ${
                    user.isActive
                      ? 'bg-success-100 text-success-800'
                      : 'bg-danger-100 text-danger-800'
                  }`}
                >
                  {user.isActive ? 'Enabled' : 'Disabled'}
                </span>
              </TableCell>
              <TableCell>
                {new Date(user.createdAt).toLocaleString()}
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-4">
                  <div className="flex items-center gap-2">
                    <Switch
                      size="sm"
                      isSelected={user.isActive}
                      onValueChange={() => handleToggleStatus(user)}
                    />
                    <span className="text-sm text-gray-600">
                      {user.isActive ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="light"
                      onPress={() => handleEdit(user)}
                    >
                      Edit
                    </Button>
                    <Button
                      size="sm"
                      color="danger"
                      variant="light"
                      onPress={() => handleDelete(user.username)}
                    >
                      Delete
                    </Button>
                  </div>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>

      {/* Create User Modal */}
      <Modal isOpen={isCreateOpen} onClose={onCreateClose}>
        <ModalContent>
          <ModalHeader>Create User</ModalHeader>
          <ModalBody>
            <div className="flex flex-col gap-4">
              <Input
                label="Username"
                value={createForm.username}
                onChange={(e) =>
                  setCreateForm({ ...createForm, username: e.target.value })
                }
              />
              <Input
                label="Password"
                type="password"
                value={createForm.password}
                onChange={(e) =>
                  setCreateForm({ ...createForm, password: e.target.value })
                }
              />
              <Select
                label="Role"
                selectedKeys={[createForm.role]}
                onSelectionChange={(keys) =>
                  setCreateForm({
                    ...createForm,
                    role: Array.from(keys)[0] as 'admin' | 'normal',
                  })
                }
              >
                <SelectItem key="admin">Administrator</SelectItem>
                <SelectItem key="normal">Normal User</SelectItem>
              </Select>
            </div>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onCreateClose}>
              Cancel
            </Button>
            <Button color="primary" onPress={handleCreate}>
              Create
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* Update User Modal */}
      <Modal isOpen={isUpdateOpen} onClose={onUpdateClose}>
        <ModalContent>
          <ModalHeader>Edit User</ModalHeader>
          <ModalBody>
            <div className="flex flex-col gap-4">
              <Input
                label="Username"
                value={updateForm.username}
                isReadOnly
              />
              <Input
                label="Password"
                type="password"
                placeholder="Leave empty to keep current password"
                value={updateForm.password || ''}
                onChange={(e) =>
                  setUpdateForm({ ...updateForm, password: e.target.value })
                }
              />
              <Select
                label="Role"
                selectedKeys={[updateForm.role || '']}
                onSelectionChange={(keys) =>
                  setUpdateForm({
                    ...updateForm,
                    role: Array.from(keys)[0] as 'admin' | 'normal',
                  })
                }
              >
                <SelectItem key="admin">Administrator</SelectItem>
                <SelectItem key="normal">Normal User</SelectItem>
              </Select>
            </div>
          </ModalBody>
          <ModalFooter>
            <Button variant="light" onPress={onUpdateClose}>
              Cancel
            </Button>
            <Button color="primary" onPress={handleUpdate}>
              Save
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}
