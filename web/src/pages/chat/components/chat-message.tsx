import { Avatar, Button } from "@heroui/react";
import { Icon } from "@iconify/react";
import { toast } from "react-hot-toast";

import { mcpService } from "../../../services/mcp";

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: Date;
  isStreaming?: boolean;
  tool_calls?: Array<{
    name: string;
    arguments: Record<string, unknown>;
  }>;
}

interface ChatMessageProps {
  message: Message;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isBot = message.sender === 'bot';

  const handleRunTool = async (name: string, args: Record<string, unknown>) => {
    try {
      // 解析 serverName:toolName 格式
      const [serverName, toolName] = name.split(':');
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

      const result = await mcpService.callTool(serverName, toolName, args);
      
      // 显示工具调用结果
      toast.success(`工具调用成功: ${result}`, {
        duration: 3000,
        position: 'bottom-right',
      });
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
        {message.tool_calls && message.tool_calls.map((tool, index) => (
          <div key={index} className="mt-2 p-2 border rounded bg-content1">
            <div className="font-medium">{tool.name}</div>
            <pre className="text-sm mt-1 p-2 bg-content2 rounded overflow-auto">
              {JSON.stringify(tool.arguments, null, 2)}
            </pre>
            <Button
              size="sm"
              color="primary"
              className="mt-2"
              startContent={<Icon icon="lucide:play" />}
              onPress={() => handleRunTool(tool.name, tool.arguments)}
            >
              运行工具
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
