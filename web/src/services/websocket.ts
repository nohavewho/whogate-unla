import toast from 'react-hot-toast';
import { v4 as uuidv4 } from 'uuid';

export interface WebSocketMessage {
  type: 'system' | 'message' | 'stream';
  content: string;
  sender: 'user' | 'bot';
  timestamp: number;
  id: string;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private messageHandlers: ((message: WebSocketMessage) => void)[] = [];
  private streamHandlers: ((chunk: string) => void)[] = [];
  private sessionId: string = '';


  constructor() {
    this.cleanup();
  }

  clearMessageHandlers() {
    this.messageHandlers = [];
    this.streamHandlers = [];
  }

  cleanup() {
    // Clear existing connection if any
    this.disconnect();
    // Clear message handlers
    this.clearMessageHandlers();
    // Generate new session ID
    this.sessionId = uuidv4();
  }

  switchChat(sessionId: string) {
    // Clear existing connection
    this.disconnect();
    // Don't clear message handlers here
    // this.clearMessageHandlers();
    // Set new session ID
    this.sessionId = sessionId;
    // Don't connect immediately, wait for first message
    return Promise.resolve();
  }

  connect() {
    if (this.ws) {
      return Promise.resolve();
    }

    return new Promise<void>((resolve) => {
      this.ws = new WebSocket(`/ws/chat?sessionId=${this.sessionId}`);

      this.ws.onopen = () => {
        resolve();
      };

      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data) as WebSocketMessage;
        if (message.type === 'stream') {
          this.streamHandlers.forEach(handler => handler(message.content));
        } else {
          this.messageHandlers.forEach(handler => handler(message));
        }
      };

      this.ws.onclose = () => {
        toast.error('WebSocket 连接已断开', {
          duration: 3000,
          position: 'bottom-right',
        });
        this.ws = null;
      };

      this.ws.onerror = (error) => {
        toast.error('WebSocket 发生错误' + error, {
          duration: 3000,
          position: 'bottom-right',
        });
      };
    });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  async sendMessage(content: string) {
    // Connect to WebSocket if not already connected
    if (!this.ws) {
      await this.connect();
    }

    const message: WebSocketMessage = {
      type: 'message',
      content,
      sender: 'user',
      timestamp: Date.now(),
      id: uuidv4(),
    };

    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  onMessage(handler: (message: WebSocketMessage) => void) {
    this.messageHandlers.push(handler);
    return () => {
      this.messageHandlers = this.messageHandlers.filter(h => h !== handler);
    };
  }

  onStream(handler: (chunk: string) => void) {
    this.streamHandlers.push(handler);
    return () => {
      this.streamHandlers = this.streamHandlers.filter(h => h !== handler);
    };
  }

  getSessionId(): string {
    return this.sessionId;
  }
}

// Create a singleton instance
export const wsService = new WebSocketService();
