import { Avatar, Button, Accordion, AccordionItem } from "@heroui/react";
import { Icon } from "@iconify/react";
import { toast } from "react-hot-toast";
import { useContext } from "react";

import { mcpService } from "../../../services/mcp";
import { wsService } from "../../../services/websocket";
import {Message, ToolCall, ToolResult} from "../../../types/message";
import { ChatContext } from "../chat-context";

interface ChatMessageProps {
  message: Message;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isBot = message.sender === 'bot';
  const { messages } = useContext(ChatContext);

  const findToolResult = (toolId: string): ToolResult | undefined => {
    return messages.find((m: Message) => m.toolResult?.toolCallId === toolId)?.toolResult;
  };

  const handleRunTool = async (tool: ToolCall) => {
    try {
      if (!tool?.function?.name) {
        toast.error('工具名称格式错误', {
          duration: 3000,
          position: 'bottom-right',
        });
        return;
      }

      // 解析 serverName:toolName 格式
      const [serverName, toolName] = tool.function.name.split(':');
      if (!serverName || !toolName) {
        toast.error('工具名称格式错误', {
          duration: 3000,
          position: 'bottom-right',
        });
        return;
      }

      const sessionId = mcpService.getSessionId(serverName);

      if (!sessionId) {
        toast.error(`服务器 ${serverName} 未连接`, {
          duration: 3000,
          position: 'bottom-right',
        });
        return;
      }

      // 解析 arguments 字符串为对象
      const args = JSON.parse(tool.function.arguments);
      const result = await mcpService.callTool(serverName, toolName, args);

      // 显示工具调用结果
      toast.success(`工具调用成功: ${result}`, {
        duration: 3000,
        position: 'bottom-right',
      });

      // 将工具调用结果作为新消息发送
      await wsService.sendToolResult(tool.function.name, tool.id, result);
    } catch (error) {
      console.error('工具调用失败:', error);
      toast.error(`工具调用失败: ${(error as Error).message}`, {
        duration: 3000,
        position: 'bottom-right',
      });
    }
  };

  return (
    <div className={`flex gap-3 mb-4 ${isBot ? 'flex-row' : 'flex-row-reverse'}`}>
      <Avatar
        size="sm"
        src={isBot ? "https://img.heroui.chat/image/avatar?w=32&h=32&u=1" : undefined}
        name={isBot ? "MCP" : "You"}
      />
      <div
        className={`px-4 py-2 rounded-lg max-w-[80%] ${
          isBot ? 'bg-content2' : 'bg-primary text-primary-foreground'
        }`}
      >
        {message.content}
        {message.isStreaming && (
          <span className="inline-block w-2 h-4 ml-1 bg-current animate-pulse" />
        )}
        {message.toolCalls?.map((tool, index) => {
          const toolResult = findToolResult(tool.id);
          return tool?.function?.name ? (
            <div key={index} className="mt-2 p-2 border rounded bg-content1">
              <div className="font-medium mb-2">{tool.function.name}</div>
              <Accordion selectionMode="multiple">
                <AccordionItem
                  key={`${tool.id}-args`}
                  title="Arguments"
                  className="px-0"
                >
                  <pre className="text-sm p-2 bg-content2 rounded overflow-auto">
                    {JSON.stringify(JSON.parse(tool.function.arguments), null, 2)}
                  </pre>
                </AccordionItem>
                {toolResult ? (
                  <AccordionItem
                    key={`${tool.id}-result`}
                    title="Result"
                    className="px-0"
                  >
                    <pre className="text-sm p-2 bg-content2 rounded overflow-auto">
                      {(() => {
                        try {
                          return JSON.stringify(JSON.parse(toolResult.result), null, 2);
                        } catch (e) {
                          return toolResult.result;
                        }
                      })()}
                    </pre>
                  </AccordionItem>
                ) : null}
              </Accordion>
              {!toolResult && (
                <Button
                  size="sm"
                  color="primary"
                  className="mt-2"
                  startContent={<Icon icon="lucide:play" />}
                  onPress={() => handleRunTool(tool)}
                >
                  运行工具
                </Button>
              )}
            </div>
          ) : null;
        })}
      </div>
    </div>
  );
}
