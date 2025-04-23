import { Avatar, Button, Accordion, AccordionItem } from "@heroui/react";
import { Icon } from "@iconify/react";
import { useContext } from "react";
import { toast } from "react-hot-toast";
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from "rehype-highlight";
import rehypeKatex from 'rehype-katex';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import 'katex/dist/katex.min.css';
import 'highlight.js/styles/github.css';

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
        <div className="prose prose-sm dark:prose-invert max-w-none">
          <ReactMarkdown
            remarkPlugins={[remarkGfm, remarkMath]}
            rehypePlugins={[rehypeHighlight, rehypeKatex]}
            components={{
              code({className, children, ...props}) {
                const match = /language-(\w+)/.exec(className || '');
                return match ? (
                  <code className={className} {...props}>
                    {children}
                  </code>
                ) : (
                  <code className="bg-gray-100 dark:bg-gray-800 rounded px-1" {...props}>
                    {children}
                  </code>
                );
              },
              p: ({children, ...props}) => <p className="my-2" {...props}>{children}</p>,
              h1: ({children, ...props}) => <h1 className="text-2xl font-bold my-4" {...props}>{children}</h1>,
              h2: ({children, ...props}) => <h2 className="text-xl font-bold my-3" {...props}>{children}</h2>,
              h3: ({children, ...props}) => <h3 className="text-lg font-bold my-2" {...props}>{children}</h3>,
              ul: ({children, ...props}) => <ul className="list-disc pl-6 my-2" {...props}>{children}</ul>,
              ol: ({children, ...props}) => <ol className="list-decimal pl-6 my-2" {...props}>{children}</ol>,
              li: ({children, ...props}) => <li className="my-1" {...props}>{children}</li>,
              blockquote: ({children, ...props}) => <blockquote className="border-l-4 border-gray-300 dark:border-gray-600 pl-4 my-2 italic" {...props}>{children}</blockquote>,
              a: ({children, ...props}) => <a className="text-blue-600 dark:text-blue-400 hover:underline" {...props}>{children}</a>,
            }}
          >
            {message.content}
          </ReactMarkdown>
        </div>
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
                        } catch {
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
