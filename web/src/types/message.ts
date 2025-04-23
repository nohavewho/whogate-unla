export interface Message {
  id: string;
  session_id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: string;
  isStreaming?: boolean;
  toolCalls?: ToolCall[];
  toolResult?: ToolResult;
}

export interface ToolCall {
  id: string;
  type: string;
  function: {
    name: string;
    arguments: string;
  };
}

export interface ToolResult {
  toolCallId: string;
  name: string;
  result: string;
}
