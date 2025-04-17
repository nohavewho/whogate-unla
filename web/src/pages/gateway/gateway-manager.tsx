import React from 'react';
import { Card, CardBody, Button, Modal, ModalContent, ModalHeader, ModalBody, ModalFooter, Textarea, useDisclosure } from "@heroui/react";
import { Icon } from '@iconify/react';
import yaml from 'js-yaml';
import { Accordion, AccordionItem, Chip } from "@heroui/react";

interface Gateway {
  id: string;
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
  const [gateways, setGateways] = React.useState<Gateway[]>([
    { id: '1', name: 'Gateway-1', config: 'name: gateway-1\nport: 8080' },
    { id: '2', name: 'Gateway-2', config: 'name: gateway-2\nport: 8081' },
  ]);
  const [currentGateway, setCurrentGateway] = React.useState<Gateway | null>(null);
  const [editConfig, setEditConfig] = React.useState('');
  const [parsedGateways, setParsedGateways] = React.useState<Gateway[]>([]);

  const handleEdit = (gateway: Gateway) => {
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
          g.id === currentGateway.id ? {...g, config: editConfig} : g
        ));
      }
      onOpenChange(false);
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
        >
          Sync Configuration
        </Button>
      </div>
      
      <div className="grid grid-cols-1 gap-4">
        {(parsedGateways || []).map((gateway) => (
          <Card key={gateway.id} className="w-full">
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
                <Accordion>
                  <AccordionItem
                    key="routing"
                    aria-label="Routing Configuration"
                    title="Routing Configuration"
                  >
                    <div className="space-y-2">
                      {(gateway.parsedConfig.routers || []).map((router, idx) => (
                        <div key={idx} className="flex items-center gap-2">
                          <Chip color="primary" variant="flat">{router.prefix}</Chip>
                          <Icon icon="lucide:arrow-right" />
                          <span>{router.server}</span>
                        </div>
                      ))}
                    </div>
                  </AccordionItem>
                  
                  {(gateway.parsedConfig.servers || []).map((server) => (
                    <AccordionItem
                      key={server.name}
                      aria-label={server.name}
                      title={
                        <div className="flex flex-col">
                          <span className="font-semibold">{server.name}</span>
                          <span className="text-sm text-default-500">{server.description}</span>
                        </div>
                      }
                    >
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
                          <h4 className="text-sm font-semibold mb-2">Available Tools:</h4>
                          <div className="flex flex-wrap gap-2">
                            {gateway.parsedConfig?.tools
                              .filter(tool => !server.allowedTools.includes(tool.name))
                              .map((tool) => (
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
                    </AccordionItem>
                  ))}
                </Accordion>
              )}
            </CardBody>
          </Card>
        ))}
      </div>

      <Modal 
        isOpen={isOpen} 
        onOpenChange={onOpenChange}
        size="2xl"
      >
        <ModalContent>
          {(onClose) => (
            <>
              <ModalHeader>Edit Gateway Configuration</ModalHeader>
              <ModalBody>
                <Textarea
                  value={editConfig}
                  onValueChange={setEditConfig}
                  placeholder="Enter YAML configuration"
                  minRows={10}
                  variant="bordered"
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