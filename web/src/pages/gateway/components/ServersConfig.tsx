import { Input, Button, Chip } from "@heroui/react";
import { useTranslation } from 'react-i18next';

import { GatewayConfig } from '../types';

interface ServersConfigProps {
  parsedConfig: GatewayConfig;
  serverFormState: {[serverIndex: number]: {name?: string; description?: string}};
  setServerFormState: (state: {[serverIndex: number]: {name?: string; description?: string}}) => void;
  updateConfig: (newData: Partial<GatewayConfig>) => void;
}

export function ServersConfig({
  parsedConfig,
  serverFormState,
  setServerFormState,
  updateConfig
}: ServersConfigProps) {
  const { t } = useTranslation();
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
                    updateConfig({
                      servers: servers.map((s, i) => 
                        i === index ? { ...s, allowedTools: updated } : s
                      )
                    });
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
                          updateConfig({
                            servers: servers.map((s, i) => 
                              i === index ? { 
                                ...s, 
                                allowedTools: [...s.allowedTools, tool.name] 
                              } : s
                            )
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
} 