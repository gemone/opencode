import React, { useState, useRef, useEffect } from 'react';
import { Message } from '../types';
import './Chat.css';

interface ChatProps {
  messages: Message[];
  onSendMessage: (content: string) => void;
}

export const Chat: React.FC<ChatProps> = ({ messages, onSendMessage }) => {
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (input.trim()) {
      onSendMessage(input.trim());
      setInput('');
    }
  };

  return (
    <div className="chat">
      <div className="chat-messages">
        {messages.length === 0 ? (
          <div className="chat-empty">
            <h2>Libcode</h2>
            <p>AI-powered development tool</p>
            <p>Send a message to get started</p>
          </div>
        ) : (
          messages.map((message) => (
            <div key={message.id} className={`chat-message chat-message-${message.role}`}>
              <div className="chat-message-content">
                {message.content}
              </div>
              <div className="chat-message-timestamp">
                {new Date(message.timestamp).toLocaleTimeString()}
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
      <form className="chat-input-form" onSubmit={handleSubmit}>
        <input
          type="text"
          className="chat-input"
          placeholder="Send a message..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          autoFocus
        />
        <button type="submit" className="chat-send-button" disabled={!input.trim()}>
          Send
        </button>
      </form>
    </div>
  );
};
