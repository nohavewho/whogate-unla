import { Input, Select, SelectItem, Radio, RadioGroup, Chip, Button, Switch } from "@heroui/react";
import Editor from '@monaco-editor/react';
import yaml from 'js-yaml';
import  { useState, useEffect, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { getTenants } from '../../../services/api';
import { getRandomLetters } from "../../../utils/utils";

interface Tenant {
  id: number;
  name: string;
  prefix: string;
  description: string;
  isActive: boolean;
}

interface ConfigEditorProps {
  config: string;
  onChange: (newConfig: string) => void;
  isDark: boolean;
  editorOptions: Record<string, unknown>;
  isEditing?: boolean;
}

interface GatewayConfig {
  name: string;
  tenant: string;
  routers?: Array<{
    server: string;
    prefix: string;
    cors?: Record<string, unknown>;
  }>;
  servers?: Array<{
    name: string;
    description: string;
    allowedTools: string[];
    config?: Record<string, unknown>;
  }>;
  tools?: Array<{
    name: string;
    description: string;
    method: string;
    endpoint: string;
    args?: Array<{
      name: string;
      position: string;
      required: boolean;
      type: string;
      description: string;
      default: string;
    }>;
    requestBody?: string;
    responseBody?: string;
    headers?: Record<string, string>;
    headersOrder?: string[];
  }>;
  mcpServers?: Array<{
    type: string;
    name: string;
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
  }>;
}

interface CorsConfig {
  allowOrigins?: string[];
  allowMethods?: string[];
  allowHeaders?: string[];
  exposeHeaders?: string[];
  allowCredentials?: boolean;
}

// 默认配置对象，用于新建或配置为空的情况
const defaultConfig: GatewayConfig = {
  name: "",
  tenant: "default",
  routers: [{
    server: "",
    prefix: "/" + getRandomLetters(4)
  }],
  servers: [{
    name: "",
    description: "",
    allowedTools: []
  }],
  tools: []
};

// 默认MCP配置对象
const defaultMCPConfig: GatewayConfig = {
  name: "",
  tenant: "default",
  routers: [{
    server: "",
    prefix: "/mcp"
  }],
  mcpServers: [{
    type: "stdio",
    name: "",
    command: "",
    args: [],
    env: {}
  }]
};

// 防抖函数
function useDebounce<T>(value: T, delay: number): [T, (value: T) => void] {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);
  const [immediateValue, setImmediateValue] = useState<T>(value);
  
  useEffect(() => {
    const handler = window.setTimeout(() => {
      setDebouncedValue(immediateValue);
    }, delay);
    
    return () => {
      window.clearTimeout(handler);
    };
  }, [immediateValue, delay]);
  
  return [debouncedValue, setImmediateValue];
}

