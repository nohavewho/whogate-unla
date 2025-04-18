import { Card, CardBody, Button, Input, Select, SelectItem, Divider } from '@heroui/react';
import type { Selection } from '@heroui/react';
import { Icon } from '@iconify/react';
import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { v4 as uuidv4 } from 'uuid';

import { getChatMessages } from '../../services/api';
import { wsService, WebSocketMessage } from '../../services/websocket';

import { ChatHistory } from './components/chat-history';
import { ChatMessage } from './components/chat-message';

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: Date;
  isStreaming?: boolean;
}

interface BackendMessage {
  id: string;
  content: string;
  sender: string;
  timestamp: string;
}

export function ChatInterface() {
  const navigate = useNavigate();
  const { sessionId } = useParams();
  const messagesEndRef = React.useRef<HTMLDivElement>(null);
  const messagesContainerRef = React.useRef<HTMLDivElement>(null);
  const [messages, setMessages] = React.useState<Message[]>([]);
  const [input, setInput] = React.useState('');
  const [selectedChat, setSelectedChat] = React.useState<string | null>(null);
  const [activeServices, setActiveServices] = React.useState<string[]>([]);
  const [isNearBottom, setIsNearBottom] = React.useState(true);
  const [page, setPage] = React.useState(1);
  const [hasMore, setHasMore] = React.useState(true);
  const [loading, setLoading] = React.useState(false);
  const [lastScrollTop, setLastScrollTop] = React.useState(0);
  const loadedSessionRef = React.useRef<string | null>(null);

  const availableServices = [
    { id: "user-svc", name: "User Service" },
    { id: "auth-svc", name: "Auth Service" },
    { id: "payment-svc", name: "Payment Service" },
  ];

  const loadMessages = React.useCallback(async (sessionId: string, pageNum: number = 1) => {
    if (loading || !hasMore) return;

    setLoading(true);
    try {
      const data = await getChatMessages(sessionId, pageNum);

      // Check if data exists
      if (!data) {
        const welcomeMessage: Message = {
          id: uuidv4(),
          content: '你好，欢迎使用MCP Gateway！',
          sender: 'bot',
          timestamp: new Date(),
        };
        setMessages([welcomeMessage]);
        return;
      }

      // Convert backend message format to frontend format
      const newMessages = data.map((msg: BackendMessage) => ({
        id: msg.id,
        content: msg.content,
        sender: msg.sender as 'user' | 'bot',
        timestamp: new Date(msg.timestamp),
      }));

      if (pageNum === 1) {
        setMessages(newMessages);
      } else {
        setMessages(prev => [...newMessages, ...prev]);
      }

      // Since we don't have hasMore in the response, we'll assume there are more messages
      // if we got a full page of messages
      setHasMore(newMessages.length === 20);
      setPage(pageNum);
    } catch (error) {
      console.error('Error loading messages:', error);
      setMessages([]);
      setHasMore(false);
    } finally {
      setLoading(false);
    }
  }, [loading, hasMore]);

  React.useEffect(() => {
    if (!sessionId) {
      // If no session ID in URL, create a new one and redirect
      wsService.cleanup();
      const newSessionId = wsService.getSessionId();
      navigate(`/chat/${newSessionId}`);
      return
    }

    // Skip if we've already loaded messages for this session
    if (loadedSessionRef.current === sessionId) {
      return;
    }

    // Clear old messages first
    setMessages([]);
    setPage(1);
    setHasMore(true);
    setSelectedChat(sessionId);
    loadedSessionRef.current = sessionId;

    // Set up message handler for regular messages
    const unsubscribe = wsService.onMessage((message: WebSocketMessage) => {
      // Skip streaming messages as they are handled by stream handler
      if (message.type === 'stream') return;

      setMessages(prev => {
        // Check if message already exists to prevent duplicates
        if (prev.some(m => m.id === message.id)) {
          return prev;
        }
        const newMessage: Message = {
          id: message.id,
          content: message.content,
          sender: message.sender,
          timestamp: new Date(message.timestamp),
        };
        return [...prev, newMessage];
      });
    });

    // Set up stream handler
    const unsubscribeStream = wsService.onStream((chunk: string) => {
      setMessages(prev => {
        const lastMessage = prev[prev.length - 1];
        // If the last message is from bot and is streaming, append to it
        if (lastMessage && lastMessage.sender === 'bot' && lastMessage.isStreaming) {
          const updatedMessages = [...prev];
          updatedMessages[prev.length - 1] = {
            ...lastMessage,
            content: lastMessage.content + chunk,
            isStreaming: true
          };
          return updatedMessages;
        }
        // If no bot message exists or last message is from user, create new one
        return [...prev, {
          id: uuidv4(),
          content: chunk,
          sender: 'bot',
          timestamp: new Date(),
          isStreaming: true
        }];
      });
    });

    // Switch to new session and load history
    wsService.switchChat(sessionId).then(async () => {
      // Load chat history
      await loadMessages(sessionId);
    });

    // Cleanup on unmount or session change
    return () => {
      unsubscribe();
      unsubscribeStream();
    };
  }, [sessionId, navigate, loadMessages]);

  // Add scroll position check and load more messages when scrolling up
  React.useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 100;
      setIsNearBottom(isNearBottom);

      // Only load more messages when user actively scrolls up
      if (scrollTop < lastScrollTop && scrollTop < 100 && hasMore && !loading && sessionId) {
        loadMessages(sessionId, page + 1);
      }
      setLastScrollTop(scrollTop);
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [sessionId, page, hasMore, loading, lastScrollTop, loadMessages]);

  // Modify auto-scroll effect to only scroll when appropriate
  React.useEffect(() => {
    if (isNearBottom) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isNearBottom]);

  const handleSend = async () => {
    if (!input.trim()) return;

    const newMessage: Message = {
      id: uuidv4(),
      content: input,
      sender: 'user',
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, newMessage]);
    await wsService.sendMessage(input);
    setInput('');
  };

  return (
    <div className="flex h-[calc(100vh-8rem)]">
      <ChatHistory
        selectedChat={selectedChat}
        onSelectChat={(id) => {
          setSelectedChat(id);
          navigate(`/chat/${id}`);
        }}
      />

      <div className="flex-1 ml-4">
        <Card className="h-full">
          <CardBody className="p-0 h-full flex flex-col">
            <div
              ref={messagesContainerRef}
              className="flex-1 overflow-auto p-4 scrollbar-hide"
            >
              {loading && page > 1 && (
                <div className="text-center text-default-500 py-2">
                  Loading more messages...
                </div>
              )}
              {messages.map((message) => (
                <ChatMessage key={message.id} message={message} />
              ))}
              <div ref={messagesEndRef} />
            </div>
            <Divider />
            <div className="p-4 flex flex-col gap-4">
              <Select
                label="Active Services"
                selectionMode="multiple"
                placeholder="Select active services"
                selectedKeys={activeServices}
                onSelectionChange={(keys: Selection) => {
                  setActiveServices(Array.from(keys) as string[]);
                }}
                className="max-w-xs"
              >
                {availableServices.map((service) => (
                  <SelectItem key={service.id}>
                    {service.name}
                  </SelectItem>
                ))}
              </Select>

              <Input
                value={input}
                onValueChange={setInput}
                placeholder="Type your message..."
                onKeyPress={(e) => e.key === 'Enter' && handleSend()}
                endContent={
                  <Button
                    isIconOnly
                    color="primary"
                    variant="light"
                    onPress={handleSend}
                  >
                    <Icon icon="lucide:send" className="text-lg" />
                  </Button>
                }
              />
            </div>
          </CardBody>
        </Card>
      </div>
    </div>
  );
}
