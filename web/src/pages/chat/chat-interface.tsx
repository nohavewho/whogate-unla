import React from 'react';
import { Card, CardBody, Input, Button, Divider } from "@heroui/react";
import { Icon } from '@iconify/react';
import { ChatHistory } from './components/chat-history';
import { ChatMessage } from './components/chat-message';
import { Select, SelectItem } from "@heroui/react";
import { useParams, useNavigate } from 'react-router-dom';
import { wsService, WebSocketMessage } from '../../services/websocket';
import { getChatMessages } from '../../services/api';

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: Date;
}

export function ChatInterface() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [messages, setMessages] = React.useState<Message[]>([]);
  const [input, setInput] = React.useState('');
  const [selectedChat, setSelectedChat] = React.useState<string | null>(null);
  const [activeServices, setActiveServices] = React.useState<string[]>([]);
  const messagesEndRef = React.useRef<HTMLDivElement>(null);
  const messagesContainerRef = React.useRef<HTMLDivElement>(null);
  const [isNearBottom, setIsNearBottom] = React.useState(true);
  const [page, setPage] = React.useState(1);
  const [hasMore, setHasMore] = React.useState(true);
  const [loading, setLoading] = React.useState(false);

  const availableServices = [
    { id: "user-svc", name: "User Service" },
    { id: "auth-svc", name: "Auth Service" },
    { id: "payment-svc", name: "Payment Service" },
  ];

  const loadMessages = async (sessionId: string, pageNum: number = 1) => {
    if (loading || !hasMore) return;
    
    setLoading(true);
    try {
      const data = await getChatMessages(sessionId, pageNum);
      
      // Check if data exists
      if (!data) {
        console.error('Invalid response format:', data);
        return;
      }

      // Convert backend message format to frontend format
      const newMessages = data.map((msg: any) => ({
        id: msg.id,
        content: msg.content,
        sender: msg.sender as 'user' | 'bot',
        timestamp: new Date(msg.timestamp),
      }));

      // Sort messages by timestamp in ascending order
      newMessages.sort((a: Message, b: Message) => a.timestamp.getTime() - b.timestamp.getTime());

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
  };

  React.useEffect(() => {
    if (!sessionId) {
      // If no session ID in URL, create a new one and redirect
      wsService.newChat();
      const newSessionId = wsService.getSessionId();
      navigate(`/chat/${newSessionId}`);
    } else {
      // Clear old messages first
      setMessages([]);
      setPage(1);
      setHasMore(true);
      setSelectedChat(sessionId);

      // Switch WebSocket connection to new session
      wsService.switchChat(sessionId).then(() => {
        // Set up message handler
        const unsubscribe = wsService.onMessage((message: WebSocketMessage & { id?: string }) => {
          setMessages(prev => {
            // Check if message already exists to prevent duplicates
            if (prev.some(m => m.id === message.id)) {
              return prev;
            }
            const newMessage: Message = {
              id: message.id || Date.now().toString(),
              content: message.content,
              sender: message.sender as 'user' | 'bot',
              timestamp: new Date(message.timestamp),
            };
            return [...prev, newMessage];
          });
        });

        // Cleanup on unmount or session change
        return () => {
          unsubscribe();
          wsService.disconnect();
          wsService.clearMessageHandlers();
        };
      });
    }
  }, [sessionId, navigate]);

  // Add scroll position check and load more messages when scrolling up
  React.useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 100;
      setIsNearBottom(isNearBottom);

      // Load more messages when scrolling to top
      if (scrollTop < 100 && hasMore && !loading && sessionId) {
        loadMessages(sessionId, page + 1);
      }
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, [sessionId, page, hasMore, loading]);

  // Modify auto-scroll effect to only scroll when appropriate
  React.useEffect(() => {
    if (isNearBottom) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isNearBottom]);

  const handleSend = () => {
    if (!input.trim()) return;

    const newMessage: Message = {
      id: Date.now().toString(),
      content: input,
      sender: 'user',
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, newMessage]);
    wsService.sendMessage(input);
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
              className="flex-1 overflow-auto p-4"
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
                onSelectionChange={setActiveServices as any}
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
