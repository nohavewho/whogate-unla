import { Avatar, Button } from "@heroui/react";
import { Icon } from "@iconify/react";

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

  const handleRunTool = (toolName: string, args: Record<string, unknown>) => {
    // TODO: Implement tool execution
    console.log('Running tool:', toolName, args);
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
              Run Tool
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
