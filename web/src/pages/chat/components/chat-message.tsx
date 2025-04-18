import { Avatar } from "@heroui/react";

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: Date;
  isStreaming?: boolean;
}

interface ChatMessageProps {
  message: Message;
}

export function ChatMessage({ message }: ChatMessageProps) {
  const isBot = message.sender === 'bot';

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
      </div>
    </div>
  );
}
