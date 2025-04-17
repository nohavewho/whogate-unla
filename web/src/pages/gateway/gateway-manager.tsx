import React from 'react';
import { Card, CardBody, Button, Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, useDisclosure } from "@heroui/react";
import { Icon } from '@iconify/react';
import yaml from 'js-yaml';
import { Accordion, AccordionItem, Chip } from "@heroui/react";
import { getMCPServers } from '../../services/api';
import Editor from '@monaco-editor/react';
import { configureMonacoYaml } from 'monaco-yaml';

declare global {
  interface Window {
    monaco: any;
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
  const [mcpservers, setMCPServers] = React.useState<Gateway[]>([]);
  const [currentMCPServer, setCurrentMCPServer] = React.useState<Gateway | null>(null);
  const [editConfig, setEditConfig] = React.useState('');
  const [parsedMCPServers, setParsedMCPServers] = React.useState<Gateway[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [selectedMCPServer, setSelectedMCPServer] = React.useState<Gateway | null>(null);

  // Configure Monaco YAML
  React.useEffect(() => {
    const monaco = window.monaco;
    if (monaco) {
      configureMonacoYaml(monaco, {
        enableSchemaRequest: true,
        schemas: [
          {
            uri: 'https://raw.githubusercontent.com/mcp-ecosystem/mcp-gateway/main/schema/gateway.json',
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
      } catch (error) {
        console.error('Failed to fetch MCP servers:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchMCPServers();
  }, []);

  const handleEdit = (server: Gateway) => {
    console.log('Editing MCP server:', server);
    setCurrentMCPServer(server);
    setEditConfig(server.config);
    onOpen();
  };

  const handleSave = () => {
    try {
      // Validate YAML
      yaml.load(editConfig);

      if (currentMCPServer) {
        setMCPServers(mcpservers.map(s =>
          s.name === currentMCPServer.name ? {...s, config: editConfig} : s
        ));
      }
      onOpenChange();
    } catch (e) {
      alert('Invalid YAML format');
    }
  };

  const handleSync = async () => {
    // TODO: Implement actual sync logic
    alert("Configuration sync triggered");
  };

  React.useEffect(() => {
    const parseConfigs = () => {
      const parsed = mcpservers.map(server => {
        try {
          const config = yaml.load(server.config) as Gateway['parsedConfig'];
          return { ...server, parsedConfig: config };
        } catch (e) {
          console.error(`Failed to parse config for ${server.name}:`, e);
          return server;
        }
      });
      setParsedMCPServers(parsed);
    };
    parseConfigs();
  }, [mcpservers]);

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">MCP Servers Manager</h1>
        <Button
          color="primary"
          startContent={<Icon icon="lucide:refresh-cw" />}
          onPress={handleSync}
          isLoading={isLoading}
        >
          Sync Configuration
        </Button>
      </div>

      {isLoading ? (
        <div className="flex justify-center items-center h-32">
          <Icon icon="lucide:loader-2" className="animate-spin text-2xl" />
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {(parsedMCPServers || []).map((server) => (
            <Card key={server.name} className="w-full">
              <CardBody className="flex flex-col gap-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold">{server.name}</h3>
                  <Button
                    isIconOnly
                    color="primary"
                    variant="light"
                    onPress={() => handleEdit(server)}
                  >
                    <Icon icon="lucide:edit" className="text-lg" />
                  </Button>
                </div>

                {server.parsedConfig && (
                  <div className="space-y-4">
                    {(server.parsedConfig.servers || []).map((serverConfig: ServerConfig) => {
                      return (
                        <div key={serverConfig.name} className="space-y-4">
                          <div>
                            <h4 className="text-sm font-semibold">{serverConfig.name}</h4>
                            <p className="text-sm text-default-500">{serverConfig.description}</p>
                          </div>

                          <div className="space-y-2">
                            <h4 className="text-sm font-semibold">Routing Configuration</h4>
                            {(server.parsedConfig?.routers ?? []).map((router: RouterConfig, idx: number) => (
                              <div key={idx} className="flex items-center gap-2">
                                <Chip color="primary" variant="flat">{router.prefix}</Chip>
                                <Icon icon="lucide:arrow-right" />
                                <span>{router.server}</span>
                              </div>
                            ))}
                          </div>

                          <div className="space-y-4">
                            <div>
                              <h4 className="text-sm font-semibold mb-2">Enabled Tools:</h4>
                              <div className="flex flex-wrap gap-2">
                                {serverConfig.allowedTools.map((tool: string) => (
                                  <Chip
                                    key={tool}
                                    variant="flat"
                                    color="success"
                                    size="sm"
                                  >
                                    {tool}
                                  </Chip>
                                ))}
                              </div>
                            </div>

                            <div>
                              <h4 className="text-sm font-semibold mb-2">All Tools:</h4>
                              <div className="flex flex-wrap gap-2">
                                {(server.parsedConfig?.tools ?? []).map((tool: ToolConfig) => (
                                  <Chip
                                    key={tool.name}
                                    variant="flat"
                                    color="default"
                                    size="sm"
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
                  theme="vs"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 14,
                    lineNumbers: 'on',
                    roundedSelection: false,
                    scrollBeyondLastLine: false,
                    readOnly: false,
                    automaticLayout: true,
                    'editor.background': '#F8F9FA',
                    'editor.foreground': '#0B0F1A',
                    'editor.lineHighlightBackground': '#F1F3F5',
                    'editor.selectionBackground': '#4C6BCF40',
                    'editor.inactiveSelectionBackground': '#4C6BCF20',
                    'editor.lineHighlightBorder': '#E5E7EB',
                    'editorCursor.foreground': '#4C6BCF',
                    'editorWhitespace.foreground': '#A6B6E8',
                  }}
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
    </div>
  );
}
