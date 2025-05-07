import { Card, CardBody, Button, Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, useDisclosure, Chip, Dropdown, DropdownTrigger, DropdownMenu, DropdownItem } from "@heroui/react";
import { Icon } from '@iconify/react';
import Editor from '@monaco-editor/react';
import yaml from 'js-yaml';
import { configureMonacoYaml } from 'monaco-yaml';
import React from 'react';

import { getMCPServers, createMCPServer, updateMCPServer, deleteMCPServer, syncMCPServers } from '../../services/api';
import { toast } from '../../utils/toast';


import OpenAPIImport from './components/OpenAPIImport';

declare global {
  interface Window {
    monaco: {
      languages: {
        yaml: {
          yamlDefaults: {
            setDiagnosticsOptions: (options: { enableSchemaRequest: boolean; schemas: Array<{ uri: string; fileMatch: string[] }> }) => void;
          };
        };
      };
    };
  }
}

interface Gateway {
  name: string;
  config: string;
  parsedConfig?: {
    routers: Array<{
      server: string;
      prefix: string;
    }>;
    servers: Array<{
      name: string;
      namespace: string;
      description: string;
      allowedTools: string[];
    }>;
    tools: Array<{
      name: string;
      description: string;
      method: string;
    }>;
  };
}

interface ServerConfig {
  name: string;
  namespace: string;
  description: string;
  allowedTools: string[];
}

interface RouterConfig {
  server: string;
  prefix: string;
}

interface ToolConfig {
  name: string;
  description: string;
  method: string;
}

