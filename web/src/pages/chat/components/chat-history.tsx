import React from 'react';
import { Card, CardBody, Button } from "@heroui/react";
import { Icon } from '@iconify/react';

interface ChatHistoryProps {
  selectedChat: string | null;
  onSelectChat: (id: string) => void;
}

export function ChatHistory({ selectedChat, onSelectChat }: ChatHistoryProps) {
  const chats = [
    { id: '1', title: 'MCP Discussion 1', date: '2024-03-10' },
    { id: '2', title: 'Gateway Config Help', date: '2024-03-09' },
    { id: '3', title: 'Deployment Issues', date: '2024-03-08' },
  ];

  return (
    <Card className="w-64 relative group">
      <div 
        className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/20 transition-colors"
        onMouseDown={(e) => {
          const startX = e.pageX;
          const startWidth = e.currentTarget.parentElement?.offsetWidth || 0;
          
          const handleMouseMove = (e: MouseEvent) => {
            const delta = e.pageX - startX;
            const newWidth = Math.max(200, Math.min(400, startWidth + delta));
            if (e.currentTarget?.parentElement) {
              e.currentTarget.parentElement.style.width = `${newWidth}px`;
            }
          };
          
          const handleMouseUp = () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
          };
          
          document.addEventListener('mousemove', handleMouseMove);
          document.addEventListener('mouseup', handleMouseUp);
        }}
      />
      <CardBody className="p-0">
        <div className="p-4">
          <Button
            color="primary"
            className="w-full"
            startContent={<Icon icon="lucide:plus" />}
          >
            New Chat
          </Button>
        </div>
        <div className="space-y-1">
          {chats.map((chat) => (
            <Button
              key={chat.id}
              variant="light"
              className={`w-full justify-start px-4 py-2 ${
                selectedChat === chat.id ? 'bg-primary-100' : ''
              }`}
              onPress={() => onSelectChat(chat.id)}
            >
              <div className="flex flex-col items-start">
                <span className="text-sm">{chat.title}</span>
                <span className="text-xs text-default-500">{chat.date}</span>
              </div>
            </Button>
          ))}
        </div>
      </CardBody>
    </Card>
  );
}