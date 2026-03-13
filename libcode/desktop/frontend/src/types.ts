export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
}

export interface Session {
  id: string;
  title: string;
  created_at: string;
  updated_at: string;
  messages?: Message[];
}

export interface Config {
  apiKey?: string;
  model?: string;
  theme?: string;
  logLevel?: string;
  storagePath?: string;
  maxTokens?: number;
  temperature?: number;
  autoSave?: boolean;
}

export interface Tool {
  id: string;
  name: string;
  description: string;
  schema: any;
}

export interface ToolResult {
  title?: string;
  output: string;
  metadata?: Record<string, any>;
  error?: string;
}
