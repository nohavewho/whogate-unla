import { v4 as uuidv4 } from 'uuid';

export interface WebSocketMessage {
  type: 'message' | 'system';
  content: string;
  sender: string;
  timestamp: number;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private messageHandlers: ((message: WebSocketMessage) => void)[] = [];
  private sessionId: string = '';
  private welcomeMessage: WebSocketMessage = {
    type: 'system',
    content: '你好，欢迎使用MCP Gateway！',
    sender: 'bot',
    timestamp: Date.now(),
  };

  constructor() {
    this.newChat();
  }

  clearMessageHandlers() {
    this.messageHandlers = [];
  }

  newChat() {
    // Clear existing connection if any
    this.disconnect();
    // Clear message handlers
    this.clearMessageHandlers();
    // Generate new session ID
    this.sessionId = uuidv4();
    // Show welcome message
    this.messageHandlers.forEach(handler => handler(this.welcomeMessage));
  }

  switchChat(sessionId: string) {
    // Clear existing connection
    this.disconnect();
    // Clear message handlers
    this.clearMessageHandlers();
    // Set new session ID
    this.sessionId = sessionId;
    // Connect to new session
    return this.connect();
  }

  connect() {
    if (this.ws) {
      return Promise.resolve();
    }

    return new Promise<void>((resolve) => {
      this.ws = new WebSocket(`ws://${window.location.host}/ws/chat?sessionId=${this.sessionId}`);

      this.ws.onopen = () => {
        resolve();
      };

      this.ws.onmessage = (event) => {
        const message = JSON.parse(event.data) as WebSocketMessage;
        this.messageHandlers.forEach(handler => handler(message));
      };

      this.ws.onclose = () => {
        this.ws = null;
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
    if (!this.ws) {
      await this.connect();
    }

    const message: WebSocketMessage = {
      type: 'message',
      content,
      sender: 'user123', // This should be replaced with actual user ID
      timestamp: Date.now(),
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

  getSessionId(): string {
    return this.sessionId;
  }

  getWelcomeMessage(): WebSocketMessage {
    return this.welcomeMessage;
  }
}

// Create a singleton instance
export const wsService = new WebSocketService(); 