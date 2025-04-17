import React from 'react';
import { Card, CardBody, Button, Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, useDisclosure } from "@heroui/react";
import { Icon } from '@iconify/react';
import yaml from 'js-yaml';
import { Accordion, AccordionItem, Chip } from "@heroui/react";
import { getMCPList } from '../../services/api';
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

export function GatewayManager() {
  const {isOpen, onOpen, onOpenChange} = useDisclosure();
  const [gateways, setGateways] = React.useState<Gateway[]>([]);
  const [currentGateway, setCurrentGateway] = React.useState<Gateway | null>(null);
  const [editConfig, setEditConfig] = React.useState('');
  const [parsedGateways, setParsedGateways] = React.useState<Gateway[]>([]);
  const [isLoading, setIsLoading] = React.useState(true);
  const [selectedGateway, setSelectedGateway] = React.useState<Gateway | null>(null);

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

  // 获取 yaml 列表
  React.useEffect(() => {
    const fetchYamlList = async () => {
      try {
        setIsLoading(true);
        const yamlList = await getMCPList();
        setGateways(yamlList);
      } catch (error) {
        console.error('Failed to fetch yaml list:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchYamlList();
  }, []);

  const handleEdit = (gateway: Gateway) => {
    console.log('Editing gateway:', gateway);
    setCurrentGateway(gateway);
    setEditConfig(gateway.config);
    onOpen();
  };

  const handleSave = () => {
    try {
      // Validate YAML
      yaml.load(editConfig);

      if (currentGateway) {
        setGateways(gateways.map(g =>
          g.name === currentGateway.name ? {...g, config: editConfig} : g
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
      const parsed = gateways.map(gateway => {
        try {
          const config = yaml.load(gateway.config) as Gateway['parsedConfig'];
          return { ...gateway, parsedConfig: config };
        } catch (e) {
          console.error(`Failed to parse config for ${gateway.name}:`, e);
          return gateway;
        }
      });
      setParsedGateways(parsed);
    };
    parseConfigs();
  }, [gateways]);

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Gateway Manager</h1>
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
          {(parsedGateways || []).map((gateway) => (
            <Card key={gateway.name} className="w-full">
              <CardBody className="flex flex-col gap-4">
                <div className="flex justify-between items-center">
                  <h3 className="text-lg font-semibold">{gateway.name}</h3>
                  <Button
                    isIconOnly
                    color="primary"
                    variant="light"
                    onPress={() => handleEdit(gateway)}
                  >
                    <Icon icon="lucide:edit" className="text-lg" />
                  </Button>
                </div>

                {gateway.parsedConfig && (
                  <div className="space-y-4">
                    {(gateway.parsedConfig.servers || []).map((server) => {
                      const config = gateway.parsedConfig!;
                      return (
                        <div key={server.name} className="space-y-4">
                          <div>
                            <h4 className="text-sm font-semibold">{server.name}</h4>
                            <p className="text-sm text-default-500">{server.description}</p>
                          </div>

                          <div className="space-y-2">
                            <h4 className="text-sm font-semibold">Routing Configuration</h4>
                            {(config.routers || []).map((router, idx) => (
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
                                {server.allowedTools.map((tool) => (
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
                                {config.tools.map((tool) => (
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
              <ModalHeader>Edit Gateway Configuration</ModalHeader>
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
