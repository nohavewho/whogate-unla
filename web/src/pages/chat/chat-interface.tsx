import React from 'react';
import { Card, CardBody, Input, Button, Avatar, Divider } from "@heroui/react";
import { Icon } from '@iconify/react';
import { ChatHistory } from './components/chat-history';
import { ChatMessage } from './components/chat-message';
import { Select, SelectItem } from "@heroui/react";
import { useParams, useNavigate } from 'react-router-dom';
import { wsService, WebSocketMessage } from '../../services/websocket';

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

  const availableServices = [
    { id: "user-svc", name: "User Service" },
    { id: "auth-svc", name: "Auth Service" },
    { id: "payment-svc", name: "Payment Service" },
  ];

  React.useEffect(() => {
    if (!sessionId) {
      // If no session ID in URL, create a new one and redirect
      wsService.newChat();
      const newSessionId = wsService.getSessionId();
      navigate(`/chat/${newSessionId}`);
    } else {
      // Show welcome message
      const welcomeMessage = wsService.getWelcomeMessage();
      const newMessage: Message = {
        id: Date.now().toString(),
        content: welcomeMessage.content,
        sender: 'bot',
        timestamp: new Date(welcomeMessage.timestamp),
      };
      setMessages([newMessage]);

      // Set up message handler
      const unsubscribe = wsService.onMessage((message: WebSocketMessage) => {
        const newMessage: Message = {
          id: Date.now().toString(),
          content: message.content,
          sender: message.sender as 'user' | 'bot',
          timestamp: new Date(message.timestamp),
        };
        setMessages(prev => [...prev, newMessage]);
      });

      // Cleanup on unmount
      return () => {
        unsubscribe();
        wsService.disconnect();
      };
    }
  }, [sessionId, navigate]);

  // Add scroll position check
  React.useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const { scrollTop, scrollHeight, clientHeight } = container;
      const isNearBottom = scrollHeight - scrollTop - clientHeight < 100;
      setIsNearBottom(isNearBottom);
    };

    container.addEventListener('scroll', handleScroll);
    return () => container.removeEventListener('scroll', handleScroll);
  }, []);

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

    setMessages([...messages, newMessage]);
    wsService.sendMessage(input);
    setInput('');
  };

  return (
    <div className="flex h-[calc(100vh-8rem)]">
      <ChatHistory
        selectedChat={selectedChat}
        onSelectChat={setSelectedChat}
      />

      <div className="flex-1 ml-4">
        <Card className="h-full">
          <CardBody className="p-0 h-full flex flex-col">
            <div 
              ref={messagesContainerRef}
              className="flex-1 overflow-auto p-4"
            >
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
