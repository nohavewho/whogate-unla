import { Input, Select, SelectItem, Button } from "@heroui/react";
import { useTranslation } from 'react-i18next';

import { GatewayConfig, HeadersFormState, KeyValueItem } from '../types';

interface ToolsConfigProps {
  parsedConfig: GatewayConfig;
  toolFormState: {[toolIndex: number]: {[field: string]: string}};
  headerFormState: HeadersFormState;
  setToolFormState: (state: {[toolIndex: number]: {[field: string]: string}}) => void;
  updateConfig: (newData: Partial<GatewayConfig>) => void;
  addHeader: (toolIndex: number, key: string, value?: string) => void;
  removeHeader: (toolIndex: number, headerIndex: number) => void;
  updateHeader: (toolIndex: number, headerIndex: number, updates: Partial<KeyValueItem>) => void;
}

export function ToolsConfig({
  parsedConfig,
  toolFormState,
  headerFormState,
  setToolFormState,
  updateConfig,
  addHeader,
  removeHeader,
  updateHeader
}: ToolsConfigProps) {
  const { t } = useTranslation();

  return (
    <div className="border-t pt-4 mt-2">
      <h3 className="text-sm font-medium mb-2">{t('gateway.tools_config')}</h3>
      {(parsedConfig?.tools || []).map((tool, index) => (
        <div key={index} className="flex flex-col gap-2 mb-4 p-3 border rounded-md">
          <Input
            label={t('gateway.tool_name')}
            value={(toolFormState[index]?.name !== undefined) ? toolFormState[index]?.name : (tool.name || "")}
            onChange={(e) => {
              setToolFormState(prev => ({
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
            value={(toolFormState[index]?.description !== undefined) ? toolFormState[index]?.description : (tool.description || "")}
            onChange={(e) => {
              setToolFormState(prev => ({
                ...prev,
                [index]: {
                  ...(prev[index] || {}),
                  description: e.target.value
                }
              }));
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
                    value={(headerFormState[index]?.[headerIndex]?.key !== undefined) 
                      ? headerFormState[index][headerIndex].key 
                      : key}
                    onChange={(e) => {
                      updateHeader(index, headerIndex, {
                        key: e.target.value
                      });
                    }}
                    placeholder="Header名称"
                  />
                  <Input
                    className="flex-1"
                    value={(headerFormState[index]?.[headerIndex]?.value !== undefined)
                      ? headerFormState[index][headerIndex].value
                      : (tool.headers?.[key] || "")}
                    onChange={(e) => {
                      updateHeader(index, headerIndex, {
                        value: e.target.value
                      });
                    }}
                    placeholder="Header值"
                  />
                  <Button
                    isIconOnly
                    color="danger"
                    onPress={() => removeHeader(index, headerIndex)}
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
                  let newKey = "Content-Type";
                  let count = 1;
                  
                  const commonHeaders = [
                    "Authorization", 
                    "Accept", 
                    "X-API-Key", 
                    "User-Agent", 
                  ];
                  
                  const existingKeys = tool.headersOrder || Object.keys(tool.headers || {});
                  
                  for (const header of commonHeaders) {
                    if (!existingKeys.includes(header)) {
                      newKey = header;
                      break;
                    }
                  }
                  
                  if (existingKeys.includes(newKey)) {
                    while (existingKeys.includes(`X-Header-${count}`)) {
                      count++;
                    }
                    newKey = `X-Header-${count}`;
                  }
                  
                  addHeader(index, newKey);
                }}
              >
                添加Header
              </Button>
            </div>
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
              }}
              placeholder="例如: {{.Response.Body}}"
            ></textarea>
          </div>
        </div>
      ))}
      {/* 添加工具按钮 */}
      <Button
        color="primary"
        className="mt-2 w-full"
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
} 