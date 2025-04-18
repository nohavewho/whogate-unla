import React, { useEffect, useState } from 'react';
import { Card, CardBody, Button } from "@heroui/react";
import { Icon } from '@iconify/react';
import { useNavigate } from 'react-router-dom';
import { wsService } from '../../../services/websocket';

interface ChatHistoryProps {
  selectedChat: string | null;
  onSelectChat: (id: string) => void;
}

interface Session {
  id: string;
  createdAt: string;
  title: string;
  lastMessage?: {
    id: string;
    sessionId: string;
    content: string;
    sender: string;
    timestamp: string;
  };
}

export function ChatHistory({ selectedChat, onSelectChat }: ChatHistoryProps) {
  const navigate = useNavigate();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchSessions();
  }, []);

  const fetchSessions = async () => {
    try {
      const response = await fetch('/api/chat/sessions');
      if (!response.ok) {
        throw new Error('Failed to fetch sessions');
      }
      const data = await response.json();
      setSessions(data);
    } catch (error) {
      console.error('Error fetching sessions:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleNewChat = () => {
    wsService.newChat();
    const newSessionId = wsService.getSessionId();
    navigate(`/chat/${newSessionId}`);
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  };

  const handleSessionSelect = (sessionId: string) => {
    onSelectChat(sessionId);
    navigate(`/chat/${sessionId}`);
  };

  return (
    <Card className="w-64 relative group">
      <div
        className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/20 transition-colors"
        onMouseDown={(e: React.MouseEvent<HTMLDivElement>) => {
          const startX = e.pageX;
          const startWidth = e.currentTarget.parentElement?.offsetWidth || 0;

          const handleMouseMove = (e: MouseEvent) => {
            const delta = e.pageX - startX;
            const newWidth = Math.max(200, Math.min(400, startWidth + delta));
            const parent = (e.target as HTMLElement).parentElement;
            if (parent) {
              parent.style.width = `${newWidth}px`;
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
            onPress={handleNewChat}
          >
            New Chat
          </Button>
        </div>
        <div className="space-y-1 px-4">
          {loading ? (
            <div className="p-4 text-center text-default-500">Loading...</div>
          ) : sessions.length === 0 ? (
            <div className="p-4 text-center text-default-500">No chat history</div>
          ) : (
            sessions.map((session) => (
              <Button
                key={session.id}
                variant="light"
                className={`w-full justify-start px-4 py-2 ${
                  selectedChat === session.id ? 'bg-primary-100' : ''
                }`}
                onPress={() => handleSessionSelect(session.id)}
              >
                <div className="flex flex-col items-start">
                  <span className="text-sm truncate w-full">
                    {session.lastMessage?.content || session.title || 'New Chat'}
                  </span>
                  <span className="text-xs text-default-500">
                    {formatDate(session.createdAt)}
                  </span>
                </div>
              </Button>
            ))
          )}
        </div>
      </CardBody>
    </Card>
  );
}