export function GatewayManager() {
  const {isOpen, onOpen, onOpenChange} = useDisclosure();
  const {isOpen: isCreateOpen, onOpen: onCreateOpen, onOpenChange: onCreateOpenChange} = useDisclosure();
  const {isOpen: isImportOpen, onOpen: onImportOpen, onOpenChange: onImportOpenChange} = useDisclosure();
  const [mcpservers, setMCPServers] = React.useState<Gateway[]>([]);
  const [currentMCPServer, setCurrentMCPServer] = React.useState<Gateway | null>(null);
  const [editConfig, setEditConfig] = React.useState('');
  const [newConfig, setNewConfig] = React.useState('');
  const [parsedMCPServers, setParsedMCPServers] = React.useState<Gateway[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [isDark, setIsDark] = React.useState(() => {
    return document.documentElement.classList.contains('dark');
  });

  // Listen for theme changes
  React.useEffect(() => {
    const observer = new globalThis.MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.attributeName === 'class') {
          setIsDark(document.documentElement.classList.contains('dark'));
        }
      });
    });

    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class']
    });

    return () => observer.disconnect();
  }, []);

  // Configure Monaco YAML
  React.useEffect(() => {
    const monaco = window.monaco;
    if (monaco) {
      configureMonacoYaml(monaco, {
        enableSchemaRequest: true,
        schemas: [
          {
            uri: '',
            fileMatch: ['*.yml', '*.yaml'],
          },
        ],
      });
    }
  }, []);

  // 获取 MCP servers 列表
  React.useEffect(() => {
    const fetchMCPServers = async () => {
      try {
        setIsLoading(true);
        const servers = await getMCPServers();
        setMCPServers(servers);
      } catch {
        toast.error('获取 MCP 服务器列表失败', {
          duration: 3000,
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchMCPServers();
  }, []);

  const handleEdit = (server: Gateway) => {
    setCurrentMCPServer(server);
    setEditConfig(server.config);
    onOpen();
  };

  const handleSave = async () => {
    try {
      // Validate YAML
      yaml.load(editConfig);

      if (currentMCPServer) {
        await updateMCPServer(currentMCPServer.name, editConfig);
        const servers = await getMCPServers();
        setMCPServers(servers);
        toast.success('配置已保存', {
          duration: 3000,
        });
      }
      onOpenChange();
    } catch {
      toast.error('Something went wrong', {
        duration: 3000,
      });
    }
  };

  const handleDelete = async (server: Gateway) => {
    try {
      await deleteMCPServer(server.name);
      const servers = await getMCPServers();
      setMCPServers(servers);
      toast.success('配置已删除', {
        duration: 3000,
      });
    } catch {
      toast.error('删除失败', {
        duration: 3000,
      });
    }
  };

  const handleSync = async () => {
    try {
      setIsLoading(true);
      await syncMCPServers();
      const servers = await getMCPServers();
      setMCPServers(servers);
      toast.success('配置已同步', {
        duration: 3000,
      });
    } catch {
      toast.error('同步失败', {
        duration: 3000,
      });
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(`已复制: ${text}`, {
        duration: 2000,
      });
    } catch {
      toast.error("复制失败，请手动复制", {
        duration: 2000,
      });
    }
  };

  const handleCreate = async () => {
    try {
      // Validate YAML format first
      try {
        yaml.load(newConfig);
      } catch {
        toast.error('Invalid YAML format', {
          duration: 3000,
        });
        return;
      }

      // If YAML is valid, proceed with creation
      await createMCPServer(newConfig);
      const servers = await getMCPServers();
      setMCPServers(servers);
      onCreateOpenChange();
      setNewConfig('');
      toast.success('创建成功', {
        duration: 3000,
      });
    } catch {
      toast.error('创建失败', {
        duration: 3000,
      });
    }
  };

  const handleImportSuccess = async () => {
    try {
      const servers = await getMCPServers();
      setMCPServers(servers);
      onImportOpenChange();
      toast.success('OpenAPI specification imported successfully', {
        duration: 3000,
      });
    } catch {
      toast.error('Failed to refresh server list', {
        duration: 3000,
      });
    }
  };

  React.useEffect(() => {
    const parseConfigs = () => {
      const parsed = mcpservers.map(server => {
        try {
          const config = yaml.load(server.config) as Gateway['parsedConfig'];
          return { ...server, parsedConfig: config };
        } catch {
          toast.error(`解析配置失败: ${server.name}`, {
            duration: 3000,
          });
          return server;
        }
      });
      setParsedMCPServers(parsed);
    };
    parseConfigs();
  }, [mcpservers]);

  const editorOptions = {
    minimap: { enabled: false },
    fontSize: 14,
    lineNumbers: 'on',
    roundedSelection: false,
    scrollBeyondLastLine: false,
    readOnly: false,
    automaticLayout: true,
    ...(isDark ? {
      'editor.background': '#1E2228',
      'editor.foreground': '#E5E7EB',
      'editor.lineHighlightBackground': '#23272E',
      'editor.selectionBackground': '#4C6BCF40',
      'editor.inactiveSelectionBackground': '#4C6BCF20',
      'editor.lineHighlightBorder': '#2D3238',
      'editorCursor.foreground': '#4C6BCF',
      'editorWhitespace.foreground': '#4B5563',
    } : {
      'editor.background': '#F8F9FA',
      'editor.foreground': '#0B0F1A',
      'editor.lineHighlightBackground': '#F1F3F5',
      'editor.selectionBackground': '#4C6BCF40',
      'editor.inactiveSelectionBackground': '#4C6BCF20',
      'editor.lineHighlightBorder': '#E5E7EB',
      'editorCursor.foreground': '#4C6BCF',
      'editorWhitespace.foreground': '#A6B6E8',
    })
  };

  return (
    <div className="container mx-auto p-4">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Gateway Manager</h1>
        <div className="flex gap-2">
          <Button
            color="primary"
            onPress={onCreateOpen}
            startContent={<Icon icon="material-symbols:add" />}
          >
            Create
          </Button>
          <Button
            color="secondary"
            variant="flat"
            onPress={onImportOpen}
            startContent={<Icon icon="material-symbols:upload" />}
            className="bg-purple-500 hover:bg-purple-600 text-white"
          >
            Import OpenAPI
          </Button>
          <Button
            color="default"
            onPress={handleSync}
            isLoading={isLoading}
            startContent={<Icon icon="material-symbols:sync" />}
          >
            Sync
          </Button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex justify-center items-center h-32">
          <Icon icon="lucide:loader-2" className="animate-spin text-2xl" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {(parsedMCPServers || []).map((server) => (
            <Card key={server.name} className="w-full hover:shadow-lg transition-shadow bg-card">
              <CardBody className="flex flex-col gap-3 p-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold truncate">{server.name}</h3>
                  <div className="flex gap-2">
                    <Button
                      isIconOnly
                      color="primary"
                      variant="light"
                      size="sm"
                      onPress={() => handleEdit(server)}
                    >
                      <Icon icon="lucide:edit" className="text-lg" />
                    </Button>
                    <Dropdown>
                      <DropdownTrigger>
                        <Button
                          isIconOnly
                          color="danger"
                          variant="light"
                          size="sm"
                        >
                          <Icon icon="lucide:more-vertical" className="text-lg" />
                        </Button>
                      </DropdownTrigger>
                      <DropdownMenu aria-label="Actions">
                        <DropdownItem
                          key="delete"
                          className="text-danger"
                          color="danger"
                          startContent={<Icon icon="lucide:trash-2" />}
                          onPress={() => handleDelete(server)}
                        >
                          删除
                        </DropdownItem>
                      </DropdownMenu>
                    </Dropdown>
                  </div>
                </div>

                {server.parsedConfig && (
                  <div className="space-y-3">
                    {(server.parsedConfig.servers || []).map((serverConfig: ServerConfig) => {
                      return (
                        <div key={serverConfig.name} className="space-y-3">
                          <div>
                            <h4 className="text-sm font-semibold truncate">{serverConfig.name}</h4>
                            <p className="text-sm text-default-500 line-clamp-2">{serverConfig.description}</p>
                          </div>

                          <div className="space-y-2">
                            <h4 className="text-sm font-semibold">Routing Configuration</h4>
                            <div className="flex flex-col gap-2">
                              {(server.parsedConfig?.routers ?? []).map((router: RouterConfig, idx: number) => (
                                <div key={idx} className="flex items-center gap-2">
                                  <Chip
                                    color="primary"
                                    variant="flat"
                                    size="sm"
                                    className="cursor-pointer hover:opacity-80 select-none"
                                    onClick={() => handleCopyToClipboard(router.prefix)}
                                  >
                                    {router.prefix}
                                  </Chip>
                                  <Icon icon="lucide:arrow-right" className="text-sm" />
                                  <Chip
                                    variant="flat"
                                    size="sm"
                                    className="cursor-pointer hover:opacity-80 select-none"
                                    onClick={() => handleCopyToClipboard(router.server)}
                                  >
                                    {router.server}
                                  </Chip>
                                </div>
                              ))}
                            </div>
                          </div>

                          <div className="space-y-3">
                            <div>
                              <h4 className="text-sm font-semibold mb-1">Enabled Tools:</h4>
                              <div className="flex flex-wrap gap-1">
                                {serverConfig.allowedTools.map((tool: string) => (
                                  <Chip
                                    key={tool}
                                    variant="flat"
                                    color="success"
                                    size="sm"
                                    className="truncate cursor-pointer hover:opacity-80 select-none"
                                    onClick={() => handleCopyToClipboard(tool)}
                                  >
                                    {tool}
                                  </Chip>
                                ))}
                              </div>
                            </div>

                            <div>
                              <h4 className="text-sm font-semibold mb-1">All Tools:</h4>
                              <div className="flex flex-wrap gap-1">
                                {(server.parsedConfig?.tools ?? []).map((tool: ToolConfig) => (
                                  <Chip
                                    key={tool.name}
                                    variant="flat"
                                    color="default"
                                    size="sm"
                                    className="truncate cursor-pointer hover:opacity-80 select-none"
                                    onClick={() => handleCopyToClipboard(tool.name)}
                                  >
                                    {tool.name}
                                  </Chip>
                                ))}
                              </div>
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </CardBody>
            </Card>
          ))}
        </div>
      )}

      <Modal
        isOpen={isOpen}
        onOpenChange={onOpenChange}
        size="3xl"
        className="w-[70%] h-[70%]"
      >
        <ModalContent className="h-[70%]">
          {(onClose) => (
            <>
              <ModalHeader>Edit MCP Server Configuration</ModalHeader>
              <ModalBody className="flex-1">
                <Editor
                  height="100%"
                  defaultLanguage="yaml"
                  value={editConfig}
                  onChange={(value) => setEditConfig(value || '')}
                  theme={isDark ? "vs-dark" : "vs"}
                  options={editorOptions}
                />
              </ModalBody>
              <ModalFooter>
                <Button color="danger" variant="light" onPress={onClose}>
                  Cancel
                </Button>
                <Button color="primary" onPress={handleSave}>
                  Save Changes
                </Button>
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>

      <Modal
        isOpen={isCreateOpen}
        onOpenChange={onCreateOpenChange}
        size="3xl"
        className="w-[70%] h-[70%]"
      >
        <ModalContent className="h-[70%]">
          {(onClose) => (
            <>
              <ModalHeader>Add New MCP Server Configuration</ModalHeader>
              <ModalBody className="flex-1">
                <Editor
                  height="100%"
                  defaultLanguage="yaml"
                  value={newConfig}
                  onChange={(value) => setNewConfig(value || '')}
                  theme={isDark ? "vs-dark" : "vs"}
                  options={editorOptions}
                />
              </ModalBody>
              <ModalFooter>
                <Button color="danger" variant="light" onPress={onClose}>
                  Cancel
                </Button>
                <Button color="primary" onPress={handleCreate}>
                  Create
                </Button>
              </ModalFooter>
            </>
          )}
        </ModalContent>
      </Modal>

      <Modal isOpen={isImportOpen} onOpenChange={onImportOpenChange} size="2xl">
        <ModalContent>
          <ModalHeader>Import OpenAPI Specification</ModalHeader>
          <ModalBody>
            <OpenAPIImport onSuccess={handleImportSuccess} />
          </ModalBody>
          <ModalFooter>
            <Button color="danger" variant="light" onPress={() => onImportOpenChange()}>
              Cancel
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </div>
  );
}
