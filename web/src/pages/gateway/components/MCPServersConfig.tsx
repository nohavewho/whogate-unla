import { Input, Select, SelectItem, Button } from "@heroui/react";
import { useTranslation } from 'react-i18next';

import { GatewayConfig, EnvFormState, KeyValueItem } from '../types';

interface MCPServersConfigProps {
  parsedConfig: GatewayConfig;
  mcpServerFormState: {[serverIndex: number]: {name?: string; url?: string; command?: string; args?: string}};
  envFormState: EnvFormState;
  setMcpServerFormState: (state: {[serverIndex: number]: {name?: string; url?: string; command?: string; args?: string}}) => void;
  updateConfig: (newData: Partial<GatewayConfig>) => void;
  addEnvVariable: (serverIndex: number, key: string, value?: string) => void;
  removeEnvVariable: (serverIndex: number, envIndex: number) => void;
  updateEnvVariable: (serverIndex: number, envIndex: number, updates: Partial<KeyValueItem>) => void;
  newEnvKey: string;
  newEnvValue: string;
  setNewEnvKey: (value: string) => void;
  setNewEnvValue: (value: string) => void;
}

export function MCPServersConfig({
  parsedConfig,
  mcpServerFormState,
  envFormState,
  setMcpServerFormState,
  updateConfig,
  addEnvVariable,
  removeEnvVariable,
  updateEnvVariable,
  newEnvKey,
  newEnvValue,
  setNewEnvKey,
  setNewEnvValue
}: MCPServersConfigProps) {
  const { t } = useTranslation();
  const mcpServers = parsedConfig?.mcpServers || [{ type: "stdio", name: "", command: "", args: [], env: {} }];

  const addServer = () => {
    const newServer = {
      type: "stdio",
      name: "",
      command: "",
      args: [],
      env: {}
    };
    updateConfig({
      mcpServers: [...mcpServers, newServer]
    });
  };

  const removeServer = (index: number) => {
    const updatedServers = mcpServers.filter((_, i) => i !== index);
    updateConfig({
      mcpServers: updatedServers
    });
  };

  return (
    <div className="border-t pt-4 mt-2">
      <h3 className="text-sm font-medium mb-2">{t('gateway.mcp_server_config')}</h3>
      {mcpServers.map((server, index) => (
        <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
          <div className="flex justify-between items-center">
            <div className="flex-1">
              <Input
                label={t('gateway.server_name')}
                value={mcpServerFormState[index]?.name !== undefined ? mcpServerFormState[index].name : (server.name || "")}
                onChange={(e) => {
                  setMcpServerFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      name: e.target.value
                    }
                  }));
                }}
              />
            </div>
            <Button
              color="danger"
              isIconOnly
              className="ml-2"
              onPress={() => removeServer(index)}
            >
              ✕
            </Button>
          </div>
          
          <Select
            label={t('gateway.mcp_type')}
            selectedKeys={[server.type || "stdio"]}
            onChange={(e) => updateConfig({
              mcpServers: mcpServers.map((s, i) => 
                i === index ? { ...s, type: e.target.value } : s
              )
            })}
            aria-label={t('gateway.mcp_type')}
          >
            <SelectItem key="stdio">stdio</SelectItem>
            <SelectItem key="sse">sse</SelectItem>
            <SelectItem key="streamable-http">streamable-http</SelectItem>
          </Select>
          
          {(server.type === 'stdio' || !server.type) && (
            <>
              <Input
                label={t('gateway.command')}
                value={mcpServerFormState[index]?.command !== undefined ? mcpServerFormState[index].command : (server.command || '')}
                onChange={(e) => {
                  setMcpServerFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      command: e.target.value
                    }
                  }));
                }}
              />
              
              <Input
                label={t('gateway.args')}
                value={mcpServerFormState[index]?.args !== undefined ? mcpServerFormState[index].args : (server.args?.join(' ') || '')}
                onChange={(e) => {
                  setMcpServerFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      args: e.target.value
                    }
                  }));
                }}
                placeholder="arg1 arg2 arg3"
              />
              
              <div className="mt-2">
                <h4 className="text-sm font-medium mb-2">{t('gateway.env_variables')}</h4>
                {Object.keys(server.env || {}).map((key, envIndex) => (
                  <div key={envIndex} className="flex items-center gap-2 mb-2">
                    <Input
                      className="flex-1"
                      value={(envFormState[index]?.[envIndex]?.key !== undefined) 
                             ? envFormState[index][envIndex].key 
                             : key}
                      onChange={(e) => {
                        updateEnvVariable(index, envIndex, {
                          key: e.target.value
                        });
                      }}
                    />
                    <Input
                      className="flex-1"
                      value={(envFormState[index]?.[envIndex]?.value !== undefined)
                             ? envFormState[index][envIndex].value
                             : (server.env?.[key] || "")}
                      onChange={(e) => {
                        updateEnvVariable(index, envIndex, {
                          value: e.target.value
                        });
                      }}
                    />
                    <Button
                      color="danger"
                      isIconOnly
                      size="sm"
                      onPress={() => removeEnvVariable(index, envIndex)}
                    >
                      ✕
                    </Button>
                  </div>
                ))}
                
                <div className="flex items-center gap-2 mt-2">
                  <Input
                    className="flex-1"
                    placeholder="KEY"
                    value={newEnvKey}
                    onChange={(e) => setNewEnvKey(e.target.value)}
                  />
                  <Input
                    className="flex-1"
                    placeholder="VALUE"
                    value={newEnvValue}
                    onChange={(e) => setNewEnvValue(e.target.value)}
                  />
                  <Button
                    color="primary"
                    size="sm"
                    onPress={() => addEnvVariable(index, newEnvKey, newEnvValue)}
                  >
                    +
                  </Button>
                </div>
              </div>
            </>
          )}
          
          {(server.type === 'sse' || server.type === 'streamable-http') && (
            <Input
              label={t('gateway.url')}
              value={mcpServerFormState[index]?.url !== undefined ? mcpServerFormState[index].url : (server.url || '')}
              onChange={(e) => {
                setMcpServerFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    url: e.target.value
                  }
                }));
              }}
            />
          )}
        </div>
      ))}
      <Button
        size="sm"
        color="primary"
        onPress={addServer}
        className="w-full"
      >
        {t('gateway.add_mcp_server')}
      </Button>
    </div>
  );
} 