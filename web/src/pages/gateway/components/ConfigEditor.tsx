import { Input, Select, SelectItem, Radio, RadioGroup, Button, Tabs, Tab } from "@heroui/react";
import Editor from '@monaco-editor/react';
import yaml from 'js-yaml';
import { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';

import { getTenants } from '../../../services/api';
import { getRandomLetters } from "../../../utils/utils";
import { defaultConfig, defaultMCPConfig } from '../constants/defaultConfig';
import { ConfigEditorProps, GatewayConfig, HeadersFormState, EnvFormState, KeyValueItem, Tenant } from '../types';

import { MCPServersConfig } from './MCPServersConfig';
import { RouterConfig } from './RouterConfig';
import { ServersConfig } from './ServersConfig';
import { ToolsConfig } from './ToolsConfig';


export function ConfigEditor({ config, onChange, isDark, editorOptions, isEditing }: ConfigEditorProps) {
  const { t } = useTranslation();
  const [isYamlMode, setIsYamlMode] = useState<boolean>(false);
  const [parsedConfig, setParsedConfig] = useState<GatewayConfig | null>(null);
  const [envKeys, setEnvKeys] = useState<string[]>([]);
  const [newEnvKey, setNewEnvKey] = useState<string>("");
  const [newEnvValue, setNewEnvValue] = useState<string>("");
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [isLoadingTenants, setIsLoadingTenants] = useState<boolean>(false);
  
  // 表单状态
  const [toolFormState, setToolFormState] = useState<{[toolIndex: number]: {[field: string]: string}}>({});
  const [generalFormState, setGeneralFormState] = useState<{name?: string; tenant?: string}>({});
  const [routerFormState, setRouterFormState] = useState<{[routerIndex: number]: {prefix?: string; server?: string}}>({});
  const [serverFormState, setServerFormState] = useState<{[serverIndex: number]: {name?: string; description?: string}}>({});
  const [headerFormState, setHeaderFormState] = useState<HeadersFormState>({});
  const [mcpServerFormState, setMcpServerFormState] = useState<{[serverIndex: number]: {name?: string; url?: string; command?: string; args?: string}}>({});
  const [envFormState, setEnvFormState] = useState<EnvFormState>({});

  // CORS相关状态
  const [selectedMethod, setSelectedMethod] = useState<{[routerIndex: number]: string}>({});
  const [newOrigin, setNewOrigin] = useState<{[routerIndex: number]: string}>({});
  const [newExposeHeader, setNewExposeHeader] = useState<{[routerIndex: number]: string}>({});
  const [newHeader, setNewHeader] = useState<{[routerIndex: number]: string}>({});

  // 工具函数
  const addHeader = (toolIndex: number, key: string, value: string = "") => {
    setHeaderFormState(prev => {
      const newState = { ...prev } as HeadersFormState;
      newState[toolIndex] = [...(prev[toolIndex] || []), { key, value }];
      return newState;
    });
  };

  const removeHeader = (toolIndex: number, headerIndex: number) => {
    setHeaderFormState(prev => {
      const newState = { ...prev } as HeadersFormState;
      if (newState[toolIndex]) {
        newState[toolIndex] = newState[toolIndex].filter((_, i) => i !== headerIndex);
      }
      return newState;
    });
  };

  const updateHeader = (toolIndex: number, headerIndex: number, updates: Partial<KeyValueItem>) => {
    setHeaderFormState(prev => {
      const newState = { ...prev } as HeadersFormState;
      if (newState[toolIndex]) {
        newState[toolIndex] = newState[toolIndex].map((item, i) =>
          i === headerIndex ? { ...item, ...updates } : item
        );
      }
      return newState;
    });
  };

  const addEnvVariable = (serverIndex: number, key: string, value: string = "") => {
    setEnvFormState(prev => {
      const newState = { ...prev } as EnvFormState;
      newState[serverIndex] = [...(prev[serverIndex] || []), { key, value }];
      return newState;
    });
  };

  const removeEnvVariable = (serverIndex: number, envIndex: number) => {
    setEnvFormState(prev => {
      const newState = { ...prev } as EnvFormState;
      if (newState[serverIndex]) {
        newState[serverIndex] = newState[serverIndex].filter((_, i) => i !== envIndex);
      }
      return newState;
    });
  };

  const updateEnvVariable = (serverIndex: number, envIndex: number, updates: Partial<KeyValueItem>) => {
    setEnvFormState(prev => {
      const newState = { ...prev } as EnvFormState;
      if (newState[serverIndex]) {
        newState[serverIndex] = newState[serverIndex].map((item, i) =>
          i === envIndex ? { ...item, ...updates } : item
        );
      }
      return newState;
    });
  };

  // 使用useCallback包装updateConfig函数
  const updateConfig = useCallback((newData: Partial<GatewayConfig>) => {
    if (!parsedConfig) {
      const baseConfig = defaultConfig;
      const updated = {
        ...baseConfig,
        ...newData
      };
      setParsedConfig(updated);
      try {
        const yamlString = yaml.dump(updated);
        onChange(yamlString);
      } catch (e) {
        console.error("Failed to generate YAML:", e);
      }
      return;
    }
    
    const updated = {
      ...parsedConfig,
      ...newData
    };
    
    if (isYamlMode && isEditing && parsedConfig.name && parsedConfig.name.trim() !== '') {
      updated.name = parsedConfig.name;
    }
    
    try {
      const newYaml = yaml.dump(updated);
      onChange(newYaml);
    } catch (e) {
      console.error("Failed to generate YAML:", e);
    }
  }, [parsedConfig, isYamlMode, isEditing, onChange]);

  const renderServerOptions = () => {
    if (!parsedConfig) return [<SelectItem key="default">{t('common.name')}</SelectItem>];
    
    if (parsedConfig.servers && parsedConfig.servers.length > 0) {
      return parsedConfig.servers.map(server => (
        <SelectItem key={server.name || "default"}>
          {server.name || t('common.name')}
        </SelectItem>
      ));
    } else if (parsedConfig.mcpServers && parsedConfig.mcpServers.length > 0) {
      return parsedConfig.mcpServers.map(server => (
        <SelectItem key={server.name || "default"}>
          {server.name || t('common.name')}
        </SelectItem>
      ));
    }
    
    return [<SelectItem key="default">{t('common.name')}</SelectItem>];
  };

  // 加载租户列表
  useEffect(() => {
    const fetchTenants = async () => {
      setIsLoadingTenants(true);
      try {
        const tenantsData = await getTenants();
        setTenants(tenantsData);
      } catch (error) {
        console.error("Failed to fetch tenants:", error);
      } finally {
        setIsLoadingTenants(false);
      }
    };

    fetchTenants();
  }, []);

  // 初始化路由表单状态
  useEffect(() => {
    if (parsedConfig?.routers) {
      const initialFormState: Record<number, any> = parsedConfig.routers.reduce((acc: Record<number, any>, router, idx) => {
        acc[idx] = {
          ...router,
          prefix: router.prefix || ""
        };
        return acc;
      }, {});
      
      setRouterFormState(initialFormState);
    }
  }, [parsedConfig?.routers]);

  // 初始化MCP服务器表单状态
  useEffect(() => {
    if (parsedConfig?.mcpServers) {
      const initialFormState: Record<number, any> = parsedConfig.mcpServers.reduce((acc: Record<number, any>, server, idx) => {
        acc[idx] = {
          name: server.name || "",
          url: server.url || "",
          command: server.command || "",
          args: server.args?.join(' ') || ""
        };
        return acc;
      }, {});
      
      setMcpServerFormState(initialFormState);
    }
  }, [parsedConfig?.mcpServers]);

  // 初始化环境变量表单状态
  useEffect(() => {
    if (parsedConfig?.mcpServers && parsedConfig.mcpServers.length > 0 && parsedConfig.mcpServers[0]?.env) {
      const initialEnvState: EnvFormState = {
        0: []
      };
      
      Object.entries(parsedConfig.mcpServers[0].env).forEach(([key, value]) => {
        initialEnvState[0].push({ key, value: String(value) });
      });
      
      setEnvFormState(initialEnvState);
      setEnvKeys(Object.keys(parsedConfig.mcpServers[0].env));
    }
  }, [parsedConfig?.mcpServers]);

  // 初始化Headers表单状态
  useEffect(() => {
    if (parsedConfig?.tools) {
      const initialHeaderState: HeadersFormState = {};
      
      parsedConfig.tools.forEach((tool, toolIndex) => {
        if (tool.headers) {
          initialHeaderState[toolIndex] = Object.entries(tool.headers).map(([key, value]) => ({
            key,
            value: String(value)
          }));
        }
      });
      
      setHeaderFormState(initialHeaderState);
    }
  }, [parsedConfig?.tools]);

  // 监听表单状态变化并更新配置
  useEffect(() => {
    if (!parsedConfig) return;

    // 处理工具表单状态
    if (Object.keys(toolFormState).length > 0) {
      const updatedTools = parsedConfig.tools ? [...parsedConfig.tools] : [];
      let hasChanges = false;
      
      Object.entries(toolFormState).forEach(([toolIndexStr, toolData]) => {
        const toolIndex = parseInt(toolIndexStr);
        if (isNaN(toolIndex) || !updatedTools[toolIndex]) return;
        
        const updatedTool = { ...updatedTools[toolIndex] };
        let toolChanged = false;
        
        Object.entries(toolData).forEach(([field, value]) => {
          if (field === 'name' && updatedTool.name !== value) {
            updatedTool.name = value;
            toolChanged = true;
          } else if (field === 'description' && updatedTool.description !== value) {
            updatedTool.description = value;
            toolChanged = true;
          } else if (field === 'endpoint' && updatedTool.endpoint !== value) {
            updatedTool.endpoint = value;
            toolChanged = true;
          } else if (field === 'method' && updatedTool.method !== value) {
            updatedTool.method = value;
            toolChanged = true;
          } else if (field === 'requestBody' && updatedTool.requestBody !== value) {
            updatedTool.requestBody = value;
            toolChanged = true;
          } else if (field === 'responseBody' && updatedTool.responseBody !== value) {
            updatedTool.responseBody = value;
            toolChanged = true;
          }
        });
        
        if (toolChanged) {
          updatedTools[toolIndex] = updatedTool;
          hasChanges = true;
        }
      });
      
      if (hasChanges) {
        updateConfig({ tools: updatedTools });
      }
    }

    // 处理一般配置表单状态
    if (Object.keys(generalFormState).length > 0) {
      let hasChanges = false;
      const updates: Partial<GatewayConfig> = {};
      
      if (generalFormState.name !== undefined && 
          parsedConfig.name !== generalFormState.name && 
          (!isEditing || !parsedConfig.name || parsedConfig.name.trim() === '')) {
        updates.name = generalFormState.name;
        hasChanges = true;
      }
      
      if (generalFormState.tenant !== undefined && parsedConfig.tenant !== generalFormState.tenant) {
        updates.tenant = generalFormState.tenant;
        hasChanges = true;
        
        if (parsedConfig.routers && parsedConfig.routers.length > 0) {
          const selectedTenant = tenants.find(t => t.name === generalFormState.tenant);
          if (selectedTenant) {
            const updatedRouters = [...parsedConfig.routers];
            updatedRouters.forEach((router, index) => {
              updatedRouters[index] = {
                ...router,
                prefix: `${selectedTenant.prefix}/${router.prefix})}`
              };
            });
            updates.routers = updatedRouters;
          }
        }
      }
      
      if (hasChanges) {
        updateConfig(updates);
      }
    }

    // 处理路由配置表单状态
    if (Object.keys(routerFormState).length > 0) {
      const updatedRouters = parsedConfig.routers ? [...parsedConfig.routers] : [];
      let hasChanges = false;
      
      Object.entries(routerFormState).forEach(([routerIndexStr, routerData]) => {
        const routerIndex = parseInt(routerIndexStr);
        if (isNaN(routerIndex) || !updatedRouters[routerIndex]) return;
        
        const updatedRouter = { ...updatedRouters[routerIndex] };
        let routerChanged = false;
        
        if (routerData.prefix !== undefined && updatedRouter.prefix !== routerData.prefix) {
          updatedRouter.prefix = routerData.prefix;
          routerChanged = true;
        }
        
        if (routerData.server !== undefined && updatedRouter.server !== routerData.server) {
          updatedRouter.server = routerData.server;
          routerChanged = true;
        }
        
        if (routerChanged) {
          updatedRouters[routerIndex] = updatedRouter;
          hasChanges = true;
        }
      });
      
      if (hasChanges) {
        updateConfig({ routers: updatedRouters });
      }
    }

    // 处理服务器配置表单状态
    if (Object.keys(serverFormState).length > 0 && parsedConfig.servers) {
      const updatedServers = [...parsedConfig.servers];
      let hasChanges = false;
      
      Object.entries(serverFormState).forEach(([serverIndexStr, serverData]) => {
        const serverIndex = parseInt(serverIndexStr);
        if (isNaN(serverIndex) || !updatedServers[serverIndex]) return;
        
        const updatedServer = { ...updatedServers[serverIndex] };
        let serverChanged = false;
        
        if (serverData.name !== undefined && updatedServer.name !== serverData.name) {
          updatedServer.name = serverData.name;
          serverChanged = true;
        }
        
        if (serverData.description !== undefined && updatedServer.description !== serverData.description) {
          updatedServer.description = serverData.description;
          serverChanged = true;
        }
        
        if (serverChanged) {
          updatedServers[serverIndex] = updatedServer;
          hasChanges = true;
        }
      });
      
      if (hasChanges) {
        updateConfig({ servers: updatedServers });
      }
    }

    // 处理Headers配置表单状态
    if (Object.keys(headerFormState).length > 0) {
      const updatedTools = parsedConfig.tools ? [...parsedConfig.tools] : [];
      let hasChanges = false;
      
      Object.entries(headerFormState).forEach(([toolIndexStr, headersArray]) => {
        const toolIndex = parseInt(toolIndexStr);
        if (isNaN(toolIndex) || !updatedTools[toolIndex]) return;
        
        const updatedTool = { ...updatedTools[toolIndex] };
        const headersObject: Record<string, string> = {};
        const headersOrder: string[] = [];
        
        headersArray.forEach(({ key, value }: KeyValueItem) => {
          if (key) {
            headersObject[key] = value;
            headersOrder.push(key);
          }
        });
        
        const currentHeaders = updatedTool.headers || {};
        const currentOrder = updatedTool.headersOrder || Object.keys(currentHeaders);
        
        const headersChanged = JSON.stringify(headersObject) !== JSON.stringify(currentHeaders) ||
                             JSON.stringify(headersOrder) !== JSON.stringify(currentOrder);
        
        if (headersChanged) {
          updatedTool.headers = headersObject;
          updatedTool.headersOrder = headersOrder;
          updatedTools[toolIndex] = updatedTool;
          hasChanges = true;
        }
      });
      
      if (hasChanges) {
        updateConfig({ tools: updatedTools });
      }
    }

    // 处理环境变量表单状态
    if (Object.keys(envFormState).length > 0 && parsedConfig.mcpServers) {
      const updatedServers = [...parsedConfig.mcpServers];
      let hasChanges = false;
      
      Object.entries(envFormState).forEach(([serverIndexStr, envArray]) => {
        const serverIndex = parseInt(serverIndexStr);
        if (isNaN(serverIndex) || !updatedServers[serverIndex]) return;
        
        const updatedServer = { ...updatedServers[serverIndex] };
        const envObject: Record<string, string> = {};
        
        envArray.forEach(({ key, value }: KeyValueItem) => {
          if (key) {
            envObject[key] = value;
          }
        });
        
        const currentEnv = updatedServer.env || {};
        const envChanged = JSON.stringify(envObject) !== JSON.stringify(currentEnv);
        
        if (envChanged) {
          updatedServer.env = envObject;
          updatedServers[serverIndex] = updatedServer;
          hasChanges = true;
        }
      });
      
      if (hasChanges) {
        updateConfig({ mcpServers: updatedServers });
      }
    }
  }, [toolFormState, generalFormState, routerFormState, serverFormState, headerFormState, envFormState, parsedConfig, updateConfig, tenants, isEditing]);

  useEffect(() => {
    try {
      if (!config || config.trim() === '') {
        setParsedConfig(defaultConfig);
        return;
      }

      const parsed = yaml.load(config) as GatewayConfig;
      setParsedConfig(parsed);
      
      if (parsed.mcpServers && parsed.mcpServers.length > 0 && parsed.mcpServers[0]?.env) {
        setEnvKeys(Object.keys(parsed.mcpServers[0].env));
      }
    } catch (e) {
      console.error("Failed to parse config:", e);
      setParsedConfig(defaultConfig);
    }
  }, [config]);

  return (
    <div className="h-full flex flex-col">
      <div className="flex justify-end mb-4">
        <Button
          color={isYamlMode ? "primary" : "default"}
          variant={isYamlMode ? "solid" : "flat"}
          onPress={() => setIsYamlMode(true)}
          className="mr-2"
          size="sm"
        >
          {t('gateway.yaml_mode')}
        </Button>
        <Button
          color={!isYamlMode ? "primary" : "default"}
          variant={!isYamlMode ? "solid" : "flat"}
          onPress={() => setIsYamlMode(false)}
          size="sm"
        >
          {t('gateway.form_mode')}
        </Button>
      </div>

      {isYamlMode ? (
        <Editor
          height="100%"
          defaultLanguage="yaml"
          value={config}
          onChange={(value) => {
            if (value !== undefined) {
              onChange(value);
            }
          }}
          theme={isDark ? "vs-dark" : "light"}
          options={editorOptions}
        />
      ) : (
        <div className="space-y-4">
          <div className="space-y-2">
            <Input
              label={t('gateway.name')}
              value={generalFormState.name !== undefined ? generalFormState.name : (parsedConfig?.name || "")}
              onChange={(e) => {
                setGeneralFormState(prev => ({
                  ...prev,
                  name: e.target.value
                }));
              }}
              isDisabled={Boolean(isEditing && parsedConfig?.name && parsedConfig.name.trim() !== '')}
              description={(isEditing && parsedConfig?.name && parsedConfig.name.trim() !== '') ? t('gateway.name_locked') : undefined}
            />
            
            <Select
              label={t('gateway.tenant')}
              selectedKeys={generalFormState.tenant !== undefined ? [generalFormState.tenant] : (parsedConfig?.tenant ? [parsedConfig.tenant] : ['default'])}
              onChange={(e) => {
                setGeneralFormState(prev => ({
                  ...prev,
                  tenant: e.target.value
                }));
              }}
              aria-label={t('gateway.tenant')}
              isLoading={isLoadingTenants}
            >
              {tenants.length > 0 ? (
                tenants.filter(tenant => tenant.isActive).map(tenant => (
                  <SelectItem key={tenant.name} textValue={tenant.name}>
                    {tenant.name}
                    {tenant.prefix && <span className="text-tiny text-default-400"> ({tenant.prefix})</span>}
                  </SelectItem>
                ))
              ) : (
                <SelectItem key="default">default</SelectItem>
              )}
            </Select>
            
            <div className="mt-2">
              <h3 className="text-sm font-medium mb-2">{t('gateway.created_at')}: {new Date().toLocaleString()}</h3>
              <h3 className="text-sm font-medium mb-2">{t('gateway.updated_at')}: {new Date().toLocaleString()}</h3>
            </div>
          </div>

          <Tabs aria-label="Configuration sections" className="w-full" disableAnimation>
            <Tab key="http-servers" title={t('gateway.http_servers')}>
              <ServersConfig
                parsedConfig={parsedConfig || defaultConfig}
                serverFormState={serverFormState}
                setServerFormState={setServerFormState}
                updateConfig={updateConfig}
              />
            </Tab>
            <Tab key="mcp-servers" title={t('gateway.mcp_servers')}>
              <MCPServersConfig
                parsedConfig={parsedConfig || defaultMCPConfig}
                mcpServerFormState={mcpServerFormState}
                envFormState={envFormState}
                setMcpServerFormState={setMcpServerFormState}
                updateConfig={updateConfig}
                addEnvVariable={addEnvVariable}
                removeEnvVariable={removeEnvVariable}
                updateEnvVariable={updateEnvVariable}
                newEnvKey={newEnvKey}
                newEnvValue={newEnvValue}
                setNewEnvKey={setNewEnvKey}
                setNewEnvValue={setNewEnvValue}
              />
            </Tab>
            <Tab key="tools" title={t('gateway.tools')}>
              <ToolsConfig
                parsedConfig={parsedConfig || defaultConfig}
                toolFormState={toolFormState}
                headerFormState={headerFormState}
                setToolFormState={setToolFormState}
                updateConfig={updateConfig}
                addHeader={addHeader}
                removeHeader={removeHeader}
                updateHeader={updateHeader}
              />
            </Tab>
            <Tab key="routers" title={t('gateway.routers')}>
              <RouterConfig
                parsedConfig={parsedConfig || defaultConfig}
                routerFormState={routerFormState}
                setRouterFormState={setRouterFormState}
                updateConfig={updateConfig}
                tenants={tenants}
                selectedMethod={selectedMethod}
                setSelectedMethod={setSelectedMethod}
                newOrigin={newOrigin}
                setNewOrigin={setNewOrigin}
                newExposeHeader={newExposeHeader}
                setNewExposeHeader={setNewExposeHeader}
                newHeader={newHeader}
                setNewHeader={setNewHeader}
                renderServerOptions={renderServerOptions}
              />
            </Tab>
          </Tabs>
        </div>
      )}
    </div>
  );
}
