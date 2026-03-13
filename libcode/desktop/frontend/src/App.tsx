import React, { useState, useEffect } from 'react';
import './App.css';
import { Chat } from './components/Chat';
import { Sidebar } from './components/Sidebar';
import { Settings } from './components/Settings';
import { Message, Session } from './types';

function App() {
  const [currentView, setCurrentView] = useState<'chat' | 'settings'>('chat');
  const [sessions, setSessions] = useState<Session[]>([]);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [config, setConfig] = useState<Record<string, any>>({});

  useEffect(() => {
    loadSessions();
    loadConfig();
  }, []);

  const loadSessions = async () => {
    try {
      const result = await (window as any).GetSessions();
      setSessions(result || []);
    } catch (error) {
      console.error('Failed to load sessions:', error);
    }
  };

  const loadConfig = async () => {
    try {
      const result = await (window as any).GetConfig();
      setConfig(result || {});
    } catch (error) {
      console.error('Failed to load config:', error);
    }
  };

  const handleSendMessage = async (content: string) => {
    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date().toISOString(),
    };

    setMessages((prev) => [...prev, userMessage]);

    try {
      const response = await (window as any).SendMessage(content);
      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: response,
        timestamp: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, assistantMessage]);
    } catch (error) {
      console.error('Failed to send message:', error);
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `Error: ${error}`,
        timestamp: new Date().toISOString(),
      };
      setMessages((prev) => [...prev, errorMessage]);
    }
  };

  const handleSelectSession = async (sessionId: string) => {
    try {
      const session = await (window as any).LoadSession(sessionId);
      setCurrentSession(session);
      setMessages(session.messages || []);
      setCurrentView('chat');
    } catch (error) {
      console.error('Failed to load session:', error);
    }
  };

  const handleNewChat = () => {
    setCurrentSession(null);
    setMessages([]);
    setCurrentView('chat');
  };

  const handleSaveConfig = async (newConfig: Record<string, any>) => {
    try {
      await (window as any).SaveConfig(newConfig);
      setConfig(newConfig);
      setCurrentView('chat');
    } catch (error) {
      console.error('Failed to save config:', error);
    }
  };

  return (
    <div className="app">
      <Sidebar
        sessions={sessions}
        currentSession={currentSession}
        onSelectSession={handleSelectSession}
        onNewChat={handleNewChat}
        onOpenSettings={() => setCurrentView('settings')}
      />
      <main className="main-content">
        {currentView === 'chat' ? (
          <Chat messages={messages} onSendMessage={handleSendMessage} />
        ) : (
          <Settings config={config} onSave={handleSaveConfig} onClose={() => setCurrentView('chat')} />
        )}
      </main>
    </div>
  );
}

export default App;
