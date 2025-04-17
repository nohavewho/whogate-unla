import React from 'react';
import { Card, CardBody, Input, Button, Avatar, Divider } from "@heroui/react";
import { Icon } from '@iconify/react';
import { ChatHistory } from './components/chat-history';
import { ChatMessage } from './components/chat-message';
import { Select, SelectItem } from "@heroui/react";

interface Message {
  id: string;
  content: string;
  sender: 'user' | 'bot';
  timestamp: Date;
}

export function ChatInterface() {
  const [messages, setMessages] = React.useState<Message[]>([]);
  const [input, setInput] = React.useState('');
  const [selectedChat, setSelectedChat] = React.useState<string | null>(null);
  const [activeServices, setActiveServices] = React.useState<string[]>([]);

  const availableServices = [
    { id: "user-svc", name: "User Service" },
    { id: "auth-svc", name: "Auth Service" },
    { id: "payment-svc", name: "Payment Service" },
  ];

  const handleSend = () => {
    if (!input.trim()) return;
    
    const newMessage: Message = {
      id: Date.now().toString(),
      content: input,
      sender: 'user',
      timestamp: new Date(),
    };
    
    setMessages([...messages, newMessage]);
    setInput('');
    
    // Simulate bot response
    setTimeout(() => {
      const botResponse: Message = {
        id: (Date.now() + 1).toString(),
        content: 'This is a sample MCP response.',
        sender: 'bot',
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, botResponse]);
    }, 1000);
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
            <div className="flex-1 overflow-auto p-4">
              {messages.map((message) => (
                <ChatMessage key={message.id} message={message} />
              ))}
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
                  <SelectItem key={service.id} value={service.id}>
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