export function ConfigEditor({ config, onChange, isDark, editorOptions, isEditing }: ConfigEditorProps) {
  const { t } = useTranslation();
  const [isYamlMode, setIsYamlMode] = useState<boolean>(false);
  const [parsedConfig, setParsedConfig] = useState<GatewayConfig | null>(null);
  const [proxyType, setProxyType] = useState<string>("http");
  const [envKeys, setEnvKeys] = useState<string[]>([]);
  const [newEnvKey, setNewEnvKey] = useState<string>("");
  const [newEnvValue, setNewEnvValue] = useState<string>("");
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [isLoadingTenants, setIsLoadingTenants] = useState<boolean>(false);
  
  // 表单状态 - 使用防抖来避免频繁更新导致的光标重置
  const [toolFormState, setToolFormState] = useState<{[toolIndex: number]: {[field: string]: string}}>({});
  const [debouncedToolForm, updateToolForm] = useDebounce(toolFormState, 500);
  
  // 通用配置表单状态
  const [generalFormState, setGeneralFormState] = useState<{name?: string; tenant?: string}>({});
  const [debouncedGeneralForm, updateGeneralForm] = useDebounce(generalFormState, 500);
  
  // 路由配置表单状态
  const [routerFormState, setRouterFormState] = useState<{[routerIndex: number]: {prefix?: string; server?: string}}>({});
  const [debouncedRouterForm, updateRouterForm] = useDebounce(routerFormState, 500);
  
  // 服务器配置表单状态
  const [serverFormState, setServerFormState] = useState<{[serverIndex: number]: {name?: string; description?: string}}>({});
  const [debouncedServerForm, updateServerForm] = useDebounce(serverFormState, 500);
  
  // Headers配置表单状态
  const [headerFormState, setHeaderFormState] = useState<{[toolIndex: number]: {[keyIndex: string]: {key?: string; value?: string}}}>({});
  const [debouncedHeaderForm, updateHeaderForm] = useDebounce(headerFormState, 500);

  // 添加一个新状态来跟踪选中的HTTP方法
  const [selectedMethod, setSelectedMethod] = useState<{[routerIndex: number]: string}>({});

  // 添加新的状态来跟踪输入值
  const [newOrigin, setNewOrigin] = useState<{[routerIndex: number]: string}>({});
  const [newHeader, setNewHeader] = useState<{[routerIndex: number]: string}>({});
  const [newExposeHeader, setNewExposeHeader] = useState<{[routerIndex: number]: string}>({});

  // 使用useCallback包装updateConfig函数
  const updateConfig = useCallback((newData: Partial<GatewayConfig>) => {
    if (!parsedConfig) {
      // 如果parsedConfig为null，使用默认配置
      const baseConfig = proxyType === "mcp" ? defaultMCPConfig : defaultConfig;
      const updated = {
        ...baseConfig,
        ...newData
      };
      
      try {
        const newYaml = yaml.dump(updated);
        onChange(newYaml);
        setParsedConfig(updated);
      } catch (e) {
        console.error("Failed to generate YAML:", e);
      }
      return;
    }
    
    // 创建更新的配置
    const updated = {
      ...parsedConfig,
      ...newData
    };
    
    // 如果是YAML模式直接修改，且在编辑模式下，确保保留原始name
    if (isYamlMode && isEditing && parsedConfig.name && parsedConfig.name.trim() !== '') {
      updated.name = parsedConfig.name;
    }
    
    try {
      const newYaml = yaml.dump(updated);
      onChange(newYaml);
    } catch (e) {
      console.error("Failed to generate YAML:", e);
    }
  }, [parsedConfig, proxyType, isYamlMode, isEditing, onChange]);

  useEffect(() => {
    try {
      if (!config || config.trim() === '') {
        setParsedConfig(defaultConfig);
        setProxyType("http");
        return;
      }

      const parsed = yaml.load(config) as GatewayConfig;
      setParsedConfig(parsed);
      
      // 根据配置判断是 MCP 还是 HTTP 代理
      if (parsed.mcpServers && parsed.mcpServers.length > 0) {
        setProxyType("mcp");
      } else if (parsed.servers && parsed.servers.length > 0) {
        setProxyType("http");
      }
      
      // 提取环境变量
      if (parsed.mcpServers && parsed.mcpServers.length > 0 && parsed.mcpServers[0]?.env) {
        setEnvKeys(Object.keys(parsed.mcpServers[0].env));
      }
    } catch (e) {
      console.error("Failed to parse config:", e);
      // 解析失败时使用默认配置
      setParsedConfig(defaultConfig);
      setProxyType("http");
    }
  }, [config]);

  // 当表单状态发生防抖变化时，应用更改到实际配置
  useEffect(() => {
    if (!parsedConfig || Object.keys(debouncedToolForm).length === 0) return;
    
    const updatedTools = parsedConfig.tools ? [...parsedConfig.tools] : [];
    let hasChanges = false;
    
    // 遍历所有工具表单状态
    Object.entries(debouncedToolForm).forEach(([toolIndexStr, toolData]) => {
      const toolIndex = parseInt(toolIndexStr);
      if (isNaN(toolIndex) || !updatedTools[toolIndex]) return;
      
      const updatedTool = { ...updatedTools[toolIndex] };
      let toolChanged = false;
      
      // 应用各字段的更改
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
        } else if (field.startsWith('header_key_') || field.startsWith('header_value_')) {
          // 处理Header的变更在单独的位置处理
        }
      });
      
      if (toolChanged) {
        updatedTools[toolIndex] = updatedTool;
        hasChanges = true;
      }
    });
    
    // 如果有变更，更新配置
    if (hasChanges) {
      updateConfig({ tools: updatedTools });
    }
  }, [debouncedToolForm, parsedConfig, updateConfig]);

  // 处理一般配置的防抖更新
  useEffect(() => {
    if (!parsedConfig || Object.keys(debouncedGeneralForm).length === 0) return;
    
    let hasChanges = false;
    const updates: Partial<GatewayConfig> = {};
    
    // 只有在创建模式或name为空时才允许修改name
    if (debouncedGeneralForm.name !== undefined && 
        parsedConfig.name !== debouncedGeneralForm.name && 
        (!isEditing || !parsedConfig.name || parsedConfig.name.trim() === '')) {
      updates.name = debouncedGeneralForm.name;
      hasChanges = true;
    }
    
    if (debouncedGeneralForm.tenant !== undefined && parsedConfig.tenant !== debouncedGeneralForm.tenant) {
      updates.tenant = debouncedGeneralForm.tenant;
      hasChanges = true;
      
      // 更新路由前缀，根据选中的租户
      if (parsedConfig.routers && parsedConfig.routers.length > 0) {
        const selectedTenant = tenants.find(t => t.name === debouncedGeneralForm.tenant);
        if (selectedTenant) {
          const updatedRouters = [...parsedConfig.routers];
          // 更新每个路由的前缀
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
  }, [debouncedGeneralForm, parsedConfig, isEditing, updateConfig, tenants, proxyType]);
  
  // 处理路由配置的防抖更新
  useEffect(() => {
    if (!parsedConfig || Object.keys(debouncedRouterForm).length === 0) return;
    
    const updatedRouters = parsedConfig.routers ? [...parsedConfig.routers] : [];
    let hasChanges = false;
    
    Object.entries(debouncedRouterForm).forEach(([routerIndexStr, routerData]) => {
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
  }, [debouncedRouterForm, parsedConfig, updateConfig]);
  
  // 处理服务器配置的防抖更新
  useEffect(() => {
    if (!parsedConfig || Object.keys(debouncedServerForm).length === 0 || !parsedConfig.servers) return;
    
    const updatedServers = [...parsedConfig.servers];
    let hasChanges = false;
    
    Object.entries(debouncedServerForm).forEach(([serverIndexStr, serverData]) => {
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
  }, [debouncedServerForm, parsedConfig, updateConfig]);

  // 处理Headers配置的防抖更新
  useEffect(() => {
    if (!parsedConfig || Object.keys(debouncedHeaderForm).length === 0) return;
    
    const updatedTools = parsedConfig.tools ? [...parsedConfig.tools] : [];
    let hasChanges = false;
    
    Object.entries(debouncedHeaderForm).forEach(([toolIndexStr, headerData]) => {
      const toolIndex = parseInt(toolIndexStr);
      if (isNaN(toolIndex) || !updatedTools[toolIndex]) return;
      
      const updatedTool = { ...updatedTools[toolIndex] };
      const updatedHeaders = { ...(updatedTool.headers || {}) };
      const updatedHeadersOrder = [...(updatedTool.headersOrder || Object.keys(updatedHeaders))];
      let headerChanged = false;
      
      Object.entries(headerData).forEach(([headerKey, data]) => {
        // 键格式为 "headerIndex_originalKey"
        const parts = headerKey.split('_');
        const headerIndex = parseInt(parts[0]);
        // 原始Key是从第二部分开始到最后，以防key中包含下划线
        const originalKey = parts.slice(1).join('_');
        
        if (isNaN(headerIndex) || !originalKey) return;
        
        if (data.key !== undefined && originalKey !== data.key) {
          // 键名发生变化
          const value = data.value !== undefined ? data.value : (updatedHeaders[originalKey] || '');
          delete updatedHeaders[originalKey];
          updatedHeaders[data.key] = value;
          
          // 更新顺序
          const orderIndex = updatedHeadersOrder.findIndex(k => k === originalKey);
          if (orderIndex !== -1) {
            updatedHeadersOrder[orderIndex] = data.key;
          }
          
          headerChanged = true;
        } else if (data.value !== undefined) {
          // 只有值发生变化
          updatedHeaders[originalKey] = data.value;
          headerChanged = true;
        }
      });
      
      if (headerChanged) {
        updatedTool.headers = updatedHeaders;
        updatedTool.headersOrder = updatedHeadersOrder;
        updatedTools[toolIndex] = updatedTool;
        hasChanges = true;
      }
    });
    
    if (hasChanges) {
      updateConfig({ tools: updatedTools });
    }
  }, [debouncedHeaderForm, parsedConfig, updateConfig]);

  const updateRouter = (index: number, data: Partial<{ server: string; prefix: string; cors?: Record<string, unknown> }>) => {
    if (!parsedConfig) return;
    
    const updatedRouters = parsedConfig.routers ? [...parsedConfig.routers] : [];
    if (!updatedRouters[index]) {
      updatedRouters[index] = { server: "", prefix: "" };
    }
    updatedRouters[index] = { ...updatedRouters[index], ...data };
    
    updateConfig({ routers: updatedRouters });
  };

  const updateServer = (index: number, data: Partial<{ name: string; description: string; allowedTools: string[]; config?: Record<string, unknown> }>) => {
    if (!parsedConfig) return;
    
    const updatedServers = parsedConfig.servers ? [...parsedConfig.servers] : [];
    if (!updatedServers[index]) {
      updatedServers[index] = { name: "", description: "", allowedTools: [] };
    }
    updatedServers[index] = { ...updatedServers[index], ...data };
    
    updateConfig({ servers: updatedServers });
  };

  const updateMCPServer = (index: number, data: Partial<{ type: string; name: string; command?: string; args?: string[]; env?: Record<string, string>; url?: string }>) => {
    if (!parsedConfig) return;
    
    const updatedServers = parsedConfig.mcpServers ? [...parsedConfig.mcpServers] : [];
    if (!updatedServers[index]) {
      updatedServers[index] = { type: "stdio", name: "" };
    }
    updatedServers[index] = { ...updatedServers[index], ...data };
    
    updateConfig({ mcpServers: updatedServers });
  };

  const addEnvVariable = () => {
    if (!newEnvKey.trim() || !parsedConfig) return;
    
    const updatedServers = parsedConfig.mcpServers ? [...parsedConfig.mcpServers] : [];
    if (updatedServers.length === 0) {
      updatedServers.push({ type: "stdio", name: parsedConfig.name || "", env: {} });
    }
    
    if (!updatedServers[0].env) {
      updatedServers[0].env = {};
    }
    
    updatedServers[0].env[newEnvKey] = newEnvValue;
    setEnvKeys([...envKeys, newEnvKey]);
    setNewEnvKey("");
    setNewEnvValue("");
    
    updateConfig({ mcpServers: updatedServers });
  };

  const removeEnvVariable = (key: string) => {
    if (!parsedConfig || !parsedConfig.mcpServers || parsedConfig.mcpServers.length === 0 || !parsedConfig.mcpServers[0].env) return;
    
    const updatedServers = [...parsedConfig.mcpServers];
    const updatedEnv = { ...updatedServers[0].env };
    delete updatedEnv[key];
    updatedServers[0].env = updatedEnv;
    
    setEnvKeys(envKeys.filter(k => k !== key));
    updateConfig({ mcpServers: updatedServers });
  };

  const handleProxyTypeChange = (type: string) => {
    setProxyType(type);
    
    // 根据所选类型创建初始配置结构
    if (type === "http" && parsedConfig) {
      // 从 MCP 切换到 HTTP
      const newConfig = { ...parsedConfig };
      delete newConfig.mcpServers;
      
      if (!newConfig.servers || newConfig.servers.length === 0) {
        newConfig.servers = [{
          name: newConfig.name || "",
          description: "HTTP Server",
          allowedTools: []
        }];
      }
      
      if (!newConfig.tools) {
        newConfig.tools = [];
      }
      
      if (!newConfig.routers || newConfig.routers.length === 0) {
        newConfig.routers = [{
          server: newConfig.servers && newConfig.servers.length > 0 ? newConfig.servers[0].name : "",
          prefix: `${getRandomLetters(4)}`
        }];
      }
      
      updateConfig(newConfig);
    } else if (type === "mcp" && parsedConfig) {
      // 从 HTTP 切换到 MCP
      const newConfig = { ...parsedConfig };
      delete newConfig.servers;
      delete newConfig.tools;
      
      if (!newConfig.mcpServers || newConfig.mcpServers.length === 0) {
        newConfig.mcpServers = [{
          type: "stdio",
          name: newConfig.name || "",
          command: "",
          args: [],
          env: {}
        }];
      }
      
      if (!newConfig.routers || newConfig.routers.length === 0) {
        // 查找当前选择的租户
        const selectedTenant = tenants.find(t => t.name === newConfig.tenant);
        const prefix = selectedTenant?.name === 'default' ? 
          '/mcp' : 
          `${selectedTenant?.prefix || '/' + newConfig.tenant}/mcp`;
        
        newConfig.routers = [{
          server: newConfig.mcpServers && newConfig.mcpServers.length > 0 ? newConfig.mcpServers[0].name : "",
          prefix: prefix
        }];
      }
      
      updateConfig(newConfig);
    }
  };

  const renderServerOptions = () => {
    if (proxyType === "http" && parsedConfig?.servers && parsedConfig.servers.length > 0) {
      return parsedConfig.servers.map(server => (
        <SelectItem key={server.name || "default"}>
          {server.name || t('common.name')}
        </SelectItem>
      ));
    } else if (proxyType === "mcp" && parsedConfig?.mcpServers && parsedConfig.mcpServers.length > 0) {
      return parsedConfig.mcpServers.map(server => (
        <SelectItem key={server.name || "default"}>
          {server.name || t('common.name')}
        </SelectItem>
      ));
    }
    
    return <SelectItem key="default">{t('common.name')}</SelectItem>;
  };

  // 工具配置部分
  const renderToolsConfig = () => {
    return (
      <div className="border-t pt-4 mt-2">
        <h3 className="text-sm font-medium mb-2">{t('gateway.tools_config')}</h3>
        {/* 工具配置相对复杂，这里只展示部分 */}
        {(parsedConfig?.tools || []).map((tool, index) => (
          <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
            {/* 工具基本配置 */}
            <Input
              label={t('gateway.tool_name')}
              value={(toolFormState[index]?.name !== undefined) ? toolFormState[index]?.name : (tool.name || "")}
              onChange={(e) => {
                // 更新临时表单状态
                setToolFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    name: e.target.value
                  }
                }));
                updateToolForm({
                  ...toolFormState,
                  [index]: {
                    ...(toolFormState[index] || {}),
                    name: e.target.value
                  }
                });
              }}
            />
            <Input
              label={t('gateway.description')}
              value={(toolFormState[index]?.description !== undefined) ? toolFormState[index]?.description : (tool.description || "")}
              onChange={(e) => {
                setToolFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    description: e.target.value
                  }
                }));
                updateToolForm({
                  ...toolFormState,
                  [index]: {
                    ...(toolFormState[index] || {}),
                    description: e.target.value
                  }
                });
              }}
            />
            <Select
              label={t('gateway.method')}
              selectedKeys={[tool.method || "GET"]}
              onChange={(e) => {
                const updatedTools = parsedConfig?.tools ? [...parsedConfig.tools] : [];
                updatedTools[index] = { ...tool, method: e.target.value };
                updateConfig({ tools: updatedTools });
              }}
              aria-label={t('gateway.method')}
            >
              <SelectItem key="GET">GET</SelectItem>
              <SelectItem key="POST">POST</SelectItem>
              <SelectItem key="PUT">PUT</SelectItem>
              <SelectItem key="DELETE">DELETE</SelectItem>
            </Select>
            <Input
              label={t('gateway.endpoint')}
              value={(toolFormState[index]?.endpoint !== undefined) ? toolFormState[index]?.endpoint : (tool.endpoint || "")}
              onChange={(e) => {
                setToolFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    endpoint: e.target.value
                  }
                }));
                updateToolForm({
                  ...toolFormState,
                  [index]: {
                    ...(toolFormState[index] || {}),
                    endpoint: e.target.value
                  }
                });
              }}
            />
            
            {/* Headers 配置 */}
            <div className="mt-2 border-t pt-2">
              <h4 className="text-sm font-medium mb-2">Headers</h4>
              <div className="flex flex-col gap-2">
                {(tool.headersOrder || Object.keys(tool.headers || {})).map((key, headerIndex) => (
                  <div key={headerIndex} className="flex items-center gap-2">
                    <Input
                      className="flex-1"
                      value={(headerFormState[index]?.[`${headerIndex}_${key}`]?.key !== undefined) 
                        ? headerFormState[index][`${headerIndex}_${key}`].key 
                        : key}
                      onChange={(e) => {
                        setHeaderFormState(prev => ({
                          ...prev,
                          [index]: {
                            ...(prev[index] || {}),
                            [`${headerIndex}_${key}`]: {
                              ...(prev[index]?.[`${headerIndex}_${key}`] || {}),
                              key: e.target.value
                            }
                          }
                        }));
                        updateHeaderForm({
                          ...headerFormState,
                          [index]: {
                            ...(headerFormState[index] || {}),
                            [`${headerIndex}_${key}`]: {
                              ...(headerFormState[index]?.[`${headerIndex}_${key}`] || {}),
                              key: e.target.value
                            }
                          }
                        });
                      }}
                      placeholder="Header名称"
                    />
                    <Input
                      className="flex-1"
                      value={(headerFormState[index]?.[`${headerIndex}_${key}`]?.value !== undefined)
                        ? headerFormState[index][`${headerIndex}_${key}`].value
                        : (tool.headers?.[key] || "")}
                      onChange={(e) => {
                        setHeaderFormState(prev => ({
                          ...prev,
                          [index]: {
                            ...(prev[index] || {}),
                            [`${headerIndex}_${key}`]: {
                              ...(prev[index]?.[`${headerIndex}_${key}`] || {}),
                              value: e.target.value
                            }
                          }
                        }));
                        updateHeaderForm({
                          ...headerFormState,
                          [index]: {
                            ...(headerFormState[index] || {}),
                            [`${headerIndex}_${key}`]: {
                              ...(headerFormState[index]?.[`${headerIndex}_${key}`] || {}),
                              value: e.target.value
                            }
                          }
                        });
                      }}
                      placeholder="Header值"
                    />
                    <Button
                      isIconOnly
                      color="danger"
                      onPress={() => {
                        const updatedTools = parsedConfig?.tools ? [...parsedConfig.tools] : [];
                        const updatedTool = { ...updatedTools[index] };
                        const updatedHeaders = updatedTool.headers ? { ...updatedTool.headers } : {};
                        const updatedHeadersOrder = updatedTool.headersOrder ? [...updatedTool.headersOrder] : Object.keys(updatedHeaders);
                        
                        // 删除header
                        delete updatedHeaders[key];
                        
                        // 更新顺序
                        updatedHeadersOrder.splice(headerIndex, 1);
                        
                        updatedTool.headers = updatedHeaders;
                        updatedTool.headersOrder = updatedHeadersOrder;
                        updatedTools[index] = updatedTool;
                        updateConfig({ tools: updatedTools });
                        
                        // 清理对应的表单状态
                        if (headerFormState[index]) {
                          const updatedHeaderFormState = { ...headerFormState };
                          
                          // 删除当前header的表单状态
                          if (updatedHeaderFormState[index][`${headerIndex}_${key}`]) {
                            delete updatedHeaderFormState[index][`${headerIndex}_${key}`];
                          }
                          
                          // 更新后续header的索引
                          for (let i = headerIndex + 1; i < (tool.headersOrder || Object.keys(tool.headers || {})).length; i++) {
                            const nextKey = (tool.headersOrder || Object.keys(tool.headers || {}))[i];
                            if (updatedHeaderFormState[index][`${i}_${nextKey}`]) {
                              updatedHeaderFormState[index][`${i-1}_${nextKey}`] = updatedHeaderFormState[index][`${i}_${nextKey}`];
                              delete updatedHeaderFormState[index][`${i}_${nextKey}`];
                            }
                          }
                          
                          setHeaderFormState(updatedHeaderFormState);
                        }
                      }}
                    >
                      ✕
                    </Button>
                  </div>
                ))}
                
                {/* 添加新的Header */}
                <Button
                  color="primary"
                  size="sm"
                  className="mt-1"
                  onPress={() => {
                    // 查找一个唯一的Key
                    let newKey = "Content-Type";
                    let count = 1;
                    
                    // 如果Content-Type已存在，尝试其他常用header
                    const commonHeaders = [
                      "Authorization", 
                      "Accept", 
                      "X-API-Key", 
                      "User-Agent", 
                      "Cache-Control"
                    ];
                    
                    // 获取现有header keys
                    const existingKeys = tool.headersOrder || Object.keys(tool.headers || {});
                    
                    // 先尝试从常用header中找一个不存在的
                    for (const header of commonHeaders) {
                      if (!existingKeys.includes(header)) {
                        newKey = header;
                        break;
                      }
                    }
                    
                    // 如果所有常用header都已存在，创建一个带数字后缀的header
                    if (existingKeys.includes(newKey)) {
                      while (existingKeys.includes(`X-Header-${count}`)) {
                        count++;
                      }
                      newKey = `X-Header-${count}`;
                    }
                    
                    // 添加到配置
                    const updatedTools = parsedConfig?.tools ? [...parsedConfig.tools] : [];
                    const updatedTool = { ...updatedTools[index] };
                    const updatedHeaders = updatedTool.headers ? { ...updatedTool.headers } : {};
                    const updatedHeadersOrder = updatedTool.headersOrder ? [...updatedTool.headersOrder] : Object.keys(updatedHeaders);
                    
                    updatedHeaders[newKey] = "";
                    updatedHeadersOrder.push(newKey);
                    
                    updatedTool.headers = updatedHeaders;
                    updatedTool.headersOrder = updatedHeadersOrder;
                    updatedTools[index] = updatedTool;
                    updateConfig({ tools: updatedTools });
                  }}
                >
                  添加Header
                </Button>
              </div>
            </div>
            
            {/* Args 配置和其他工具配置部分... */}
            <div className="mt-2 border-t pt-2">
              <h4 className="text-sm font-medium mb-2">参数 (Args)</h4>
              {/* Args 配置内容 */}
            </div>
            
            {/* Request Body */}
            <div className="mt-2 border-t pt-2">
              <h4 className="text-sm font-medium mb-2">请求体 (Request Body)</h4>
              <textarea
                className="w-full border rounded p-2"
                rows={5}
                value={(toolFormState[index]?.requestBody !== undefined) ? toolFormState[index]?.requestBody : (tool.requestBody || "")}
                onChange={(e) => {
                  setToolFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      requestBody: e.target.value
                    }
                  }));
                  updateToolForm({
                    ...toolFormState,
                    [index]: {
                      ...(toolFormState[index] || {}),
                      requestBody: e.target.value
                    }
                  });
                }}
                placeholder='例如: {"uid": "{{.Args.uid}}"}'
              ></textarea>
            </div>
            
            {/* Response Body */}
            <div className="mt-2 border-t pt-2">
              <h4 className="text-sm font-medium mb-2">响应体 (Response Body)</h4>
              <textarea
                className="w-full border rounded p-2"
                rows={5}
                value={(toolFormState[index]?.responseBody !== undefined) ? toolFormState[index]?.responseBody : (tool.responseBody || "")}
                onChange={(e) => {
                  setToolFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      responseBody: e.target.value
                    }
                  }));
                  updateToolForm({
                    ...toolFormState,
                    [index]: {
                      ...(toolFormState[index] || {}),
                      responseBody: e.target.value
                    }
                  });
                }}
                placeholder="例如: {{.Response.Body}}"
              ></textarea>
            </div>
          </div>
        ))}
        {/* 添加工具按钮 */}
        <Button
          color="primary"
          className="mt-2"
          onPress={() => {
            const updatedTools = parsedConfig?.tools ? [...parsedConfig.tools] : [];
            updatedTools.push({ 
              name: "", 
              description: "", 
              method: "GET", 
              endpoint: "",
              headers: {
                "Content-Type": "application/json"
              },
              headersOrder: ["Content-Type"],
              args: [],
              requestBody: "",
              responseBody: "{{.Response.Body}}"
            });
            updateConfig({ tools: updatedTools });
          }}
        >
          {t('gateway.add_tool')}
        </Button>
      </div>
    );
  };
  
  // 服务器配置部分
  const renderServersConfig = () => {
    const servers = parsedConfig?.servers || [{ name: "", description: "", allowedTools: [] }];
    
    return (
      <div className="border-t pt-4 mt-2">
        <h3 className="text-sm font-medium mb-2">{t('gateway.server_config')}</h3>
        {servers.map((server, index) => (
          <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
            <Input
              label={t('gateway.server_name')}
              value={serverFormState[index]?.name !== undefined ? serverFormState[index].name : (server.name || "")}
              onChange={(e) => {
                setServerFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    name: e.target.value
                  }
                }));
                updateServerForm({
                  ...serverFormState,
                  [index]: {
                    ...(serverFormState[index] || {}),
                    name: e.target.value
                  }
                });
              }}
            />
            <Input
              label={t('gateway.description')}
              value={serverFormState[index]?.description !== undefined ? serverFormState[index].description : (server.description || "")}
              onChange={(e) => {
                setServerFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    description: e.target.value
                  }
                }));
                updateServerForm({
                  ...serverFormState,
                  [index]: {
                    ...(serverFormState[index] || {}),
                    description: e.target.value
                  }
                });
              }}
            />
            <div>
              <h4 className="text-sm font-medium mb-2">{t('gateway.allowed_tools')}</h4>
              <div className="flex flex-wrap gap-1">
                {server.allowedTools && server.allowedTools.map((tool, toolIndex) => (
                  <Chip 
                    key={toolIndex}
                    onClose={() => {
                      const updated = [...server.allowedTools];
                      updated.splice(toolIndex, 1);
                      updateServer(index, { allowedTools: updated });
                    }}
                  >
                    {tool}
                  </Chip>
                ))}
              </div>
              <div className="mt-2">
                <h4 className="text-sm font-medium mb-2">{t('gateway.add_tool')}</h4>
                <div className="flex flex-wrap gap-2">
                  {(parsedConfig?.tools || [])
                    .filter(tool => !server.allowedTools?.includes(tool.name || ""))
                    .map(tool => (
                      <Button 
                        key={tool.name}
                        size="sm"
                        variant="flat"
                        color="primary"
                        className="min-w-0"
                        onPress={() => {
                          if (tool.name && server.allowedTools && !server.allowedTools.includes(tool.name)) {
                            updateServer(index, { 
                              allowedTools: [...server.allowedTools, tool.name] 
                            });
                          }
                        }}
                      >
                        + {tool.name || t('common.name')}
                      </Button>
                    ))
                  }
                  {(parsedConfig?.tools || []).length > 0 && 
                   (parsedConfig?.tools || []).every(tool => server.allowedTools?.includes(tool.name || "")) && (
                    <span className="text-sm text-gray-500">{t('gateway.tools')} {t('common.already')} {t('common.all')} {t('common.add')}</span>
                  )}
                  {(parsedConfig?.tools || []).length === 0 && (
                    <span className="text-sm text-gray-500">{t('gateway.tools')} {t('common.none')} {t('common.available')}</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  };
  
  const renderHttpServers = () => {
    return (
      <>
        {renderToolsConfig()}
        {renderServersConfig()}
        {renderRouterConfig()}
      </>
    );
  };

  const renderMCPServers = () => {
    // 即使parsedConfig为null或mcpServers为空，也显示一个默认的MCP服务器配置表单
    const mcpServers = parsedConfig?.mcpServers || [{ type: "stdio", name: "", command: "", args: [], env: {} }];
    
    return (
      <>
        <div className="border-t pt-4 mt-2">
          <h3 className="text-sm font-medium mb-2">{t('gateway.mcp_server_config')}</h3>
          {mcpServers.map((server, index) => (
            <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
              <Input
                label={t('gateway.server_name')}
                value={server.name || ""}
                onChange={(e) => updateMCPServer(index, { name: e.target.value })}
              />
              
              <Select
                label={t('gateway.mcp_type')}
                selectedKeys={[server.type || "stdio"]}
                onChange={(e) => updateMCPServer(index, { type: e.target.value })}
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
                    value={server.command || ''}
                    onChange={(e) => updateMCPServer(index, { command: e.target.value })}
                  />
                  
                  <Input
                    label={t('gateway.args')}
                    value={server.args?.join(' ') || ''}
                    onChange={(e) => {
                      const args = e.target.value.split(' ').filter(Boolean);
                      updateMCPServer(index, { args });
                    }}
                    placeholder="arg1 arg2 arg3"
                  />
                  
                  <div className="mt-2">
                    <h4 className="text-sm font-medium mb-2">{t('gateway.env_variables')}</h4>
                    {envKeys.map(key => (
                      <div key={key} className="flex items-center gap-2 mb-2">
                        <Input
                          className="flex-1"
                          value={key}
                          isReadOnly
                        />
                        <Input
                          className="flex-1"
                          value={server.env?.[key] || ''}
                          onChange={(e) => {
                            const updatedEnv = { ...(server.env || {}) };
                            updatedEnv[key] = e.target.value;
                            updateMCPServer(index, { env: updatedEnv });
                          }}
                        />
                        <Button
                          color="danger"
                          isIconOnly
                          size="sm"
                          onPress={() => removeEnvVariable(key)}
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
                        onPress={addEnvVariable}
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
                  value={server.url || ''}
                  onChange={(e) => updateMCPServer(index, { url: e.target.value })}
                />
              )}
            </div>
          ))}
        </div>
        
        {renderRouterConfig()}
      </>
    );
  };

  // 单独创建路由配置渲染函数，供HTTP和MCP模式共用
  const renderRouterConfig = () => {
    // 即使parsedConfig为null或routers为空，也显示一个默认的路由配置表单
    const selectedTenant = tenants.find(t => t.name === parsedConfig?.tenant);
    
    const routers = parsedConfig?.routers || [{ 
      server: "", 
      prefix: "/"
    }];
    
    return (
      <div className="border-t pt-4 mt-2">
        <h3 className="text-sm font-medium mb-2">{t('gateway.router_config')}</h3>
        {routers.map((router, index) => (
          <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
            <div className="flex gap-2">
                <Input
                  label={t('gateway.prefix')}
                  value={
                    routerFormState[index]?.prefix !== undefined 
                      ? routerFormState[index].prefix.replace(selectedTenant?.prefix || "", "")
                      : (router.prefix || "").replace(selectedTenant?.prefix || "", "")
                  }
                  startContent={
                    <div className="pointer-events-none flex items-center">
                      <span className="text-default-400 text-small">{selectedTenant?.prefix}</span>
                    </div>
                  }
                  onChange={(e) => {
                    // Get the path part from user input and validate it
                    const pathPart = e.target.value.trim();
                    // 完整前缀 = 租户前缀 + 用户输入
                    const fullPrefix = `${selectedTenant?.prefix}${pathPart}`;

                    console.log(fullPrefix)
                    
                    setRouterFormState(prev => ({
                      ...prev,
                      [index]: {
                        ...(prev[index] || {}),
                        prefix: fullPrefix
                      }
                    }));
                    updateRouterForm({
                      ...routerFormState,
                      [index]: {
                        ...(routerFormState[index] || {}),
                        prefix: fullPrefix
                      }
                    });
                  }}
                  className="flex-1"
                />
              <Select
                label={t('gateway.server')}
                selectedKeys={routerFormState[index]?.server !== undefined ? [routerFormState[index].server] : (router.server ? [router.server] : [])}
                className="flex-1"
                aria-label={t('gateway.server')}
                onChange={(e) => {
                  setRouterFormState(prev => ({
                    ...prev,
                    [index]: {
                      ...(prev[index] || {}),
                      server: e.target.value
                    }
                  }));
                  updateRouterForm({
                    ...routerFormState,
                    [index]: {
                      ...(routerFormState[index] || {}),
                      server: e.target.value
                    }
                  });
                }}
              >
                {renderServerOptions()}
              </Select>
              <Button
                isIconOnly
                color="danger"
                className="self-end mb-2"
                onPress={() => {
                  if (parsedConfig?.routers && parsedConfig.routers.length > 1) {
                    const updatedRouters = [...parsedConfig.routers];
                    updatedRouters.splice(index, 1);
                    updateConfig({ routers: updatedRouters });
                  }
                }}
                // 如果只有一个路由则禁用删除按钮
                isDisabled={!parsedConfig?.routers || parsedConfig.routers.length <= 1}
              >
                ✕
              </Button>
            </div>
            
            {/* CORS配置部分 */}
            <div className="mt-3">
              <div className="flex items-center gap-2">
                <Switch 
                  size="sm"
                  isSelected={Boolean(router.cors)}
                  onValueChange={(isSelected) => {
                    if (isSelected) {
                      // 启用CORS并设置默认值
                      updateRouter(index, {
                        cors: {
                          allowOrigins: ['*'],
                          allowMethods: ['GET', 'POST', 'PUT', 'OPTIONS'],
                          allowHeaders: ['Content-Type', 'Authorization', 'Mcp-Session-Id'],
                          exposeHeaders: ['Mcp-Session-Id'],
                          allowCredentials: true
                        }
                      });
                    } else {
                      // 禁用CORS - 确保完全删除cors属性
                      if (parsedConfig?.routers && parsedConfig.routers[index]) {
                        const updatedRouters = [...parsedConfig.routers];
                        const { cors: _, ...restRouter } = updatedRouters[index];
                        updatedRouters[index] = restRouter;
                        updateConfig({ routers: updatedRouters });
                      }
                    }
                  }}
                />
                <span className="text-sm font-medium">{t('gateway.enable_cors')}</span>
              </div>
              
              {router.cors && renderCorsConfig(router, index)}
            </div>
          </div>
        ))}
        {/* 添加路由按钮 */}
        <Button
          color="primary"
          className="mt-2"
          onPress={() => {
            const updatedRouters = parsedConfig?.routers ? [...parsedConfig.routers] : [];
            const serverName = proxyType === "http" 
              ? (parsedConfig?.servers && parsedConfig.servers.length > 0 ? parsedConfig.servers[0].name : "") 
              : (parsedConfig?.mcpServers && parsedConfig.mcpServers.length > 0 ? parsedConfig.mcpServers[0].name : "");
            
            updatedRouters.push({ 
              server: serverName,
              prefix: '/' + getRandomLetters(4)
            });
            updateConfig({ routers: updatedRouters });
          }}
        >
          {t('common.add')}
        </Button>
      </div>
    );
  };

  // 修复CORS相关部分的类型问题
  const renderCorsConfig = (router: { cors?: Record<string, unknown> }, index: number) => {
    const corsConfig = router.cors as CorsConfig;
    if (!corsConfig) return null;
    
    return (
      <div className="mt-2 pl-4 border-l-2 border-gray-200">
        {/* 允许的源 */}
        <div className="mb-3">
          <h4 className="text-sm font-medium mb-1">{t('gateway.allow_origins')}</h4>
          <div className="flex flex-wrap gap-1 mb-1">
            {(corsConfig.allowOrigins || []).map((origin: string, originIndex: number) => (
              <Chip 
                key={originIndex}
                onClose={() => {
                  const updatedCors = {...corsConfig};
                  updatedCors.allowOrigins = (updatedCors.allowOrigins || []).filter((_: string, i: number) => i !== originIndex);
                  updateRouter(index, { cors: updatedCors });
                }}
              >
                {origin}
              </Chip>
            ))}
          </div>
          <div className="flex gap-2">
            <Input 
              size="sm"
              placeholder="例如: https://example.com 或 *"
              className="flex-1"
              value={newOrigin[index] || ''}
              onChange={(e) => {
                setNewOrigin({
                  ...newOrigin,
                  [index]: e.target.value
                });
              }}
            />
            <Button
              size="sm"
              onPress={() => {
                if (newOrigin[index]?.trim()) {
                  const updatedCors = {...corsConfig};
                  updatedCors.allowOrigins = [...(updatedCors.allowOrigins || []), newOrigin[index].trim()];
                  updateRouter(index, { cors: updatedCors });
                  setNewOrigin({
                    ...newOrigin,
                    [index]: ''
                  });
                }
              }}
            >
              {t('common.add')}
            </Button>
          </div>
        </div>
        
        {/* 允许的方法 */}
        <div className="mb-3">
          <h4 className="text-sm font-medium mb-1">{t('gateway.allow_methods')}</h4>
          <div className="flex flex-wrap gap-1 mb-1">
            {(corsConfig.allowMethods || []).map((method: string, methodIndex: number) => (
              <Chip 
                key={methodIndex}
                onClose={() => {
                  const updatedCors = {...corsConfig};
                  updatedCors.allowMethods = (updatedCors.allowMethods || []).filter((_: string, i: number) => i !== methodIndex);
                  updateRouter(index, { cors: updatedCors });
                }}
              >
                {method}
              </Chip>
            ))}
          </div>
          <div className="flex gap-2">
            <Select
              size="sm"
              className="flex-1"
              id={`method-select-${index}`}
              aria-label={t('gateway.http_method')}
              selectedKeys={selectedMethod[index] ? [selectedMethod[index]] : []}
              onChange={(e) => {
                setSelectedMethod({
                  ...selectedMethod,
                  [index]: e.target.value
                });
              }}
            >
              {['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS', 'HEAD', 'PATCH'].map(method => (
                <SelectItem key={method}>{method}</SelectItem>
              ))}
            </Select>
            <Button
              size="sm"
              onPress={() => {
                if (selectedMethod[index]) {
                  const method = selectedMethod[index];
                  const updatedCors = {...corsConfig};
                  // 确保方法不重复
                  if (!(updatedCors.allowMethods || []).includes(method)) {
                    updatedCors.allowMethods = [...(updatedCors.allowMethods || []), method];
                    updateRouter(index, { cors: updatedCors });
                  }
                }
              }}
            >
              {t('common.add')}
            </Button>
          </div>
        </div>
        
        {/* 允许的头部 */}
        <div className="mb-3">
          <h4 className="text-sm font-medium mb-1">{t('gateway.allow_headers')}</h4>
          <div className="flex flex-wrap gap-1 mb-1">
            {(corsConfig.allowHeaders || []).map((header: string, headerIndex: number) => (
              <Chip 
                key={headerIndex}
                onClose={() => {
                  const updatedCors = {...corsConfig};
                  updatedCors.allowHeaders = (updatedCors.allowHeaders || []).filter((_: string, i: number) => i !== headerIndex);
                  updateRouter(index, { cors: updatedCors });
                }}
              >
                {header}
              </Chip>
            ))}
          </div>
          <div className="flex gap-2">
            <Input 
              size="sm"
              placeholder="例如: Content-Type"
              className="flex-1"
              list={`common-headers-${index}`}
              value={newHeader[index] || ''}
              onChange={(e) => {
                setNewHeader({
                  ...newHeader,
                  [index]: e.target.value
                });
              }}
            />
            <datalist id={`common-headers-${index}`}>
              <option value="Content-Type" />
              <option value="Authorization" />
              <option value="X-Requested-With" />
              <option value="Accept" />
              <option value="Origin" />
              <option value="Mcp-Session-Id" />
            </datalist>
            <Button
              size="sm"
              onPress={() => {
                if (newHeader[index]?.trim()) {
                  const updatedCors = {...corsConfig};
                  updatedCors.allowHeaders = [...(updatedCors.allowHeaders || []), newHeader[index].trim()];
                  updateRouter(index, { cors: updatedCors });
                  setNewHeader({
                    ...newHeader,
                    [index]: ''
                  });
                }
              }}
            >
              {t('common.add')}
            </Button>
          </div>
        </div>
        
        {/* 暴露的头部 */}
        <div className="mb-3">
          <h4 className="text-sm font-medium mb-1">{t('gateway.expose_headers')}</h4>
          <div className="flex flex-wrap gap-1 mb-1">
            {(corsConfig.exposeHeaders || []).map((header: string, headerIndex: number) => (
              <Chip 
                key={headerIndex}
                onClose={() => {
                  const updatedCors = {...corsConfig};
                  updatedCors.exposeHeaders = (updatedCors.exposeHeaders || []).filter((_: string, i: number) => i !== headerIndex);
                  updateRouter(index, { cors: updatedCors });
                }}
              >
                {header}
              </Chip>
            ))}
          </div>
          <div className="flex gap-2">
            <Input 
              size="sm"
              placeholder="例如: Content-Length"
              className="flex-1"
              list={`common-expose-headers-${index}`}
              value={newExposeHeader[index] || ''}
              onChange={(e) => {
                setNewExposeHeader({
                  ...newExposeHeader,
                  [index]: e.target.value
                });
              }}
            />
            <datalist id={`common-expose-headers-${index}`}>
              <option value="Content-Length" />
              <option value="Mcp-Session-Id" />
              <option value="X-Rate-Limit" />
            </datalist>
            <Button
              size="sm"
              onPress={() => {
                if (newExposeHeader[index]?.trim()) {
                  const updatedCors = {...corsConfig};
                  updatedCors.exposeHeaders = [...(updatedCors.exposeHeaders || []), newExposeHeader[index].trim()];
                  updateRouter(index, { cors: updatedCors });
                  setNewExposeHeader({
                    ...newExposeHeader,
                    [index]: ''
                  });
                }
              }}
            >
              {t('common.add')}
            </Button>
          </div>
        </div>
        
        {/* 允许携带凭证 */}
        <div className="mb-3 flex items-center gap-2">
          <Switch 
            size="sm"
            isSelected={Boolean(corsConfig.allowCredentials)}
            onValueChange={(isSelected) => {
              const updatedCors = {...corsConfig};
              updatedCors.allowCredentials = isSelected;
              updateRouter(index, { cors: updatedCors });
            }}
          />
          <span className="text-sm">{t('gateway.credentials')}</span>
        </div>
      </div>
    );
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
      const selectedTenant = tenants.find(t => t.name === parsedConfig?.tenant);
      
      const initialFormState: Record<number, any> = parsedConfig.routers.reduce((acc: Record<number, any>, router, idx) => {
        // 检查是否需要应用默认值
        const tenantPrefix = selectedTenant?.prefix || "";
        let pathPart = (router.prefix || "").replace(tenantPrefix, "");
        
        acc[idx] = {
          ...router,
          prefix: tenantPrefix + pathPart
        };
        return acc;
      }, {});
      
      setRouterFormState(initialFormState);
    }
  }, [parsedConfig?.routers, parsedConfig?.tenant, tenants]);

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
          height="90%"
          defaultLanguage="yaml"
          value={config}
          onChange={(value) => {
            if (!value) {
              onChange('');
              return;
            }
            
            // 如果在编辑模式下且已存在配置且有name，则确保不会修改name
            if (isEditing && parsedConfig?.name && parsedConfig.name.trim() !== '') {
              try {
                const parsedValue = yaml.load(value) as GatewayConfig;
                if (parsedValue.name !== parsedConfig.name) {
                  // 如果name被修改了，恢复原始name
                  parsedValue.name = parsedConfig.name;
                  // 重新生成YAML
                  const newYaml = yaml.dump(parsedValue);
                  onChange(newYaml);
                  return;
                }
              } catch (e) {
                // 解析错误，直接使用原始输入
                console.error("Failed to parse YAML for name check:", e);
              }
            }
            
            // 正常更新
            onChange(value);
          }}
          theme={isDark ? "vs-dark" : "vs"}
          options={editorOptions}
        />
      ) : (
        <div className="flex flex-col gap-4 h-full overflow-y-auto">
          <Input 
            label={t('gateway.name')}
            value={generalFormState.name !== undefined ? generalFormState.name : (parsedConfig?.name || '')}
            onChange={(e) => {
              setGeneralFormState(prev => ({
                ...prev,
                name: e.target.value
              }));
              updateGeneralForm({
                ...generalFormState,
                name: e.target.value
              });
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
              updateGeneralForm({
                ...generalFormState,
                tenant: e.target.value
              });
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
          
          <div className="border-t pt-4 mt-2">
            <h3 className="text-sm font-medium mb-2">{t('gateway.proxy_type')}</h3>
            <RadioGroup
              value={proxyType}
              onValueChange={handleProxyTypeChange}
              orientation="horizontal"
            >
              <Radio value="http">HTTP Proxy</Radio>
              <Radio value="mcp">MCP Proxy</Radio>
            </RadioGroup>
          </div>
          
          {proxyType === "http" && renderHttpServers()}
          
          {proxyType === "mcp" && renderMCPServers()}
        </div>
      )}
    </div>
  );
}
