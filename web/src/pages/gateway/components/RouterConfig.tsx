import { Input, Select, SelectItem, Button, Switch, Chip } from "@heroui/react";
import { useTranslation } from 'react-i18next';

import { GatewayConfig, CorsConfig, Tenant } from '../types';

interface RouterConfigProps {
  parsedConfig: GatewayConfig;
  routerFormState: {[routerIndex: number]: {prefix?: string; server?: string}};
  setRouterFormState: (state: {[routerIndex: number]: {prefix?: string; server?: string}}) => void;
  updateConfig: (newData: Partial<GatewayConfig>) => void;
  tenants: Tenant[];
  selectedMethod: {[routerIndex: number]: string};
  setSelectedMethod: (state: {[routerIndex: number]: string}) => void;
  newOrigin: {[routerIndex: number]: string};
  setNewOrigin: (state: {[routerIndex: number]: string}) => void;
  newExposeHeader: {[routerIndex: number]: string};
  setNewExposeHeader: (state: {[routerIndex: number]: string}) => void;
  newHeader: {[routerIndex: number]: string};
  setNewHeader: (state: {[routerIndex: number]: string}) => void;
  renderServerOptions: () => JSX.Element[];
}

export function RouterConfig({
  parsedConfig,
  routerFormState,
  setRouterFormState,
  updateConfig,
  tenants,
  selectedMethod,
  setSelectedMethod,
  newOrigin,
  setNewOrigin,
  newExposeHeader,
  setNewExposeHeader,
  newHeader,
  setNewHeader,
  renderServerOptions
}: RouterConfigProps) {
  const { t } = useTranslation();
  const selectedTenant = tenants.find(t => t.name === parsedConfig?.tenant);
  const routers = parsedConfig?.routers || [{ server: "", prefix: "/" }];

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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  if (!(updatedCors.allowMethods || []).includes(method)) {
                    updatedCors.allowMethods = [...(updatedCors.allowMethods || []), method];
                    updateConfig({
                      routers: routers.map((r, i) => 
                        i === index ? { ...r, cors: updatedCors } : r
                      )
                    });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
                  updateConfig({
                    routers: routers.map((r, i) => 
                      i === index ? { ...r, cors: updatedCors } : r
                    )
                  });
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
              updateConfig({
                routers: routers.map((r, i) => 
                  i === index ? { ...r, cors: updatedCors } : r
                )
              });
            }}
          />
          <span className="text-sm">{t('gateway.credentials')}</span>
        </div>
      </div>
    );
  };

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
                const pathPart = e.target.value.trim();
                const fullPrefix = `${selectedTenant?.prefix}${pathPart}`;
                
                setRouterFormState(prev => ({
                  ...prev,
                  [index]: {
                    ...(prev[index] || {}),
                    prefix: fullPrefix
                  }
                }));
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
              }}
            >
              {renderServerOptions()}
            </Select>
            <Button
              isIconOnly
              color="danger"
              className="self-end mb-2"
              onPress={() => {
                if (routers.length > 1) {
                  const updatedRouters = [...routers];
                  updatedRouters.splice(index, 1);
                  updateConfig({ routers: updatedRouters });
                }
              }}
              isDisabled={routers.length <= 1}
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
                    updateConfig({
                      routers: routers.map((r, i) => 
                        i === index ? {
                          ...r,
                          cors: {
                            allowOrigins: ['*'],
                            allowMethods: ['GET', 'POST', 'PUT', 'OPTIONS'],
                            allowHeaders: ['Content-Type', 'Authorization', 'Mcp-Session-Id'],
                            exposeHeaders: ['Mcp-Session-Id'],
                            allowCredentials: true
                          }
                        } : r
                      )
                    });
                  } else {
                    const updatedRouters = [...routers];
                    const { cors: _, ...restRouter } = updatedRouters[index];
                    updatedRouters[index] = restRouter;
                    updateConfig({ routers: updatedRouters });
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
          const updatedRouters = [...routers];
          const serverName = parsedConfig?.servers?.[0]?.name || parsedConfig?.mcpServers?.[0]?.name || "";
          
          updatedRouters.push({ 
            server: serverName,
            prefix: '/' + Math.random().toString(36).substring(2, 6)
          });
          updateConfig({ routers: updatedRouters });
        }}
      >
        {t('common.add')}
      </Button>
    </div>
  );
} 